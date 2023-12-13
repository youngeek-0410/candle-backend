package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

type Question struct {
	Question_id int `json:"question_id"`
	Statement string `json:"statement"`
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

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error loading AWS configuration: %v", err))
	}

	svc := dynamodb.NewFromConfig(cfg)

	result, err := svc.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return serverErrorResponse(fmt.Errorf("error scanning DynamoDB table: %v", err))
	}

	questions := make([]Question, 0)
	for _, i := range result.Items {
		question := Question{}
		err := attributevalue.UnmarshalMap(i, &question)
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
