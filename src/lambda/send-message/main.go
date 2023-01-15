package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
)

type DynamoDbEvent struct {
  DataType struct {
    S string `json:"S"`
  } `json:"DataType"`
  DataValue struct {
    S string `json:"S"`
  } `json:"DataValue"`
  ID struct {
    N string `json:"N"`
  } `json:"Id"`
  UpdatedAt struct {
    N string `json:"N"`
  } `json:"UpdatedAt"`
}

func handler(ctx context.Context, e events.DynamoDBEvent) error {
  bot, err := linebot.New(
    os.Getenv("channelSecret"),
    os.Getenv("accessToken"),
  )
  userId := os.Getenv("userId")
  //var messages []linebot.SendingMessage
  // append some message to messages

  fmt.Println("send-message")

  for _, r := range e.Records {
    if r.Change.NewImage["DataValue"].String() == "True" {
      _, err = bot.PushMessage(userId, linebot.NewTextMessage("yes")).Do()
      if err != nil {
        fmt.Println(err)
      }
    } else {
      _, err = bot.PushMessage(userId, linebot.NewTextMessage("no")).Do()
      if err != nil {
        fmt.Println(err)
      }
    }
  }

  return nil
}

func main() {
  lambda.Start(handler)
}
