package favourites

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"bingeon-lambda/internal/platform/dynamo"
)

// DynamoStore keeps favourites in the table: the user's partition holds a
// FAVOURITES item (the stats) and a FAV#<movieId> item per favourite.
type DynamoStore struct {
	DB *dynamo.Client
}

const (
	statsSK   = "FAVOURITES"
	favPrefix = "FAV#"
)

func partition(user string) string { return "USER#" + user }

func favSK(id int64) string { return favPrefix + strconv.FormatInt(id, 10) }

type statsItem struct {
	PK      string `dynamodbav:"PK"`
	SK      string `dynamodbav:"SK"`
	Count   int    `dynamodbav:"count"`
	Day     string `dynamodbav:"day"`
	Changes int    `dynamodbav:"changes"`
	Version int64  `dynamodbav:"version"`
}

type favItem struct {
	PK     string    `dynamodbav:"PK"`
	SK     string    `dynamodbav:"SK"`
	ID     int64     `dynamodbav:"id"`
	Title  string    `dynamodbav:"title"`
	Year   int       `dynamodbav:"year,omitempty"`
	Poster string    `dynamodbav:"poster,omitempty"`
	Added  time.Time `dynamodbav:"added"`
}

func (it favItem) favourite() Favourite {
	return Favourite{ID: it.ID, Title: it.Title, Year: it.Year, Poster: it.Poster, Added: it.Added}
}

func (s *DynamoStore) Stats(ctx context.Context, user string) (Stats, error) {
	var it statsItem
	_, err := s.DB.Get(ctx, partition(user), statsSK, true, &it)
	return Stats{Count: it.Count, Day: it.Day, Changes: it.Changes, Version: it.Version}, err
}

func (s *DynamoStore) Get(ctx context.Context, user string, id int64) (Favourite, bool, error) {
	var it favItem
	ok, err := s.DB.Get(ctx, partition(user), favSK(id), true, &it)
	return it.favourite(), ok, err
}

func (s *DynamoStore) List(ctx context.Context, user string) ([]Favourite, error) {
	items, err := s.DB.Query(ctx, partition(user), favPrefix)
	if err != nil {
		return nil, err
	}
	out := make([]Favourite, 0, len(items))
	for _, raw := range items {
		var it favItem
		if err := attributevalue.UnmarshalMap(raw, &it); err != nil {
			return nil, err
		}
		out = append(out, it.favourite())
	}
	return out, nil
}

func (s *DynamoStore) Add(ctx context.Context, user string, read, next Stats, f Favourite) error {
	stats, err := s.stats(user, read, next)
	if err != nil {
		return err
	}
	fav, err := attributevalue.MarshalMap(favItem{
		PK: partition(user), SK: favSK(f.ID), ID: f.ID, Title: f.Title, Year: f.Year, Poster: f.Poster, Added: f.Added,
	})
	if err != nil {
		return err
	}
	return conflict(s.DB.Write(ctx, stats, types.TransactWriteItem{Put: &types.Put{Item: fav}}))
}

func (s *DynamoStore) Remove(ctx context.Context, user string, read, next Stats, id int64) error {
	stats, err := s.stats(user, read, next)
	if err != nil {
		return err
	}
	return conflict(s.DB.Write(ctx, stats, types.TransactWriteItem{Delete: &types.Delete{Key: dynamo.Key(partition(user), favSK(id))}}))
}

func (s *DynamoStore) DeleteAll(ctx context.Context, user string) error {
	items, err := s.DB.Query(ctx, partition(user), favPrefix)
	if err != nil {
		return err
	}
	keys := []dynamo.Item{dynamo.Key(partition(user), statsSK)}
	for _, raw := range items {
		var it favItem
		if err := attributevalue.UnmarshalMap(raw, &it); err != nil {
			return err
		}
		keys = append(keys, dynamo.Key(it.PK, it.SK))
	}
	return s.DB.DeleteAll(ctx, keys)
}

// stats is the write of next, on condition the stats are still at read.
func (s *DynamoStore) stats(user string, read, next Stats) (types.TransactWriteItem, error) {
	return dynamo.Versioned(statsItem{
		PK: partition(user), SK: statsSK, Count: next.Count, Day: next.Day, Changes: next.Changes, Version: next.Version,
	}, read.Version)
}

// conflict is a failed condition as ErrConflict: the stats moved on.
func conflict(err error) error {
	if len(dynamo.Failed(err)) > 0 {
		return ErrConflict
	}
	return err
}
