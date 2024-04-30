package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/line/line-bot-sdk-go/linebot"
)

type LineEvents struct {
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

func unmarshalRequest(request string)(LineEvents, error ){
	var lineEvents LineEvents
	err := json.Unmarshal([]byte(request), &lineEvents)
	if err != nil {
		return LineEvents{}, err
	}
	return lineEvents, nil
}

func sendReplytoLine(replyToken string, message string) error {
	channelAccessToken := os.Getenv("CHANNEL_ACCESS_TOKEN")
	channelSecret := os.Getenv("CHANNEL_SECRET")

	bot, err := linebot.New(
		channelSecret,
		channelAccessToken,
	)

    if err != nil {
        return err
    }

	// 応答メッセージを作成
	replyMessage := linebot.NewTextMessage(message)

	// メッセージを送信
	_, err = bot.ReplyMessage(replyToken, replyMessage).Do()
	if err != nil {
		return err
	}
	return nil
}

func validateLineEvents(lineEvent LineEvents) (string, error) {
	for _, event := range lineEvent.Events {
		if event.Type != "message"|| event.Message.Type != "image" {
			err := sendReplytoLine(event.ReplyToken, "ごめんなさい、画像以外は対応していません😢")
			if err != nil {
				return "", err
			}
			continue
		}
	}

	return lineEvent.Events[0].Message.Id, nil
}

func getImageFromLineBot(messageID string) (*bytes.Reader, error){
	url := "https://api-data.line.me/v2/bot/message/"+messageID+"/content"
	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+os.Getenv("CHANNEL_ACCESS_TOKEN"))
    client := new(http.Client)
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

	for key, values := range resp.Header {
        for _, value := range values {
            fmt.Printf("%s: %s\n", key, value)
        }
    }

    bytesResp, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    reader := bytes.NewReader(bytesResp)

	if err != nil {
		return nil, err
	}
	return reader, nil
}

func detectLabel(imageReader *bytes.Reader) (*rekognition.DetectLabelsOutput, error){
	image := make([]byte, imageReader.Len())
	_, err := io.ReadFull(imageReader, image)
	if err != nil {
		return nil, err
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := rekognition.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: []byte(image),
		},
		MaxLabels: aws.Int64(10),
		MinConfidence: aws.Float64(90),
	}

	result, err := svc.DetectLabels(input)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func parseResult(result *rekognition.DetectLabelsOutput) string {
	var message string
	for _, label := range result.Labels {
		fmt.Println(*label.Name, *label.Confidence)
		message += *label.Name
		if(label.Parents != nil){
			message += " ("
			for _, parents := range label.Parents {
				if(parents == nil){
					break
				}
				fmt.Println(*parents.Name)
				message += *parents.Name + ", "
			}
			fmt.Println("----------")
			message += ")\n"
		}
	}
	return message
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lineEvents, err := unmarshalRequest(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	
	messageId, validateErr := validateLineEvents(lineEvents)
	if validateErr != nil {
		sendReplytoLine(lineEvents.Events[0].ReplyToken, "申し訳ありません、サーバー側でエラーが起きています🙇")
		return events.APIGatewayProxyResponse{}, validateErr
	}

	reader, err := getImageFromLineBot(messageId)
	if err != nil {
		sendReplytoLine(lineEvents.Events[0].ReplyToken, "申し訳ありません、サーバー側でエラーが起きています🙇")
		return events.APIGatewayProxyResponse{}, err
	}

	result, err := detectLabel(reader)
	if err != nil {
		sendReplytoLine(lineEvents.Events[0].ReplyToken, "申し訳ありません、サーバー側でエラーが起きています🙇")
		return events.APIGatewayProxyResponse{}, err
	}

	resultMessage := parseResult(result)
	sendReplytoLine(lineEvents.Events[0].ReplyToken, resultMessage)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}