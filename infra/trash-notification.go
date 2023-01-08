package main

import (
  "github.com/aws/aws-cdk-go/awscdk/v2"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
  "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
  "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
  "github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
  "github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
  "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
  "github.com/aws/constructs-go/constructs/v10"
  "github.com/aws/jsii-runtime-go"
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

  // The code that defines your stack goes here

  // SQS を作成
  trashNotificationQueue := awssqs.NewQueue(stack, jsii.String("trashNotificationQueue"), &awssqs.QueueProps{
    VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
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
  })
  trashNotificationAPiGw.Root().
    AddMethod(jsii.String("POST"), awsapigateway.NewAwsIntegration(&awsapigateway.AwsIntegrationProps{
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
    }), nil).
    AddMethodResponse(&awsapigateway.MethodResponse{
      StatusCode: jsii.String("200"),
      ResponseModels: &map[string]awsapigateway.IModel{
        "application/json": awsapigateway.Model_EMPTY_MODEL(),
      },
      ResponseParameters: nil,
    })

  // Lambda 作成し、DynamoDB の操作権限を付与
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
    },
  })
  trashNotificationTable.GrantReadWriteData(setStatus)

  // SQS をトリガーとするために SQS と Lambda を紐付け
  setStatus.AddEventSource(awslambdaeventsources.NewSqsEventSource(trashNotificationQueue, &awslambdaeventsources.SqsEventSourceProps{
    BatchSize: jsii.Number(1),
    Enabled:   jsii.Bool(true),
  }))

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
  return nil

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
  // return &awscdk.Environment{
  //  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
  //  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
  // }
}
