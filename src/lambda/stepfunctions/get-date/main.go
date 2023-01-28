package main

import (
  "context"
  "github.com/aws/aws-lambda-go/lambda"
  "log"
  "strconv"
  "time"
)

type DayOfWeekNumber struct {
  DayOfWeekNumber string `json:"dayOfWeekNumber"`
}

func handler(ctx context.Context) (DayOfWeekNumber, error) {

  jst, err := time.LoadLocation("Asia/Tokyo")
  if err != nil {
    log.Println(err)
    return DayOfWeekNumber{
      DayOfWeekNumber: "999",
    }, nil
  }
  tomorrow := time.Now().AddDate(0, 0, 1).In(jst).Weekday()
  tomorrowNumberString := strconv.Itoa(int(tomorrow))
  return DayOfWeekNumber{
    DayOfWeekNumber: tomorrowNumberString,
  }, nil
}

func main() {
  lambda.Start(handler)
}
