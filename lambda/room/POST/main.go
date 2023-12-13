package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// note:Response body
type request struct {
	RoomId string `json:"room_id"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req request
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return createEmptyResponseWithStatus(http.StatusInternalServerError), err
	}
	roomTableName := "CandleBackendRoomTable"
	if t, exists := os.LookupEnv("ROOM_TABLE_NAME"); exists {
		roomTableName = t
	}

	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return createEmptyResponseWithStatus(http.StatusInternalServerError), err
	}
	if req.RoomId == "" {
		fmt.Println("INFO:room_id is empty")
		return createEmptyResponseWithStatus(http.StatusBadRequest), nil
	}
	exists := createRoom(cfg, ctx, req, roomTableName)

	if exists {
		return createEmptyResponseWithStatus(http.StatusConflict), nil
	}

	return createEmptyResponseWithStatus(http.StatusCreated), nil
}

func main() {
	lambda.Start(handler)
}

func createEmptyResponseWithStatus(statuCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: statuCode,
	}
}

func createRoom(cfg aws.Config, ctx context.Context, req request, roomTableName string) bool {
	svc := dynamodb.NewFromConfig(cfg)
	ttl := time.Now().Add(12 * time.Hour).Unix()
	_, err := svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(roomTableName),
		Item: map[string]types.AttributeValue{
			"room_id":      &types.AttributeValueMemberS{Value: req.RoomId},
			"participants": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			"TTL":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%v", ttl)},
		},
		ConditionExpression: aws.String("attribute_not_exists(room_id)"),
	})
	return err != nil
}
