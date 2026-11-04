package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"bingeon-lambda/internal/platform/dynamo"
)

// Table is the real table named by TABLE_NAME (with AWS credentials), for
// a store's test to run against too; nil when none is named.
func Table(t *testing.T) *dynamo.Client {
	if os.Getenv("TABLE_NAME") == "" {
		return nil
	}
	db, err := dynamo.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// TestUser is a user id no one else has: a test on a real table writes
// under it, and deletes what it wrote.
func TestUser() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "TEST-" + hex.EncodeToString(b)
}
