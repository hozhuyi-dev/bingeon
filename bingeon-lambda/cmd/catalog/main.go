// Command catalog serves the read-only movie catalogue to anyone, behind
// CloudFront.
//
// For now it answers every path as an unknown action, to check CloudFront
// reaches it; the API phase replaces it with the router.
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handle(_ context.Context, _ events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return events.LambdaFunctionURLResponse{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "public, max-age=60, s-maxage=60",
		},
		Body: `{"error":{"code":"not_found","message":"No such catalogue action."}}`,
	}, nil
}

func main() {
	lambda.Start(handle)
}
