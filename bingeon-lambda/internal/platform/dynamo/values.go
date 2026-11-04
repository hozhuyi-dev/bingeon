package dynamo

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// S is a string attribute value.
func S(s string) types.AttributeValue { return &types.AttributeValueMemberS{Value: s} }

// N is a number attribute value.
func N(n int64) types.AttributeValue { return &types.AttributeValueMemberN{Value: itoa(n)} }
