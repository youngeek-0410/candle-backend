package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
type RequestBody struct {
	UserID string `json:"user_id" dynamodbav:"user_id"`
}

type ResponseBody struct {
	UserID              string `json:"user_id"`
	IsSanta             bool   `json:"is_santa"`
	QuestionID          string `json:"question_id"`
	QuestionDescription string `json:"question_description"`
}

type Question struct {
	QuestionID int    `json:"question_id" dynamodbav:"question_id"`
	Statement  string `json:"statement" dynamodbav:"statement"`
}

type QuestionResponse struct {
	Questions []Question `json:"questions"`
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
func ReturnSantaCandidateList(users []UserData) []UserData {
	//サンタ疑惑のあるユーザーリストの返却
	falseCountByUser := make(map[string]int)
	var santaCandidateList []UserData
	maxFalseCount := -1
	for _, user := range users {
		for _, answer := range user.Answers {
			if !answer.Answer {
				falseCountByUser[user.UserID]++
			}
		}
		if falseCountByUser[user.UserID] > maxFalseCount {
			santaCandidateList = []UserData{user}
			maxFalseCount = falseCountByUser[user.UserID]
		} else if falseCountByUser[user.UserID] == maxFalseCount {
			santaCandidateList = append(santaCandidateList, user)
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

func DecidingSantaAndQuestion(santaCandidateList []UserData, allUserData []UserData) (string, string) {
	trueQueMap := returnNumberOfTrueForEachQuestion(allUserData)

	var maxTrueCount int
	maxTrueCount = -1
	var torchQuestionID string
	var santaUserID string

	for _, santaData := range santaCandidateList {
		for _, answer := range santaData.Answers {
			trueCount := trueQueMap[answer.QuestionID]
			if trueCount == 0 {
				continue
			}
			if (!answer.Answer) && (trueCount > maxTrueCount) {
				maxTrueCount = trueCount
				torchQuestionID = answer.QuestionID
				santaUserID = santaData.UserID
			}
		}
	}
	return santaUserID, torchQuestionID
}

func createErrorResponseWithStatus(statusCode int, responseMessage string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       responseMessage,
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func getQuestionDescriptionFromQuestionID(cfg aws.Config, ctx context.Context, questionID int) (string, error) {
	tableName := "CandleBackendQuestionTable"
	if t, exists := os.LookupEnv("QUESTION_TABLE_NAME"); exists {
		tableName = t
	}

	svc := dynamodb.NewFromConfig(cfg)

	response, err := svc.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		fmt.Println("Error scanning DynamoDB table:", err)
		return "", err
	}

	var queRes []Question
	err = attributevalue.UnmarshalListOfMaps(response.Items, &queRes)
	if err != nil {
		fmt.Println("Error unmarshaling DynamoDB response:", err)
		return "", err
	}

	fmt.Println(queRes)

	for _, que := range queRes {
		if que.QuestionID == questionID {
			return que.Statement, nil
		}
	}
	return "", errors.New("Invalid QuestionID")
}

func gameStartHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomID := event.PathParameters["room_id"]
	if roomID == "" {
		return createErrorResponseWithStatus(400, "Incorrect path parameter")
	}

	var req RequestBody
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return createErrorResponseWithStatus(500, "JSON parse error")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return createErrorResponseWithStatus(500, "Internal server error")
	}

	if err != nil {
		return createErrorResponseWithStatus(500, err.Error())
	}

	roomResult, err := getRoomData(cfg, ctx, roomID)
	if err != nil {
		return createErrorResponseWithStatus(500, "DB get error")
	}

	allUserData, err := getAllUserData(cfg, ctx, roomResult.Participants)
	if err != nil {
		return createErrorResponseWithStatus(500, err.Error())
	}

	santaCandidateList := ReturnSantaCandidateList(allUserData)
	santaUserID, torchQuestionID := DecidingSantaAndQuestion(santaCandidateList, allUserData)
	fmt.Println("torch")
	fmt.Println(torchQuestionID)

	var responseBody ResponseBody
	//先ほど取得したサンタのuser_idとリクエストボディのuser_idが一致したらサンタである
	if santaUserID == req.UserID {
		responseBody.IsSanta = true
	} else {
		responseBody.IsSanta = false
	}
	responseBody.QuestionID = torchQuestionID

	intQueID, err := strconv.Atoi(responseBody.QuestionID)
	if err != nil {
		return createErrorResponseWithStatus(500, err.Error())
	}
	description, err := getQuestionDescriptionFromQuestionID(cfg, ctx, intQueID)
	if err != nil {
		return createErrorResponseWithStatus(500, err.Error())
	}

	responseBody.UserID = req.UserID
	responseBody.QuestionDescription = description

	json, _ := json.Marshal(responseBody)

	return events.APIGatewayProxyResponse{
		Body:       string(json),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func main() {
	lambda.Start(gameStartHandler)
}
