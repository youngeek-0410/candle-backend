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

func InsertData(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return event.PhysicalResourceID, nil, fmt.Errorf("TABLE_NAME environment variable is not set")
	}

	items := []QuestionItem{
		{QuestionID: 1, Statement: "料理をすることは好きですか？"},
		{QuestionID: 2, Statement: "読書は好きですか？"},
		{QuestionID: 3, Statement: "映画鑑賞は好きですか？"},
		{QuestionID: 4, Statement: "『ポケモン』は好きですか？"},
		{QuestionID: 5, Statement: "ジョギングは好きですか？"},
		{QuestionID: 6, Statement: "カラオケは好きですか？"},
		{QuestionID: 7, Statement: "コンサートに行くことは好きですか？"},
		{QuestionID: 8, Statement: "ボードゲームは好きですか？"},
		{QuestionID: 9, Statement: "アニメを見ることは好きですか？"},
		{QuestionID: 10, Statement: "筋トレは好きですか？"},
		{QuestionID: 11, Statement: "手芸や工作は好きですか？"},
		{QuestionID: 12, Statement: "キャンプは好きですか？"},
		{QuestionID: 13, Statement: "海外旅行は好きですか？"},
		{QuestionID: 14, Statement: "幼少期にスポーツチームに所属していましたか？"},
		{QuestionID: 15, Statement: "歴史の授業が好きでしたか？"},
		{QuestionID: 16, Statement: "学校の科学の授業が得意でしたか？"},
		{QuestionID: 17, Statement: "スパイシーな食べ物が好きですか？"},
		{QuestionID: 18, Statement: "過去に100冊以上の本を読んだことがありますか？"},
		{QuestionID: 19, Statement: "ペットを飼ったことがありますか？"},
		{QuestionID: 20, Statement: "世の中の動向を追っていますか"},
	}

	for _, item := range items {
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return event.PhysicalResourceID, nil, fmt.Errorf("failed to marshal item: %v", err)
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			return event.PhysicalResourceID, nil, fmt.Errorf("failed to put item: %v", err)
		}
	}

	return event.PhysicalResourceID, nil, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(InsertData))
}
