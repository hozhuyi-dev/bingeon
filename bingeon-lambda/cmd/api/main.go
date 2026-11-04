// Command api is the Lambda signed-in users invoke directly with
// identity-pool credentials.
//
// For now it only answers with the caller's Cognito identity, to check that
// Lambda passes it on; the API phase replaces it with the router.
package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type response struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}

func handle(ctx context.Context, _ json.RawMessage) (response, error) {
	lc, _ := lambdacontext.FromContext(ctx)
	if lc == nil || lc.Identity.CognitoIdentityID == "" {
		return response{Error: &errorBody{Code: "unauthenticated", Message: "Sign in first."}}, nil
	}
	return response{Data: map[string]string{"user": lc.Identity.CognitoIdentityID}}, nil
}

func main() {
	lambda.Start(handle)
}
