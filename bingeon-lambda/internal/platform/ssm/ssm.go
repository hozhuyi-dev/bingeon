// Package ssm reads the keys bingeon keeps in Parameter Store as
// SecureStrings (written by `pnpm secrets:set`).
package ssm

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Secret reads the SecureString named by the environment variable env
// (e.g. TMDB_TOKEN_PARAM), decrypted: once, at a cold start.
func Secret(ctx context.Context, env string) (string, error) {
	name := os.Getenv(env)
	if name == "" {
		return "", fmt.Errorf("%s is not set", env)
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}
	out, err := awsssm.NewFromConfig(cfg).GetParameter(ctx, &awsssm.GetParameterInput{
		Name: aws.String(name), WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", name, err)
	}
	return aws.ToString(out.Parameter.Value), nil
}
