package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"bingeon-lambda/internal/platform/apperr"
)

// browserMaxAge is how long a browser keeps a cached reply: short, so a
// fix reaches it soon, while CloudFront keeps it for the action's ttl.
const browserMaxAge = 5 * time.Minute

// refusalMaxAge is how long an invalid or unknown request's reply is kept:
// briefly, as it won't change but shouldn't fill the cache.
const refusalMaxAge = time.Minute

// FunctionURL serves the router behind a Lambda function URL: GET
// <prefix>/<action>?<query> is the action, the query its payload (one
// value per name, as strings; handlers decode numbers with `json:",string"`).
func (r *Router) FunctionURL(prefix string) func(context.Context, events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return func(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		method := req.RequestContext.HTTP.Method
		if method != http.MethodGet && method != http.MethodHead {
			return reply(http.StatusMethodNotAllowed, "no-store", Failed(apperr.Invalid("Only GET is served."))), nil
		}
		action, ok := strings.CutPrefix(req.RawPath, prefix+"/")
		if !ok {
			return reply(http.StatusNotFound, cacheFor(refusalMaxAge, refusalMaxAge), Failed(apperr.NotFound("No such path."))), nil
		}
		payload, err := json.Marshal(req.QueryStringParameters)
		if err != nil {
			return events.LambdaFunctionURLResponse{}, err
		}
		res, ttl := r.handle(ctx, Request{Action: action, Payload: payload})
		status, cache := http.StatusOK, cacheFor(browserMaxAge, ttl)
		if res.Error != nil {
			status, cache = statusOf(res.Error.Code)
		} else if ttl == 0 {
			cache = "no-store"
		}
		out := reply(status, cache, res)
		if method == http.MethodHead {
			out.Body = ""
		}
		return out, nil
	}
}

func cacheFor(browser, shared time.Duration) string {
	return fmt.Sprintf("public, max-age=%d, s-maxage=%d", int(browser.Seconds()), int(shared.Seconds()))
}

// statusOf is an error's HTTP status, and how long it may be cached: what
// the request itself got wrong for a minute, a failure not at all.
func statusOf(code apperr.Code) (int, string) {
	switch code {
	case apperr.CodeInvalid:
		return http.StatusBadRequest, cacheFor(refusalMaxAge, refusalMaxAge)
	case apperr.CodeNotFound:
		return http.StatusNotFound, cacheFor(refusalMaxAge, refusalMaxAge)
	case apperr.CodeBusy:
		return http.StatusTooManyRequests, "no-store"
	case apperr.CodeUnavailable:
		return http.StatusBadGateway, "no-store"
	}
	return http.StatusInternalServerError, "no-store"
}

func reply(status int, cache string, res Response) events.LambdaFunctionURLResponse {
	body, err := json.Marshal(res)
	if err != nil {
		body = []byte(`{"error":{"code":"internal","message":"Something went wrong."}}`)
		status, cache = http.StatusInternalServerError, "no-store"
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Cache-Control": cache,
		},
		Body: string(body),
	}
}
