package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context) (string, error) {
  fmt.Println("test")
  return "test run success", nil
}

func main() {
  lambda.Start(Handler)
}
