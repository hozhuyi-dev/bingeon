// Command catalog serves the movie catalogue to anyone, behind CloudFront:
// GET /catalog/<action>?<query> (see router.FunctionURL). It reads the
// TMDB token once, at a cold start, and nothing else.
package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"bingeon-lambda/internal/app"
	"bingeon-lambda/internal/platform/ssm"
	"bingeon-lambda/internal/platform/tmdb"
)

func main() {
	token, err := ssm.Secret(context.Background(), "TMDB_TOKEN_PARAM")
	if err != nil {
		log.Fatal(err)
	}
	r := app.Catalog(tmdb.New(token))
	r.LogRequests = true
	lambda.Start(r.FunctionURL("/catalog"))
}
