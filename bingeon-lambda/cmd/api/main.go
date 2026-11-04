// Command api is the Lambda signed-in users invoke directly with
// identity-pool credentials: {action, payload} in, {data} or {error} out
// (see platform/router). The user is the invocation's Cognito identity,
// never an id in the payload.
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"bingeon-lambda/internal/app"
	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/picks"
	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/dynamo"
	"bingeon-lambda/internal/platform/gemini"
	"bingeon-lambda/internal/platform/router"
	"bingeon-lambda/internal/platform/ssm"
	"bingeon-lambda/internal/platform/tmdb"
	"bingeon-lambda/internal/platform/user"
)

func main() {
	ctx := context.Background()
	db, err := dynamo.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	token, err := ssm.Secret(ctx, "TMDB_TOKEN_PARAM")
	if err != nil {
		log.Fatal(err)
	}
	key, err := ssm.Secret(ctx, "GEMINI_KEY_PARAM")
	if err != nil {
		log.Fatal(err)
	}
	picker, err := gemini.New(ctx, key)
	if err != nil {
		log.Fatal(err)
	}
	limit, err := strconv.Atoi(os.Getenv("GEMINI_DAILY_LIMIT"))
	if err != nil || limit <= 0 {
		log.Fatalf("GEMINI_DAILY_LIMIT must be a positive number, not %q", os.Getenv("GEMINI_DAILY_LIMIT"))
	}

	r := app.API(app.Deps{
		Favourites:       &favourites.DynamoStore{DB: db},
		Picks:            &picks.DynamoStore{DB: db},
		Movies:           tmdb.New(token),
		Picker:           picker,
		GeminiDailyLimit: limit,
		Now:              time.Now,
	})
	r.LogRequests = true
	lambda.Start(func(ctx context.Context, req router.Request) (router.Response, error) {
		lc, _ := lambdacontext.FromContext(ctx)
		if lc == nil || lc.Identity.CognitoIdentityID == "" {
			return router.Failed(apperr.Unauthenticated("Sign in first.")), nil
		}
		return r.Handle(user.With(ctx, lc.Identity.CognitoIdentityID), req)
	})
}
