package main

import (
	"context"
	"errors"
	"fmt"
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

type RoomData struct {
	RoomID       string   `json:"room_id" dynamodbav:"room_id"`
	Participants []string `json:"participants" dynamodbav:"participants"`
}

func getRoomData(cfg aws.Config, ctx context.Context, roomID string) (RoomData, error) {
	svc := dynamodb.NewFromConfig(cfg)
	tableName := "CandleBackendRoomTable"
	if t, exists := os.LookupEnv("USER_TABLE_NAME"); exists {
		tableName = t
	}

	response, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"room_id": &types.AttributeValueMemberS{Value: roomID},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		return RoomData{}, err
	}

	var roomData RoomData
	if err = attributevalue.UnmarshalMap(response.Item, &roomData); err != nil {
		return RoomData{}, err
	}

	return roomData, nil
}

func getAllUserData(cfg aws.Config, ctx context.Context, userIDList []string) ([]UserData, error) {
	svc := dynamodb.NewFromConfig(cfg)
	tableName := "CandleBackendUserTable"
	if t, exists := os.LookupEnv("USER_TABLE_NAME"); exists {
		tableName = t
	}

	var allUserInfo []UserData
	for _, userID := range userIDList {
		var userInfo UserData
		response, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
			Key: map[string]types.AttributeValue{
				"user_id": &types.AttributeValueMemberS{Value: userID},
			},
			TableName: aws.String(tableName),
		})
		if err != nil {
			return nil, err
		}

		if response.Item == nil {
			return nil, errors.New("User data response is empty")
		}

		if err = attributevalue.UnmarshalMap(response.Item, &userInfo); err != nil {
			return nil, err
		}

		allUserInfo = append(allUserInfo, userInfo)
	}

	return allUserInfo, nil

}
func ReturnSantaCandidateList(users []UserData) []string {
	//サンタ疑惑のあるユーザーリストの返却
	falseCountByUser := make(map[string]int)
	for _, user := range users {
		for _, answer := range user.Answers {
			if !answer.Answer {
				falseCountByUser[user.UserID]++
			}
		}
	}

	var santaCandidateList []string
	maxFalseCount := -1

	for userName, userFalseCount := range falseCountByUser {
		if userFalseCount > maxFalseCount {
			santaCandidateList = []string{userName}
			maxFalseCount = userFalseCount
		} else if userFalseCount == maxFalseCount {
			santaCandidateList = append(santaCandidateList, userName)
		}
	}
	return santaCandidateList
}

func returnNumberOfTrueForEachQuestion(userData []UserData) map[string]int {
	totalCount := make(map[string]int)
	for _, user := range userData {
		for _, ans := range user.Answers {
			key := ans.QuestionID
			if ans.Answer {
				totalCount[key]++
			}
		}
	}
	return totalCount
}

//func DecidingSantaAndQuestion(santaCandidateList []string, allUserData []UserData) {
//	trueQueMap := returnNumberOfTrueForEachQuestion(allUserData)
//	fmt.Println(trueQueMap)
//}

func DecidingSantaAndQuestion(allUserData []UserData) {
	trueQueMap := returnNumberOfTrueForEachQuestion(allUserData)
	fmt.Println(trueQueMap)
}

func createErrorResponseWithStatus(statusCode int, responseMessage string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       responseMessage,
		StatusCode: statusCode,
	}, nil
}

func gameStartHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomID := event.PathParameters["room_id"]
	if roomID == "" {
		return createErrorResponseWithStatus(400, "Incorrect path parameter")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return createErrorResponseWithStatus(500, "Internal server error")
	}

	roomResult, err := getRoomData(cfg, ctx, roomID)
	if err != nil {
		return createErrorResponseWithStatus(500, "DB get error")
	}

	allUserData, err := getAllUserData(cfg, ctx, roomResult.Participants)
	if err != nil {
		return createErrorResponseWithStatus(500, err.Error())
	}

	//santaCandidateList := ReturnSantaCandidateList(allUserData)
	_ = ReturnSantaCandidateList(allUserData)
	DecidingSantaAndQuestion(allUserData)

	return events.APIGatewayProxyResponse{
		Body:       "ok",
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(gameStartHandler)
}
