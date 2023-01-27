package main

import (
  "github.com/aws/aws-cdk-go/awscdk/v2"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
  "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
  "github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
  "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
  "github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
  "github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
  "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctions"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctionstasks"
  "github.com/aws/constructs-go/constructs/v10"
  "github.com/aws/jsii-runtime-go"
  "os"
)

type TrashNotificationStackProps struct {
  awscdk.StackProps
}

func NewTrashNotificationStack(scope constructs.Construct, id string, props *TrashNotificationStackProps) awscdk.Stack {
  var sprops awscdk.StackProps
  if props != nil {
    sprops = props.StackProps
  }
  stack := awscdk.NewStack(scope, &id, &sprops)

  env := scope.Node().TryGetContext(jsii.String("env")).(string)

  // The code that defines your stack goes here

  // 各種変数を設定
  var accessToken string
  var channelSecret string
  var userId string
  var domainName string
  if env == "local" {
    accessToken = os.Getenv("accessToken")
    channelSecret = os.Getenv("channelSecret")
    userId = os.Getenv("userId")
    domainName = os.Getenv("domainName")
  } else {
    accessToken = *awsssm.StringParameter_ValueFromLookup(stack, jsii.String("line_access_token"))
    channelSecret = *awsssm.StringParameter_ValueFromLookup(stack, jsii.String("line_channel_secret"))
    userId = *awsssm.StringParameter_ValueFromLookup(stack, jsii.String("line_user_id"))
    domainName = *awsssm.StringParameter_ValueFromLookup(stack, jsii.String("domain"))
  }

  // ドメイン名から hosted zone の情報取得
  hostedZone := awsroute53.HostedZone_FromLookup(stack, jsii.String("hostedZone"), &awsroute53.HostedZoneProviderProps{
    DomainName: jsii.String(domainName),
  })

  // ACM 取得
  lineApiAcm := awscertificatemanager.NewCertificate(stack, jsii.String("lineApiAcm"), &awscertificatemanager.CertificateProps{
    DomainName: jsii.String("lineapiv2." + domainName),
    Validation: awscertificatemanager.CertificateValidation_FromDns(hostedZone),
  })

  // SQS を作成
  trashNotificationQueue := awssqs.NewQueue(stack, jsii.String("trashNotificationQueue"), &awssqs.QueueProps{
    VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
    RetentionPeriod:   awscdk.Duration_Seconds(jsii.Number(300)),
  })

  // DynamoDB を作成
  trashNotificationTable := awsdynamodb.NewTable(stack, jsii.String("trashNotificationTable"), &awsdynamodb.TableProps{
    PartitionKey: &awsdynamodb.Attribute{
      Name: jsii.String("Id"),
      Type: awsdynamodb.AttributeType_NUMBER,
    },
    SortKey: &awsdynamodb.Attribute{
      Name: jsii.String("DataType"),
      Type: awsdynamodb.AttributeType_STRING,
    },
    BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
    RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
    Stream:        awsdynamodb.StreamViewType_NEW_AND_OLD_IMAGES,
  })

  // API GW が SQS を呼び出すためのロールを作成
  trashNotificationRole := awsiam.NewRole(stack, jsii.String("trashNotificationRole"), &awsiam.RoleProps{
    AssumedBy: awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), nil),
    ManagedPolicies: &[]awsiam.IManagedPolicy{
      awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSQSFullAccess")),
    },
  })

  // API GW を作成し、SQS と紐付け
  trashNotificationAPiGw := awsapigateway.NewRestApi(stack, jsii.String("trashNotificationAPiGw"), &awsapigateway.RestApiProps{
    DeployOptions: &awsapigateway.StageOptions{
      //DataTraceEnabled: jsii.Bool(true),
      //LoggingLevel:     awsapigateway.MethodLoggingLevel_INFO,
    },
    DomainName: &awsapigateway.DomainNameOptions{
      Certificate: lineApiAcm,
      DomainName:  jsii.String("lineapiv2." + domainName),
    },
  })
  // root パスの設定
  trashNotificationAPiGw.Root().
    AddMethod(
      jsii.String("POST"),
      // SQS インテグレーションの設定
      awsapigateway.NewAwsIntegration(&awsapigateway.AwsIntegrationProps{
        Service:               jsii.String("sqs"),
        IntegrationHttpMethod: jsii.String("POST"),
        Path:                  jsii.String(*awscdk.Stack_Of(stack).Account() + "/" + *trashNotificationQueue.QueueName()),
        Options: &awsapigateway.IntegrationOptions{
          CredentialsRole: trashNotificationRole,
          IntegrationResponses: &[]*awsapigateway.IntegrationResponse{
            {
              StatusCode: jsii.String("200"),
            },
          },
          RequestParameters: &map[string]*string{
            "integration.request.header.Content-Type": jsii.String("'application/x-www-form-urlencoded'"),
          },
          RequestTemplates: &map[string]*string{
            "application/json": jsii.String("Action=SendMessage&MessageBody=$input.body"),
          },
        },
      }),
      // リクエストヘッダの検証
      &awsapigateway.MethodOptions{
        RequestParameters: &map[string]*bool{
          "method.request.header.x-line-signature": jsii.Bool(true),
        },
        RequestValidator: awsapigateway.NewRequestValidator(stack, jsii.String("requestValidator"), &awsapigateway.RequestValidatorProps{
          ValidateRequestParameters: jsii.Bool(true),
          RestApi:                   trashNotificationAPiGw,
        }),
        RequestValidatorOptions: nil,
      }).
    // レスポンスの設定
    AddMethodResponse(&awsapigateway.MethodResponse{
      StatusCode: jsii.String("200"),
      ResponseModels: &map[string]awsapigateway.IModel{
        "application/json": awsapigateway.Model_EMPTY_MODEL(),
      },
      ResponseParameters: nil,
    })

  // localstack の場合、レコード登録時に TTL の部分でエラーになるため
  if env != "local" {
    awsroute53.NewCnameRecord(stack, jsii.String("cnameRecord"), &awsroute53.CnameRecordProps{
      Zone:       hostedZone,
      RecordName: jsii.String("lineapiv2"),
      DomainName: trashNotificationAPiGw.DomainName().DomainNameAliasDomainName(),
    })
  }

  // ステータスを変更するための Lambda 作成し、DynamoDB の操作権限を付与
  setStatus := awslambda.NewFunction(stack, jsii.String("setStatus"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_GO_1_X(),
    Code: awslambda.AssetCode_FromAsset(jsii.String("./../src/lambda/set-status"), &awss3assets.AssetOptions{
      Bundling: &awscdk.BundlingOptions{
        Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
        Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main"),
        User:    jsii.String("root"),
      },
    }),
    Handler: jsii.String("main"),
    Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
    //LogRetention: awslogs.RetentionDays_ONE_DAY,  disabled for local
    Environment: &map[string]*string{
      "tableName": trashNotificationTable.TableName(),
      "env":       jsii.String(env),
    },
  })
  trashNotificationTable.GrantReadWriteData(setStatus)

  // SQS をトリガーとするために SQS と Lambda を紐付け
  setStatus.AddEventSource(awslambdaeventsources.NewSqsEventSource(trashNotificationQueue, &awslambdaeventsources.SqsEventSourceProps{
    BatchSize: jsii.Number(1),
    Enabled:   jsii.Bool(true),
  }))

  // Lambda 作成し、DynamoDB の操作権限を付与
  sendMessage := awslambda.NewFunction(stack, jsii.String("sendMessage"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_GO_1_X(),
    Code: awslambda.AssetCode_FromAsset(jsii.String("./../src/lambda/send-message"), &awss3assets.AssetOptions{
      Bundling: &awscdk.BundlingOptions{
        Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
        Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main"),
        User:    jsii.String("root"),
      },
    }),
    Handler: jsii.String("main"),
    Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
    //LogRetention: awslogs.RetentionDays_ONE_DAY,  disabled for local
    Environment: &map[string]*string{
      "env":           jsii.String(env),
      "accessToken":   jsii.String(accessToken),
      "channelSecret": jsii.String(channelSecret),
      "userId":        jsii.String(userId),
    },
  })
  sendMessage.AddEventSource(awslambdaeventsources.NewDynamoEventSource(trashNotificationTable, &awslambdaeventsources.DynamoEventSourceProps{
    StartingPosition: awslambda.StartingPosition_LATEST,
    Enabled:          jsii.Bool(true),
  }))
  trashNotificationTable.GrantStreamRead(sendMessage)

  // ゴミ捨てステータスを初期化する Lambda 作成し、DynamoDB の操作権限を付与
  resetStatus := awslambda.NewFunction(stack, jsii.String("resetStatus"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_GO_1_X(),
    Code: awslambda.AssetCode_FromAsset(jsii.String("./../src/lambda/reset-status"), &awss3assets.AssetOptions{
      Bundling: &awscdk.BundlingOptions{
        Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
        Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main"),
        User:    jsii.String("root"),
      },
    }),
    Handler: jsii.String("main"),
    Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
    //LogRetention: awslogs.RetentionDays_ONE_DAY,  disabled for local
    Environment: &map[string]*string{
      "tableName": trashNotificationTable.TableName(),
      "env":       jsii.String(env),
    },
  })
  trashNotificationTable.GrantReadWriteData(resetStatus)

  // 毎日0時に Lambda を実行する Event Bridge
  awsevents.NewRule(stack, jsii.String("reset"), &awsevents.RuleProps{
    Enabled: jsii.Bool(true),
    Schedule: awsevents.Schedule_Cron(&awsevents.CronOptions{
      Hour:   jsii.String("15"),
      Minute: jsii.String("0"),
    }),
    Targets: &[]awsevents.IRuleTarget{
      awseventstargets.NewLambdaFunction(resetStatus, &awseventstargets.LambdaFunctionProps{}),
    },
  })

  // ### Step Functions ### //
  getDate := awslambda.NewFunction(stack, jsii.String("getDate"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_GO_1_X(),
    Code: awslambda.AssetCode_FromAsset(jsii.String("./../src/lambda/stepfunctions/get-date"), &awss3assets.AssetOptions{
      Bundling: &awscdk.BundlingOptions{
        Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
        Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main"),
        User:    jsii.String("root"),
      },
    }),
    Handler: jsii.String("main"),
    Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
    //LogRetention: awslogs.RetentionDays_ONE_DAY,  disabled for local
    Environment: &map[string]*string{},
  })

  //sendMessageWithSF :=
  sendMessageWithSF := awslambda.NewFunction(stack, jsii.String("sendMessageWithSF"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_GO_1_X(),
    Code: awslambda.AssetCode_FromAsset(jsii.String("./../src/lambda/stepfunctions/send-message-with-SF"), &awss3assets.AssetOptions{
      Bundling: &awscdk.BundlingOptions{
        Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
        Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main"),
        User:    jsii.String("root"),
      },
    }),
    Handler: jsii.String("main"),
    Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
    //LogRetention: awslogs.RetentionDays_ONE_DAY,  disabled for local
    Environment: &map[string]*string{
      "env":           jsii.String(env),
      "accessToken":   jsii.String(accessToken),
      "channelSecret": jsii.String(channelSecret),
      "userId":        jsii.String(userId),
    },
  })

  choiceIsFinished :=
    awsstepfunctions.NewChoice(stack, jsii.String("Choice - Check IsFinished"), &awsstepfunctions.ChoiceProps{}).
      When(awsstepfunctions.Condition_StringEquals(jsii.String("$.Item.DataValue.S"), jsii.String("False")), awsstepfunctionstasks.NewLambdaInvoke(stack, jsii.String("Lambda Invoke - Send Line message"), &awsstepfunctionstasks.LambdaInvokeProps{
        Timeout:        awscdk.Duration_Seconds(jsii.Number(30)),
        LambdaFunction: sendMessageWithSF,
        OutputPath:     jsii.String("$.Payload"),
      })).
      When(awsstepfunctions.Condition_StringEquals(jsii.String("$.Item.DataValue.S"), jsii.String("True")), awsstepfunctions.NewSucceed(stack, jsii.String("Success"), &awsstepfunctions.SucceedProps{}))

  //dynamoGetItem := awsstepfunctionstasks.NewDynamoGetItem(stack, jsii.String("DynamoDB - GetItem"), &awsstepfunctionstasks.DynamoGetItemProps{
  //  Key: &map[string]awsstepfunctionstasks.DynamoAttributeValue{
  //    "Id":       awsstepfunctionstasks.DynamoAttributeValue_FromNumber(awsstepfunctions.JsonPath_NumberAt(jsii.String("$.dayOfWeekNumber"))),
  //    "DataType": awsstepfunctionstasks.DynamoAttributeValue_FromString(jsii.String("IsFinished")),
  //  },
  //  Table: trashNotificationTable,
  //}).Next(choiceIsFinished)

  choiceCheckDayOfWeekNumber :=
    awsstepfunctions.NewChoice(stack, jsii.String("Choice - Check Day of Week Number"), &awsstepfunctions.ChoiceProps{}).
      When(awsstepfunctions.Condition_StringEquals(jsii.String("$.dayOfWeekNumber"), jsii.String("4")), awsstepfunctionstasks.NewDynamoGetItem(stack, jsii.String("DynamoDB - GetItem"), &awsstepfunctionstasks.DynamoGetItemProps{
        Key: &map[string]awsstepfunctionstasks.DynamoAttributeValue{
          "Id":       awsstepfunctionstasks.DynamoAttributeValue_FromNumber(awsstepfunctions.JsonPath_NumberAt(jsii.String("$.dayOfWeekNumber"))),
          "DataType": awsstepfunctionstasks.DynamoAttributeValue_FromString(jsii.String("IsFinished")),
        },
        Table: trashNotificationTable,
      }).Next(choiceIsFinished)).
      When(awsstepfunctions.Condition_StringEquals(jsii.String("$.dayOfWeekNumber"), jsii.String("5")), awsstepfunctions.NewSucceed(stack, jsii.String("1Success"), &awsstepfunctions.SucceedProps{}))

  init := awsstepfunctionstasks.NewLambdaInvoke(stack, jsii.String("Lambda Invoke - Get Day of Week Number"), &awsstepfunctionstasks.LambdaInvokeProps{
    Timeout:        awscdk.Duration_Seconds(jsii.Number(30)),
    LambdaFunction: getDate,
    OutputPath:     jsii.String("$.Payload"),
  }).Next(choiceCheckDayOfWeekNumber)

  definition := init

  awsstepfunctions.NewStateMachine(stack, jsii.String("stateMachine"), &awsstepfunctions.StateMachineProps{
    Definition:       definition,
    StateMachineType: awsstepfunctions.StateMachineType_STANDARD,
  })

  // Output
  awscdk.NewCfnOutput(stack, jsii.String("dynamoDbName"), &awscdk.CfnOutputProps{
    Value: trashNotificationTable.TableName(),
  })

  awscdk.NewCfnOutput(stack, jsii.String("sqsEndpoint"), &awscdk.CfnOutputProps{
    Value: trashNotificationQueue.QueueUrl(),
  })

  return stack
}

func main() {
  defer jsii.Close()

  app := awscdk.NewApp(nil)

  NewTrashNotificationStack(app, "TrashNotificationStack", &TrashNotificationStackProps{
    awscdk.StackProps{
      Env: env(),
    },
  })

  app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
  // If unspecified, this stack will be "environment-agnostic".
  // Account/Region-dependent features and context lookups will not work, but a
  // single synthesized template can be deployed anywhere.
  //---------------------------------------------------------------------------
  //return nil

  // Uncomment if you know exactly what account and region you want to deploy
  // the stack to. This is the recommendation for production stacks.
  //---------------------------------------------------------------------------
  // return &awscdk.Environment{
  //  Account: jsii.String("123456789012"),
  //  Region:  jsii.String("us-east-1"),
  // }

  // Uncomment to specialize this stack for the AWS Account and Region that are
  // implied by the current CLI configuration. This is recommended for dev
  // stacks.
  //---------------------------------------------------------------------------
  return &awscdk.Environment{
    Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
    Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
  }
}
