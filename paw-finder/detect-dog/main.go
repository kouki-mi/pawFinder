package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"encoding/json"
)

type LineEvent struct {
	Events []struct {
		Type      string `json:"type"`
		ReplyToken string `json:"replyToken"`
		Message   struct {
			Type string `json:"type"`
			Id string 	`json:"id"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"events"`
}

func validateRequest(request string)(LineEvent, error ){
	var lineEvent LineEvent
	err := json.Unmarshal([]byte(request), &lineEvent)
	if err != nil {
		return LineEvent{}, err
	}
	return lineEvent, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lineEvent, err := validateRequest(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	fmt.Println(lineEvent);

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}