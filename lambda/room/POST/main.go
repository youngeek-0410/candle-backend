package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Body string `json:"body"`
}

func handler(ctx context.Context,event events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse,error){
	fmt.Println("room POST")
	return events.APIGatewayProxyResponse{
		Body: "room POST",
		StatusCode: 200,
	},nil
}

func main() {
	lambda.Start(handler)
}
