package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

type requestBody struct {
	UserID     string `json:"user_id"`
	FireUserID string `json:"fire_user_id"`
	QuestionID int    `json:"question_id"`
}

type answer struct {
	QuestionID int  `json:"question_id" dynamodbav:"question_id"`
	Answer     bool `json:"answer" dynamodbav:"answer"`
}

type user struct {
	UserID   string   `json:"user_id" dynamodbav:"user_id"`
	Answers  []answer `json:"answers" dynamodbav:"answers"`
	Nickname string   `json:"nickname" dynamodbav:"nickname"`
	RoomID   string   `json:"room_id" dynamodbav:"room_id"`
	IsSanta  bool     `json:"is_santa" dynamodbav:"is_santa"`
}

type response struct {
	Fired bool `json:"fired"`
}

var svc *dynamodb.Client

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("room/{%s}/reult/%s\n", event.PathParameters["room_id"], event.HTTPMethod)
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return serverErrorResponse(fmt.Errorf("TABLE_NAME environment variable is not set"))
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return serverErrorResponse(err)
	}

	svc = dynamodb.NewFromConfig(cfg)

	var body requestBody
	err = json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return badRequestErrorResponse(err)
	}

	// 火を灯されたユーザの取得
	result, err := getUser(ctx, tableName, body.UserID)
	if err != nil {
		return badRequestErrorResponse(err)
	}
	var firedUser user
	err = attributevalue.UnmarshalMap(result.Item, &firedUser)
	if err != nil {
		serverErrorResponse(err)
	}

	roomID := event.PathParameters["room_id"]
	if roomID == "" {
		return badRequestErrorResponse(errors.New("empty path parameters"))
	}
	roomID, err = url.PathUnescape(roomID)
	if err != nil {
		return serverErrorResponse(errors.New("could not decode room_id"))
	}
	// ユーザがルームに入っているか
	if roomID != firedUser.RoomID {
		return badRequestErrorResponse(errors.New("user not found in the room"))
	}

	// 火を灯したユーザの取得
	result, err = getUser(ctx, tableName, body.FireUserID)
	if err != nil {
		return badRequestErrorResponse(err)
	}
	var fireUser user
	err = attributevalue.UnmarshalMap(result.Item, &fireUser)
	if err != nil {
		serverErrorResponse(err)
	}

	// ユーザがルームに入っているか
	if roomID != fireUser.RoomID {
		return badRequestErrorResponse(errors.New("user not found in the room"))
	}

	// 火を灯したユーザがサンタでなければ灯されたユーザの火は消えない
	is_fire := !fireUser.IsSanta
	// 火を灯したユーザがサンタでなくても質問が false なら火は消える
	if is_fire { // 配布される質問は全てサンタが false にしたもののはずだけど一応確認
		for _, ans := range fireUser.Answers {
			if ans.QuestionID == body.QuestionID {
				is_fire = ans.Answer
			}
		}
	}

	_, err = svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: body.UserID},
		},
		UpdateExpression: aws.String("set fire = :f, fired_by = :u"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":f": &types.AttributeValueMemberBOOL{Value: is_fire},
			":u": &types.AttributeValueMemberS{Value: fireUser.UserID},
		},
	})
	if err != nil {
		return serverErrorResponse(err)
	}

	resp, err := json.Marshal(response{Fired: is_fire})
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error marshalling items to JSON: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(resp),
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func serverErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       `{"message": "Internal Server Error"}`,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func badRequestErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       `{"message": "Bad Request"}`,
		Headers:    map[string]string{"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"},
	}, nil
}

func getUser(ctx context.Context, tableName string, userID string) (*dynamodb.GetItemOutput, error) {
	return svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},
	})
}

func main() {
	lambda.Start(handler)
}
