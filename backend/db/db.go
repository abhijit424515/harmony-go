package db

import (
	"context"
	"errors"
	"fmt"
	"harmony/backend/common"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func createTableIfNotExists(ctx context.Context, client *dynamodb.Client, tableName string, keySchema []types.KeySchemaElement, attributes []types.AttributeDefinition, gsi []types.GlobalSecondaryIndex) error {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
	if err == nil {
		log.Printf("Table %s already exists, skipping creation.\n", tableName)
		return nil
	}

	var notFoundErr *types.ResourceNotFoundException
	b := errors.As(err, &notFoundErr)
	if !b {
		return err // Return error if it's something other than table not existing
	}

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:              aws.String(tableName),
		KeySchema:              keySchema,
		AttributeDefinitions:   attributes,
		BillingMode:            types.BillingModePayPerRequest,
		GlobalSecondaryIndexes: gsi,
	})
	if err != nil {
		return fmt.Errorf("[error] failed to create table %s: %w", tableName, err)
	}

	log.Printf("Table %s created successfully.\n", tableName)
	return nil
}

func setupTables() {
	userTKS := []types.KeySchemaElement{
		{AttributeName: aws.String("_id"), KeyType: types.KeyTypeHash},
	}
	userTA := []types.AttributeDefinition{
		{AttributeName: aws.String("_id"), AttributeType: types.ScalarAttributeTypeS},
		{AttributeName: aws.String("email"), AttributeType: types.ScalarAttributeTypeS},
	}
	userGSI := []types.GlobalSecondaryIndex{
		{
			IndexName: aws.String("email-index"),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("email"), KeyType: types.KeyTypeHash},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeAll,
			},
		},
	}

	bufferTKS := []types.KeySchemaElement{
		{AttributeName: aws.String("_id"), KeyType: types.KeyTypeHash},
	}
	bufferTA := []types.AttributeDefinition{
		{AttributeName: aws.String("_id"), AttributeType: types.ScalarAttributeTypeS},
		{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
	}
	bufferGSI := []types.GlobalSecondaryIndex{
		{
			IndexName: aws.String("userid-index"),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeAll,
			},
		},
	}

	if err := createTableIfNotExists(common.Ctx, common.Dbc, "user", userTKS, userTA, userGSI); err != nil {
		log.Fatalf("[error] creating user table: %v", err)
	}
	if err := createTableIfNotExists(common.Ctx, common.Dbc, "buffer", bufferTKS, bufferTA, bufferGSI); err != nil {
		log.Fatalf("[error] creating buffer table: %v", err)
	}
}

func Setup() error {
	cfg, err := config.LoadDefaultConfig(common.Ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
		return err
	}

	common.Dbc = dynamodb.NewFromConfig(cfg)
	setupTables()
	return nil
}
