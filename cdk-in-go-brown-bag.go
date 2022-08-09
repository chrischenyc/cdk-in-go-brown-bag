package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v10"
)

type CdkInGoBrownBagStackProps struct {
	awscdk.StackProps
}

func NewCdkInGoBrownBagStack(scope constructs.Construct, id string, props *CdkInGoBrownBagStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// 02 - s3 bucket
	bucket := awss3.NewBucket(stack, jsii.String("bucket"), &awss3.BucketProps{
		BucketName:        stack.StackName(),
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EnforceSSL:        jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	// 06 - DynamoDB table
	table := awsdynamodb.NewTable(stack, jsii.String("table"), &awsdynamodb.TableProps{
		TableName:   stack.StackName(),
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("pk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("sk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// 03 - lambda function
	lambda := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("lambda"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String(fmt.Sprintf("%v-s3-event-handler", *stack.StackName())),
		Entry:        jsii.String("./src/s3-event-handler"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Environment: &map[string]*string{
			"TABLE_NAME": table.TableName(),
		},
	})

	// 04 - trigger lambda function by S3 event
	lambda.AddEventSource(awslambdaeventsources.NewS3EventSource(bucket, &awslambdaeventsources.S3EventSourceProps{
		Events: &[]awss3.EventType{
			awss3.EventType_OBJECT_CREATED_PUT,
		},
	}))

	// 05 - grant lambda function permissions to access resources
	bucket.GrantRead(lambda, "*") // permission not required by this use case. demonstration purpose
	table.GrantReadWriteData(lambda)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	// inject env vars: dev, staging, prod
	NewCdkInGoBrownBagStack(app, "CdkInGoBrownBagStack", &CdkInGoBrownBagStackProps{
		awscdk.StackProps{
			StackName: jsii.String("cdk-demo-chris"),
			Env:       env(),
			Tags: &map[string]*string{
				"Environment": jsii.String("dev"),
			},
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
