package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/guregu/dynamo"
  "os"
)

type Message struct {
  Type string `json:"type"`
  Id   string `json:"id"`
  Text string `json:"text"`
}

type DeliveryContext struct {
  IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
  Type   string `json:"type"`
  UserId string `json:"userId"`
}

type Events []struct {
  Type            string          `json:"type"`
  Message         Message         `json:"message"`
  WebhookEventId  string          `json:"webhookEventId"`
  DeliveryContext DeliveryContext `json:"deliveryContext"`
  Timestamp       int64           `json:"timestamp"`
  Source          Source          `json:"source"`
  ReplyToken      string          `json:"replyToken"`
  Mode            string          `json:"mode"`
}

type SqsMessageFromLine struct {
  Destination string `json:"destination"`
  Events      Events `json:"events"`
}

type TrashData struct {
  Id        int    `dynamo:"Id"`
  DataType  string `dynamo:"DataType"`
  DataValue any    `dynamo:"DataValue"`
}

//type DataValue struct {
//  DataValue string `dynamo:"DataValue"`
//}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

  sess, err := session.NewSession(&aws.Config{
    Region:   aws.String("ap-northeast-1"),
    Endpoint: aws.String("http://localhost:4566"),
  })
  if err != nil {
    fmt.Println(err)
  }

  db := dynamo.New(sess)
  table := db.Table(os.Getenv("tableName"))
  var results []TrashData
  err = table.Get("Id", 1).All(&results)
  if err != nil {
    fmt.Println(err)
  }

  for i, _ := range results {
    fmt.Println(results[i])
  }

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
