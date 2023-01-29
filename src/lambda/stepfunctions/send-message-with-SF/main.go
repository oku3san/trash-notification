package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
)

func handler(ctx context.Context) error {
  bot, err := linebot.New(
    os.Getenv("channelSecret"),
    os.Getenv("accessToken"),
  )
  if err != nil {
    fmt.Println(err)
  }
  userId := os.Getenv("userId")

  _, err = bot.PushMessage(userId, linebot.NewTemplateMessage(
    "今日のゴミ捨てメッセージ",
    linebot.NewConfirmTemplate(
      "ゴミ捨てしましたか？",
      linebot.NewMessageAction("はい", "はい"),
      linebot.NewMessageAction("いいえ", "いいえ"),
    ),
  )).Do()
  if err != nil {
    fmt.Println(err)
  }
  return nil
}

func main() {
  lambda.Start(handler)
}
