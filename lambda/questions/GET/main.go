package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Question struct {
	QuestionID int `json:"question_id" dynamodbav:"question_id"`
	Statement string `json:"statement" dynamodbav:"statement"`
}

type Response struct {
	Questions []Question `json:"questions"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("questions\n")

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return serverErrorResponse(fmt.Errorf("TABLE_NAME environment variable is not set"))
	}

	sess := session.Must(session.NewSession())

	svc := dynamodb.New(sess)

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error scanning DynamoDB table: %v", err))
	}

	questions := make([]Question, 0)
	for _, i := range result.Items {
		question := Question{}
		err := dynamodbattribute.UnmarshalMap(i, &question)
		if err != nil {
			return serverErrorResponse(fmt.Errorf("error unmarshalling item: %v", err))
		}
		questions = append(questions, question)
	}

	jsonResponse, err := json.Marshal(Response{Questions: questions})
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error marshalling items to JSON: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:		string(jsonResponse),
		Headers:	map[string]string{"Content-Type": "application/json"},
	}, nil
}

func serverErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:		`{"message": "Internal Server Error"}`,
		Headers:	map[string]string{"Content-Type": "application/json"},
	}, nil
}


func main() {
	lambda.Start(handler)
}
