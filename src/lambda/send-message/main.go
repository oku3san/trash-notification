package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
  "strconv"
  "time"
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
  var messages []linebot.SendingMessage
  for _, r := range e.Records {

    updatedAtString := r.Change.NewImage["UpdatedAt"].Number()
    currentHour := getHourInJST(updatedAtString)

    if currentHour == 0 {
      return nil
    } else {
      if r.Change.NewImage["DataValue"].String() == "True" {
        messages = append(messages, linebot.NewStickerMessage("6370", "11088025"))
      } else {
        messages = append(messages, linebot.NewStickerMessage("8515", "16581257"))
      }
      _, err = bot.PushMessage(userId, messages...).Do()
      if err != nil {
        fmt.Println(err)
      }
    }
  }
  return nil
}

func getHourInJST(unixTime string) int {
  timezone, _ := time.LoadLocation("Asia/Tokyo")
  updatedAtInt64, _ := strconv.ParseInt(unixTime, 10, 64)
  date := time.Unix(updatedAtInt64, 0).In(timezone)
  return date.Hour()
}

func main() {
  lambda.Start(handler)
}
