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
  dayOfWeek := time.Now().In(jst).Weekday()
  dayOfWeekNumberString := strconv.Itoa(int(dayOfWeek))
  return DayOfWeekNumber{
    DayOfWeekNumber: dayOfWeekNumberString,
  }, nil
}

func main() {
  lambda.Start(handler)
}
