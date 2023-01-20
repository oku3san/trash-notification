package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/guregu/dynamo"
  "os"
  "time"
)

type TrashData struct {
  Id        int    `dynamo:"Id"`
  DataType  string `dynamo:"DataType"`
  DataValue any    `dynamo:"DataValue"`
  UpdatedAt int    `dynamo:"UpdatedAt,omitempty"`
}

func handler(ctx context.Context) error {

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

  for i := 0; i < 7; i++ {
    t := TrashData{Id: i, DataType: "IsFinished", DataValue: "False", UpdatedAt: int(time.Now().Unix())}
    if err := table.Put(t).Run(); err != nil {
      fmt.Printf("failed to put item[%v]\n", err)
    }
  }

  return nil
}

func main() {
  lambda.Start(handler)
}
