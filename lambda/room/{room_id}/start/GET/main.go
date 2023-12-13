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

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("room/{%s}/start/%s\n", event.PathParameters["room_id"],event.HTTPMethod)
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("room/{%s}/start/%s", event.PathParameters["room_id"],event.HTTPMethod),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
