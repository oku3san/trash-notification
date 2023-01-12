package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, e events.DynamoDBEvent) error {

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
