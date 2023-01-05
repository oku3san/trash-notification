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

type Item struct {
  Id     int
  MyText string
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

  // クライアントの設定
  dynamoDbRegion := "ap-northeast-1"
  disableSsl := false

  // DynamoDB Localを利用する場合はEndpointのURLを設定する
  dynamoDbEndpoint := os.Getenv("dynamoDbEndpoint")
  if len(dynamoDbEndpoint) != 0 {
    disableSsl = true
  }

  // デフォルトでは東京リージョンを指定
  if len(dynamoDbRegion) == 0 {
    dynamoDbRegion = "ap-northeast-1"
  }

  db := dynamo.New(session.New(), &aws.Config{
    Region:     aws.String(dynamoDbRegion),
    Endpoint:   aws.String(dynamoDbEndpoint),
    DisableSSL: aws.Bool(disableSsl),
  })

  table := db.Table(os.Getenv("tableName"))

  // 単純なCRUD - Create
  item := Item{
    Id:     0,
    MyText: "My First Text",
  }

  if err := table.Put(item).Run(); err != nil {
    fmt.Printf("Failed to put item[%v]\n", err)
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
