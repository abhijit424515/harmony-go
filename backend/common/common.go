package common

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	Ctx context.Context
	Dbc *dynamodb.Client
)
