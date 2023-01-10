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
  "time"
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

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

  // 曜日の番号を取得
  dayOfWeekNumber := getDayOfWeek()

  // DynamoDB 接続の初期設定
  var endpoint string
  if os.Getenv("env") == "local" {
    endpoint = "http://localhost:4566"
  }
  sess, err := session.NewSession(&aws.Config{
    Region:   aws.String("ap-northeast-1"),
    Endpoint: aws.String(endpoint),
  })
  if err != nil {
    fmt.Printf("failed new session [%v]", err)
  }
  db := dynamo.New(sess)
  table := db.Table(os.Getenv("tableName"))

  // SQS のメッセージ取得
  for _, record := range sqsEvent.Records {
    var sqsMessageFromLine SqsMessageFromLine
    if err := json.Unmarshal([]byte(record.Body), &sqsMessageFromLine); err != nil {
      fmt.Printf("failed unmarshal json %v\n", err)
      return err
    }
    message := sqsMessageFromLine.Events[0].Message.Text

    if message == "はい" {
      t := TrashData{Id: dayOfWeekNumber, DataType: "IsFinished", DataValue: "True"}
      if err := table.Put(t).Run(); err != nil {
        fmt.Printf("failed to put item[%v]\n", err)
      }
    } else {
      t := TrashData{Id: dayOfWeekNumber, DataType: "IsFinished", DataValue: "False"}
      if err := table.Put(t).Run(); err != nil {
        fmt.Printf("failed to put item[%v]\n", err)
      }
    }
  }

  return nil
}

func main() {
  lambda.Start(handler)
}

func getDayOfWeek() int {
  jst, err := time.LoadLocation("Asia/Tokyo")
  if err != nil {
    fmt.Println(err)
  }
  dayOfWeek := time.Now().In(jst).Weekday()
  dayOfWeekNumber := int(dayOfWeek)

  return dayOfWeekNumber
}
