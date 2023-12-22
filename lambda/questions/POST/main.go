package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Question struct {
	QuestionID int    `json:"question_id" dynamodbav:"question_id"`
	Statement  string `json:"statement" dynamodbav:"statement"`
}

type requestBody struct {
	Questions []Question `json:"questions"`
}

type response struct {
	Questions []Question `json:"questions"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("TABLE_NAME")
	fmt.Println(tableName)
	if tableName == "" {
		return serverErrorResponse(fmt.Errorf("TABLE_NAME environment variable is not set"))
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return serverErrorResponse(err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	var req requestBody
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return serverErrorResponse(err)
	}
	fmt.Println("request body: %v", req)

	for _, q := range req.Questions {
		fmt.Println(strconv.Itoa(q.QuestionID), q.Statement)
		_, err := svc.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]types.AttributeValue{
				"question_id": &types.AttributeValueMemberN{Value: strconv.Itoa(q.QuestionID)},
				"statement":   &types.AttributeValueMemberS{Value: q.Statement},
			},
			//ConditionExpression: aws.String("attribute_not_exists(question_id)"),
		})
		if err != nil {
			return serverErrorResponse(fmt.Errorf("error putting questions to dynamodb: %v", err))
		}
	}

	jsonResponse, err := json.Marshal(response{Questions: req.Questions})
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error marshalling items to JSON: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResponse),
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

func main() {
	lambda.Start(handler)
}
