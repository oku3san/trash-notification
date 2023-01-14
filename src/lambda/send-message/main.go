package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
)

func handler(ctx context.Context, e events.DynamoDBEvent) error {
  bot, err := linebot.New(
    os.Getenv("channelSecret"),
    os.Getenv("accessToken"),
  )
  userId := os.Getenv("userId")
  //var messages []linebot.SendingMessage
  // append some message to messages

  _, err = bot.PushMessage(userId, linebot.NewTextMessage("test")).Do()
  if err != nil {
    fmt.Println(err)
  }

  fmt.Println("send-message")

  for _, record := range e.Records {
    s, _ := json.Marshal(record.Change.NewImage)
    fmt.Println(string(s))
  }

  return nil
}

func main() {
  lambda.Start(handler)
}
