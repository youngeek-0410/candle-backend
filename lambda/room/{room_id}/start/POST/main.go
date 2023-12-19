package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"os"
)

type Answer struct {
	QuestionID string `json:"question_id" dynamodbav:"question_id"`
	Answer     bool   `json:"answer" dynamodbav:"answer"`
}

type UserData struct {
	UserID   string   `json:"user_id" dynamodbav:"user_id"`
	NickName string   `json:"nickname" dynamodbav:"nickname"`
	RoomID   string   `json:"room_id" dynamodbav:"room_id"`
	Answers  []Answer `json:"answers" dynamodbav:"answers"`
}

func gameStartHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomID := event.PathParameters["room_id"]
	if roomID == "" {
		createEmptyResponseWithStatus(400, "Incorrect path parameter")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		createEmptyResponseWithStatus(500, "Internal server error")
	}

	result, err := getAllQuestionAnswers(cfg, ctx, roomID)
	if err != nil {
		createEmptyResponseWithStatus(500, "DB get error")
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		createEmptyResponseWithStatus(500, "JSON parse error")
	}

	return events.APIGatewayProxyResponse{
		Body:       string(jsonResult),
		StatusCode: 200,
	}, nil
}

func getAllQuestionAnswers(cfg aws.Config, ctx context.Context, roomID string) (*UserData, error) {
	svc := dynamodb.NewFromConfig(cfg)
	tableName := "CandleBackendUserTable"
	if t, exists := os.LookupEnv("USER_TABLE_NAME"); exists {
		tableName = t
	}

	key := map[string]types.AttributeValue{
		"room_id": &types.AttributeValueMemberS{Value: roomID},
	}

	params := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	}

	allUserData, err := svc.GetItem(ctx, params)
	if err != nil {
		return nil, err
	}

	if allUserData.Item == nil {
		return nil, err
	}

	var results UserData
	err = attributevalue.UnmarshalMap(allUserData.Item, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}

func createEmptyResponseWithStatus(statusCode int, responseMessage string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       responseMessage,
		StatusCode: statusCode,
	}, nil
}

func main() {
	lambda.Start(gameStartHandler)
}
