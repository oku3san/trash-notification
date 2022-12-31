package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
)

type SqsMessageFromLine struct {
  Destination string `json:"destination"`
  Events      []struct {
    Type    string `json:"type"`
    Message struct {
      Type string `json:"type"`
      Id   string `json:"id"`
      Text string `json:"text"`
    } `json:"message"`
    WebhookEventId  string `json:"webhookEventId"`
    DeliveryContext struct {
      IsRedelivery bool `json:"isRedelivery"`
    } `json:"deliveryContext"`
    Timestamp int64 `json:"timestamp"`
    Source    struct {
      Type   string `json:"type"`
      UserId string `json:"userId"`
    } `json:"source"`
    ReplyToken string `json:"replyToken"`
    Mode       string `json:"mode"`
  } `json:"events"`
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
  for _, record := range sqsEvent.Records {
    //fmt.Println(message)
    var sqsMessageFromLine SqsMessageFromLine
    if err := json.Unmarshal([]byte(record.Body), &sqsMessageFromLine); err != nil {
      fmt.Println(err)
      return err
    }
    fmt.Printf("%+v\n", sqsMessageFromLine.Events[0].Message.Text)
  }

  return nil
}

func main() {
  lambda.Start(handler)
}
