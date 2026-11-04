// Package dynamo is a thin layer over the bingeon table for the slices'
// stores: items are plain attributes (attributevalue), every call says
// what it consumed (see meter), and a write conditional on a version is
// one helper (Versioned).
package dynamo

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"bingeon-lambda/internal/platform/meter"
)

// Item is an item as DynamoDB holds it.
type Item = map[string]types.AttributeValue

type Client struct {
	DB    *dynamodb.Client
	Table string
}

// New opens the table named by TABLE_NAME.
func New(ctx context.Context) (*Client, error) {
	name := os.Getenv("TABLE_NAME")
	if name == "" {
		return nil, errors.New("TABLE_NAME is not set")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(retryer))
	if err != nil {
		return nil, err
	}
	return &Client{DB: dynamodb.NewFromConfig(cfg), Table: name}, nil
}

// retryer rides out the small table's throttling: more attempts than the
// SDK's 3, backing off up to two seconds, and no client-side retry quota,
// which a burst of throttled calls would otherwise empty for every call
// after it.
func retryer() aws.Retryer {
	return retry.NewStandard(func(o *retry.StandardOptions) {
		o.MaxAttempts = 6
		o.MaxBackoff = 2 * time.Second
		o.RateLimiter = ratelimit.None
	})
}

// Key is an item's key.
func Key(pk, sk string) Item {
	return Item{"PK": &types.AttributeValueMemberS{Value: pk}, "SK": &types.AttributeValueMemberS{Value: sk}}
}

// Get reads one item into out; false when there is none.
func (c *Client) Get(ctx context.Context, pk, sk string, consistent bool, out any) (bool, error) {
	res, err := c.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &c.Table, Key: Key(pk, sk), ConsistentRead: aws.Bool(consistent),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		return false, err
	}
	consumed(ctx, "rcu", res.ConsumedCapacity)
	if res.Item == nil {
		return false, nil
	}
	return true, attributevalue.UnmarshalMap(res.Item, out)
}

// Query reads every item in a partition whose sort key starts with prefix,
// eventually consistent.
func (c *Client) Query(ctx context.Context, pk, prefix string) ([]Item, error) {
	in := &dynamodb.QueryInput{
		TableName:              &c.Table,
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: Item{
			":pk": &types.AttributeValueMemberS{Value: pk}, ":prefix": &types.AttributeValueMemberS{Value: prefix},
		},
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}
	var out []Item
	pages := dynamodb.NewQueryPaginator(c.DB, in)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		consumed(ctx, "rcu", page.ConsumedCapacity)
		out = append(out, page.Items...)
	}
	return out, nil
}

// Write runs writes as one: a single one by itself, several as a
// transaction (all or none, at twice the units).
func (c *Client) Write(ctx context.Context, writes ...types.TransactWriteItem) error {
	if len(writes) == 1 {
		return c.writeOne(ctx, writes[0])
	}
	for _, w := range writes {
		setTable(w, &c.Table)
	}
	res, err := c.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: writes, ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		return err
	}
	for _, cc := range res.ConsumedCapacity {
		consumed(ctx, "wcu", &cc)
	}
	return nil
}

func (c *Client) writeOne(ctx context.Context, w types.TransactWriteItem) error {
	total := types.ReturnConsumedCapacityTotal
	switch {
	case w.Put != nil:
		res, err := c.DB.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &c.Table, Item: w.Put.Item, ConditionExpression: w.Put.ConditionExpression,
			ExpressionAttributeNames: w.Put.ExpressionAttributeNames, ExpressionAttributeValues: w.Put.ExpressionAttributeValues,
			ReturnConsumedCapacity: total,
		})
		if err != nil {
			return err
		}
		consumed(ctx, "wcu", res.ConsumedCapacity)
	case w.Update != nil:
		res, err := c.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: &c.Table, Key: w.Update.Key, UpdateExpression: w.Update.UpdateExpression,
			ConditionExpression: w.Update.ConditionExpression, ExpressionAttributeNames: w.Update.ExpressionAttributeNames,
			ExpressionAttributeValues: w.Update.ExpressionAttributeValues, ReturnConsumedCapacity: total,
		})
		if err != nil {
			return err
		}
		consumed(ctx, "wcu", res.ConsumedCapacity)
	case w.Delete != nil:
		res, err := c.DB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &c.Table, Key: w.Delete.Key, ConditionExpression: w.Delete.ConditionExpression,
			ExpressionAttributeNames: w.Delete.ExpressionAttributeNames, ExpressionAttributeValues: w.Delete.ExpressionAttributeValues,
			ReturnConsumedCapacity: total,
		})
		if err != nil {
			return err
		}
		consumed(ctx, "wcu", res.ConsumedCapacity)
	default:
		return errors.New("dynamo: a write with nothing to do")
	}
	return nil
}

func setTable(w types.TransactWriteItem, table *string) {
	switch {
	case w.Put != nil:
		w.Put.TableName = table
	case w.Update != nil:
		w.Update.TableName = table
	case w.Delete != nil:
		w.Delete.TableName = table
	case w.ConditionCheck != nil:
		w.ConditionCheck.TableName = table
	}
}

// DeleteAll deletes items by key, 25 at a time, not as one: what a failure
// leaves is deleted by trying again.
func (c *Client) DeleteAll(ctx context.Context, keys []Item) error {
	for start := 0; start < len(keys); start += 25 {
		var requests []types.WriteRequest
		for _, key := range keys[start:min(start+25, len(keys))] {
			requests = append(requests, types.WriteRequest{DeleteRequest: &types.DeleteRequest{Key: key}})
		}
		pending := map[string][]types.WriteRequest{c.Table: requests}
		for attempt := 0; len(pending[c.Table]) > 0; attempt++ {
			if attempt == 5 {
				return errors.New("dynamo: deletes still throttled after 5 attempts")
			}
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
			}
			res, err := c.DB.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: pending, ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
			})
			if err != nil {
				return err
			}
			for _, cc := range res.ConsumedCapacity {
				consumed(ctx, "wcu", &cc)
			}
			pending = res.UnprocessedItems
		}
	}
	return nil
}

// Versioned is a put of item, whose `version` attribute the caller has
// moved on from read, on condition the stored one is still at read (0:
// there is none yet).
func Versioned(item any, read int64) (types.TransactWriteItem, error) {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return types.TransactWriteItem{}, err
	}
	put := &types.Put{Item: av}
	if read == 0 {
		put.ConditionExpression = aws.String("attribute_not_exists(PK)")
	} else {
		put.ConditionExpression = aws.String("version = :read")
		put.ExpressionAttributeValues = Item{":read": &types.AttributeValueMemberN{Value: itoa(read)}}
	}
	return types.TransactWriteItem{Put: put}, nil
}

// Failed are the indexes of the writes whose condition failed, in a
// Write's order; none when err is something else.
func Failed(err error) []int {
	var single *types.ConditionalCheckFailedException
	if errors.As(err, &single) {
		return []int{0}
	}
	var tx *types.TransactionCanceledException
	if !errors.As(err, &tx) {
		return nil
	}
	var out []int
	for i, reason := range tx.CancellationReasons {
		if aws.ToString(reason.Code) == "ConditionalCheckFailed" {
			out = append(out, i)
		}
	}
	return out
}

func consumed(ctx context.Context, name string, cc *types.ConsumedCapacity) {
	if cc != nil && cc.CapacityUnits != nil {
		meter.Add(ctx, name, *cc.CapacityUnits)
	}
}
