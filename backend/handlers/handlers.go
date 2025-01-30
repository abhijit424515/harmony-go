package handlers

import (
	"fmt"
	"harmony/backend/common"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type BufType string

const (
	TextType  BufType = "text"
	ImageType BufType = "image"
)

type User struct {
	Id    string `dynamodbav:"_id"`
	Email string `dynamodbav:"email"`
}

type Buffer struct {
	Id     string  `dynamodbav:"_id"`
	UserId string  `dynamodbav:"user_id"`
	Time   int64   `dynamodbav:"time"`
	Ttl    int64   `dynamodbav:"ttl"`
	Type   BufType `dynamodbav:"type"`
	Data   []byte  `dynamodbav:"data"`
}

func UpsertBuffer(userid string, data []byte, t BufType) error {
	items, err := common.Dbc.Query(common.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("buffer"),
		IndexName:              aws.String("userid-index"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userid},
		},
	})
	if err != nil {
		return err
	}

	var _id string
	if items.Count > 0 {
		_id = items.Items[0]["_id"].(*types.AttributeValueMemberS).Value
	} else {
		_id = uuid.New().String()
	}

	ttl := time.Now().Add(2 * time.Minute).Unix()

	item, err := attributevalue.MarshalMap(Buffer{
		Id:     _id,
		UserId: userid,
		Time:   time.Now().Unix(),
		Ttl:    ttl,
		Type:   t,
		Data:   data,
	})
	if err != nil {
		return err
	}

	_, err = common.Dbc.PutItem(common.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("buffer"), Item: item,
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateUser(email string) error {
	items, err := common.Dbc.Query(common.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("user"),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return err
	}

	if items.Count > 0 {
		fmt.Println("[error] Email already in use")
		return nil
	}

	item, err := attributevalue.MarshalMap(User{
		Id:    uuid.New().String(),
		Email: email,
	})
	if err != nil {
		return err
	}

	_, err = common.Dbc.PutItem(common.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("user"), Item: item, ConditionExpression: aws.String("attribute_not_exists(email)"),
	})
	if err != nil {
		return err
	}

	return nil
}
