package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type response struct {
	Result bool `json:"result"`
}
type room struct {
	RoomId       string   `json:"room_id" dynamodbav:"room_id"`
	Participants []string `json:"participants" dynamodbav:"participants"`
}
type user struct {
	UserId   string   `json:"user_id" dynamodbav:"user_id"`
	Nickname string   `json:"nickname" dynamodbav:"nickname"`
	RoomId   string   `json:"room_id" dynamodbav:"room_id"`
	Fired    bool     `json:"fired" dynamodbav:"fired"`
	Answers  []answer `json:"answers" dynamodbav:"answers"`
}
type answer struct {
	QuestionId string `json:"question_id" dynamodbav:"question_id"`
	Answer     bool   `json:"answer" dynamodbav:"answer"`
}

var errorNotFired = errors.New("Not_fired")

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	const roomTableName = "CandleBackendRoomTable" // || os.LookupEnv("ROOM_TABLE_NAME")
	const userTableName = "CandleBackendUserTable" // || os.LookupEnv("USER_TABLE_NAME")
	roomId := event.PathParameters["room_id"]
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return createResponseWithStatus(http.StatusInternalServerError), err
	}
	//check if user exists and in room
	targetRoom, err := getRoom(cfg, roomId, roomTableName)
	if err != nil {
		return createResponseWithStatus(http.StatusInternalServerError), err
	}
	if targetRoom.RoomId == "" {
		fmt.Printf("INFO:room %v not found\n", roomId)
		return createResponseWithStatus(http.StatusNotFound), nil
	}
	numberParticipants := len(targetRoom.Participants)
	numberFired := 0
	for _, participant := range targetRoom.Participants {
		u, err := getUser(cfg, participant, userTableName)
		if err == errorNotFired {
			fmt.Printf("INFO:room %v, user %v not fired\n", roomId, participant)
			return createResponseWithStatus(http.StatusAccepted), nil
		} else if err != nil {
			return createResponseWithStatus(http.StatusInternalServerError), err
		}
		if u.Fired {
			numberFired++
		}
	}
	fmt.Printf("INFO:room %v, numberParticipants %v, numberFired %v\n", roomId, numberParticipants, numberFired)
	resp := response{
		Result: numberParticipants/2 < numberFired,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return createResponseWithStatus(http.StatusInternalServerError), err
	}
	return events.APIGatewayProxyResponse{
		Body:       string(jsonResp),
		StatusCode: http.StatusOK,

		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil

}

func main() {
	lambda.Start(handler)
}

func createResponseWithStatus(statuCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statuCode,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}
}
func getRoom(cfg aws.Config, roomId string, tableName string) (room, error) {
	svc := dynamodb.NewFromConfig(cfg)
	resp, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"room_id": &types.AttributeValueMemberS{Value: roomId},
		},
	})
	var r room
	if err != nil {
		return r, err
	}
	err = attributevalue.UnmarshalMap(resp.Item, &r)
	return r, err
}

func getUser(cfg aws.Config, userId string, tableName string) (user, error) {
	svc := dynamodb.NewFromConfig(cfg)
	resp, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if resp.Item["fired"] == nil {
		return user{}, errorNotFired
	}
	var u user
	if err != nil {
		return u, err
	}
	err = attributevalue.UnmarshalMap(resp.Item, &u)
	return u, err
}
