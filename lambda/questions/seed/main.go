package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type QuestionItem struct {
	QuestionID int    `json:"question_id" dynamodbav:"question_id"`
	Statement  string `json:"statement" dynamodbav:"statement"`
}

func InsertData(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return "", nil, fmt.Errorf("TABLE_NAME environment variable is not set")
	}

	items := []QuestionItem{
		{QuestionID: 1, Statement: "『ワンピース』は好きですか？"},
		{QuestionID: 2, Statement: "『進撃の巨人』は好きですか？"},
		{QuestionID: 3, Statement: "『ちいかわ』は好きですか？"},
		{QuestionID: 4, Statement: "『ポケモン』は好きですか？"},
		{QuestionID: 5, Statement: "温泉は好きですか？"},
	}

	for _, item := range items {
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return "", nil, fmt.Errorf("failed to marshal item: %v", err)
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			return "", nil, fmt.Errorf("failed to put item: %v", err)
		}
	}

	return
}

func main() {
	lambda.Start(cfn.LambdaWrap(InsertData))
}
