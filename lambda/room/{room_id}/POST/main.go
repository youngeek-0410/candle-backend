package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"net/http"
	"os"
)

type Answer struct {
	QuestionID string `json:"question_id" dynamodbav:"question_id"`
	Answer     bool   `json:"answer" dynamodbav:"answer"`
}
type UserData struct {
	UserID   string `json:"user_id" dynamodbav:"user_id"`
	NickName string `json:"nickname" dynamodbav:"nickname"`
	RoomID   string   `json:"room_id" dynamodbav:"room_id"`
	Answers  []Answer `json:"answers" dynamodbav:"answers"`
}

type requestBody struct {
	NickName string   `json:"nickname"`
	Answers  []Answer `json:"answers"`
}

func enterRoomHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomId := event.PathParameters["room_id"]
	if roomId == "" {
		return createEmptyResponseWithStatus(400, "Incorrect path parameter")
	}

	var req requestBody
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return createEmptyResponseWithStatus(500, "JSON parse error")
	}

	//リクエストボディにuser_idは含まれていないので新しい構造体を使ってデータ挿入
	var userData UserData
	userId := uuid.New()

	userData.UserID = userId.String()
	userData.NickName = req.NickName
	userData.Answers = req.Answers
	userData.RoomID = roomId

	// 書き込み処理
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		createEmptyResponseWithStatus(500, "")
	}

	if err = insertUserDataToCandleBackendUserTable(cfg, ctx, userData); err != nil {
		createEmptyResponseWithStatus(500, "Data write error.")
	}

	jsonUserData, err := json.Marshal(userData)
	if err != nil {
		createEmptyResponseWithStatus(500, "JSON parse error.")
	}
	return events.APIGatewayProxyResponse{
		Body:       string(jsonUserData),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func createEmptyResponseWithStatus(statusCode int, responseMessage string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       responseMessage,
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func insertUserDataToCandleBackendUserTable(cfg aws.Config, ctx context.Context, userData UserData) error {
	svc := dynamodb.NewFromConfig(cfg)

	tableName := "CandleBackendUserTable"
	if t, exists := os.LookupEnv("USER_TABLE_NAME"); exists {
		tableName = t
	}

	var answers []types.AttributeValue
	for _, ans := range userData.Answers {
		ansMap := map[string]types.AttributeValue{
			"question_id": &types.AttributeValueMemberS{Value: ans.QuestionID},
			"answer":      &types.AttributeValueMemberBOOL{Value: ans.Answer},
		}
		answers = append(answers, &types.AttributeValueMemberM{Value: ansMap})
	}

	params := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"user_id":   &types.AttributeValueMemberS{Value: userData.UserID},
			"nickname":  &types.AttributeValueMemberS{Value: userData.NickName},
			"room_id": &types.AttributeValueMemberS{Value: userData.RoomID},
			"answers":   &types.AttributeValueMemberL{Value: answers},
		},
	}

	_, err := svc.PutItem(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(enterRoomHandler)
}
