package main

import (
	"context"
	"encoding/json"
	"errors"
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

type RequestBody struct {
	UserID string `json:"user_id" dynamodbav:"user_id"`
}

type ResponseBody struct {
	UserID              string `json:"user_id"`
	IsSanta             bool   `json:"is_santa"`
	QuestionID          string `json:"question_id"`
	QuestionDescription string `json:"question_description"`
}
//type Question struct {
//	QuestionID int    `json:"question_id" dynamodbav:"question_id"`
//	Statement  string `json:"statement" dynamodbav:"statement"`
//}
//type QuestionResponse struct {
//	Questions []Question `json:"questions"`
//}
//
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
	// edit by ishibe
	var santaCandidateList []UserData
	maxFalseCount := -1
	//
	for _, user := range users {
		for _, answer := range user.Answers {
			if !answer.Answer {
				falseCountByUser[user.UserID]++
			}
		}
		// edit by ishibe
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
	}, nil
}

//func getQuestionDescriptionFromQuestionID(cfg aws.Config, ctx context.Context, questionID int) (string, error) {
//	tableName := "CandleBackendQuestionTable"
//	if t, exists := os.LookupEnv("QUESTION_TABLE_NAME"); exists {
//		tableName = t
//	}
//
//	svc := dynamodb.NewFromConfig(cfg)
//
//	response, err := svc.Scan(ctx, &dynamodb.ScanInput{
//		TableName: aws.String(tableName),
//	})
//	if err != nil {
//		return "", err
//	}
//	var queRes QuestionResponse
//	err = attributevalue.UnmarshalListOfMaps(response.Items, &queRes)
//	if err != nil {
//		return "", err
//	}
//}


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

	var responseBody ResponseBody
	//先ほど取得したサンタのuser_idとリクエストボディのuser_idが一致したらサンタである
	if santaUserID == req.UserID {
		responseBody.IsSanta = true
		responseBody.QuestionID = torchQuestionID
	} else {
		responseBody.IsSanta = false
		responseBody.QuestionID = torchQuestionID
	}

	responseBody.UserID = req.UserID
	responseBody.QuestionDescription = "ちいかわは好きですか。"

	json,_ := json.Marshal(responseBody)

	return events.APIGatewayProxyResponse{
		Body:       string(json),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(gameStartHandler)
}
