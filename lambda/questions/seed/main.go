package main

import (
    "context"
    "fmt"
    "os"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type QuestionItem struct {
    QuestionID int    `json:"question_id"`
    Statement  string `json:"statement"`
}

func InsertData(ctx context.Context) error {
    sess := session.Must(session.NewSession())
    svc := dynamodb.New(sess)

    tableName := os.Getenv("TABLE_NAME")
    if tableName == "" {
        return fmt.Errorf("TABLE_NAME environment variable is not set")
    }

    items := []QuestionItem{
        {QuestionID: 1, Statement: "『ワンピース』は好きですか？"},
        {QuestionID: 2, Statement: "『進撃の巨人』は好きですか？"},
        {QuestionID: 3, Statement: "『ちいかわ』は好きですか？"},
        {QuestionID: 4, Statement: "『ポケモン』は好きですか？"},
        {QuestionID: 5, Statement: "政権を支持しますか？"},
    }

    for _, item := range items {
        av, err := dynamodbattribute.MarshalMap(item)
        if err != nil {
            return fmt.Errorf("failed to marshal item: %v", err)
        }

        input := &dynamodb.PutItemInput{
            Item:      av,
            TableName: aws.String(tableName),
        }

        _, err = svc.PutItem(input)
        if err != nil {
            return fmt.Errorf("failed to put item: %v", err)
        }
    }

    return nil
}

func main() {
    lambda.Start(InsertData)
}
