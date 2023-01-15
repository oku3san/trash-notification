package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "log"
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
    item := make(map[string]types.AttributeValue)

    for k, v := range r.Change.NewImage {
      log.Println("arrtibute info", k, v, v.DataType())
      if v.DataType() == events.DataTypeString {
        item[k] = &types.AttributeValueMemberS{Value: v.String()}
      } else if v.DataType() == events.DataTypeBoolean {
        item[k] = &types.AttributeValueMemberBOOL{Value: v.Boolean()}
      }
    }
  }

  _, err = bot.PushMessage(userId, linebot.NewTextMessage("test")).Do()
  if err != nil {
    fmt.Println(err)
  }

  return nil
}

func main() {
  lambda.Start(handler)
}
