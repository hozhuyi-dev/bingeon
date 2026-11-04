package picks

import (
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"bingeon-lambda/internal/platform/day"
	"bingeon-lambda/internal/platform/dynamo"
)

// DynamoStore keeps picks in the table: a PICKS item in the user's
// partition, and a QUOTA#<day> item counting the day's Gemini calls.
type DynamoStore struct {
	DB *dynamo.Client
}

const (
	picksSK = "PICKS"
	quotaSK = "GEMINI"
	// quotaKept is how long a day's count outlives the day, then TTL
	// deletes it (free).
	quotaKept = 48 * time.Hour
)

func partition(user string) string { return "USER#" + user }

func quotaPK(day string) string { return "QUOTA#" + day }

type picksItem struct {
	PK      string     `dynamodbav:"PK"`
	SK      string     `dynamodbav:"SK"`
	Day     string     `dynamodbav:"day"`
	Made    int        `dynamodbav:"made"`
	Last    []pickItem `dynamodbav:"last"`
	MadeAt  time.Time  `dynamodbav:"madeAt"`
	Version int64      `dynamodbav:"version"`
}

type pickItem struct {
	ID     int64  `dynamodbav:"id"`
	Title  string `dynamodbav:"title"`
	Year   int    `dynamodbav:"year,omitempty"`
	Poster string `dynamodbav:"poster,omitempty"`
	Reason string `dynamodbav:"reason"`
}

type quotaItem struct {
	Used int `dynamodbav:"n"`
}

func (s *DynamoStore) Get(ctx context.Context, user string) (Picks, error) {
	var it picksItem
	if _, err := s.DB.Get(ctx, partition(user), picksSK, true, &it); err != nil {
		return Picks{}, err
	}
	p := Picks{Day: it.Day, Made: it.Made, MadeAt: it.MadeAt, Version: it.Version}
	for _, pk := range it.Last {
		p.Last = append(p.Last, Pick(pk))
	}
	return p, nil
}

func (s *DynamoStore) Reserve(ctx context.Context, user string, read, next Picks, today string, limit int) error {
	put, err := s.put(user, read, next)
	if err != nil {
		return err
	}
	expires := day.End(time.Now()).Add(quotaKept).Unix()
	count := types.TransactWriteItem{Update: &types.Update{
		Key:                      dynamo.Key(quotaPK(today), quotaSK),
		UpdateExpression:         aws.String("ADD #n :one SET expires = :expires"),
		ConditionExpression:      aws.String("attribute_not_exists(#n) OR #n < :limit"),
		ExpressionAttributeNames: map[string]string{"#n": "n"},
		ExpressionAttributeValues: dynamo.Item{
			":one": dynamo.N(1), ":limit": dynamo.N(int64(limit)), ":expires": dynamo.N(expires),
		},
	}}
	err = s.DB.Write(ctx, put, count)
	failed := dynamo.Failed(err)
	switch {
	case slices.Contains(failed, 1):
		return ErrQuotaFull
	case len(failed) > 0:
		return ErrConflict
	}
	return err
}

func (s *DynamoStore) Save(ctx context.Context, user string, read, next Picks) error {
	put, err := s.put(user, read, next)
	if err != nil {
		return err
	}
	err = s.DB.Write(ctx, put)
	if len(dynamo.Failed(err)) > 0 {
		return ErrConflict
	}
	return err
}

func (s *DynamoStore) QuotaUsed(ctx context.Context, today string) (int, error) {
	var it quotaItem
	_, err := s.DB.Get(ctx, quotaPK(today), quotaSK, false, &it)
	return it.Used, err
}

func (s *DynamoStore) Delete(ctx context.Context, user string) error {
	return s.DB.DeleteAll(ctx, []dynamo.Item{dynamo.Key(partition(user), picksSK)})
}

// put is the write of next, on condition the picks are still at read.
func (s *DynamoStore) put(user string, read, next Picks) (types.TransactWriteItem, error) {
	it := picksItem{PK: partition(user), SK: picksSK, Day: next.Day, Made: next.Made, MadeAt: next.MadeAt, Version: next.Version}
	for _, p := range next.Last {
		it.Last = append(it.Last, pickItem(p))
	}
	return dynamo.Versioned(it, read.Version)
}
