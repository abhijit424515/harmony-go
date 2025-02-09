package handlers

import (
	"encoding/base64"
	"fmt"
	"harmony/backend/common"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v4"
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

func GetBuffer(userid string) ([]byte, BufType, int64, error) {
	items, err := common.Dbc.Query(common.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("buffer"),
		IndexName:              aws.String("userid-index"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userid},
		},
	})
	if err != nil {
		return nil, TextType, 0, err
	}

	if items.Count == 0 {
		return nil, TextType, 0, fmt.Errorf("no buffer found")
	}

	buf := items.Items[0]["data"].(*types.AttributeValueMemberB).Value
	ttl_ := items.Items[0]["ttl"].(*types.AttributeValueMemberN).Value
	buf_type := items.Items[0]["type"].(*types.AttributeValueMemberS).Value

	ttl, _ := strconv.ParseInt(ttl_, 10, 64)
	if time.Now().Unix() > ttl {
		return nil, TextType, 0, fmt.Errorf("buffer expired")
	}

	bt := TextType
	if buf_type == string(ImageType) {
		bt = ImageType
	}

	return buf, bt, ttl, nil
}

func UpsertBuffer(userid string, data []byte, t BufType) (int64, error) {
	items, err := common.Dbc.Query(common.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("buffer"),
		IndexName:              aws.String("userid-index"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userid},
		},
	})
	if err != nil {
		return 0, err
	}

	var _id string
	if items.Count > 0 {
		_id = items.Items[0]["_id"].(*types.AttributeValueMemberS).Value
	} else {
		_id = uuid.New().String()
	}

	ttl := time.Now().Add(common.Lifetime).Unix()

	item, err := attributevalue.MarshalMap(Buffer{
		Id:     _id,
		UserId: userid,
		Time:   time.Now().Unix(),
		Ttl:    ttl,
		Type:   t,
		Data:   data,
	})
	if err != nil {
		return 0, err
	}

	_, err = common.Dbc.PutItem(common.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("buffer"), Item: item,
	})
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

func CreateOrGetUser(email string) (string, error) {
	items, err := common.Dbc.Query(common.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("user"),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return "", err
	}

	if items.Count > 0 {
		return items.Items[0]["_id"].(*types.AttributeValueMemberS).Value, nil
	}

	uid := uuid.New().String()
	item, err := attributevalue.MarshalMap(User{
		Id:    uid,
		Email: email,
	})
	if err != nil {
		return "", err
	}

	_, err = common.Dbc.PutItem(common.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("user"), Item: item, ConditionExpression: aws.String("attribute_not_exists(email)"),
	})
	if err != nil {
		return "", err
	}

	return uid, nil
}

func GenerateAccessToken(payload map[string]interface{}, expirationTime time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	for key, value := range payload {
		claims[key] = value
	}

	t := time.Now()
	claims["exp"] = t.Add(expirationTime).Unix()
	claims["iat"] = t.Unix()

	decodedKey, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SK"))
	if err != nil {
		return "", fmt.Errorf("[error] decoding signing key: %v", err)
	}

	tokenString, err := token.SignedString(decodedKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyAndDecodeToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		decodedKey, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SK"))
		if err != nil {
			return "", fmt.Errorf("[error] decoding signing key: %v", err)
		}
		return decodedKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token has expired")
			}
		} else {
			return nil, fmt.Errorf("missing expiration claim")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
