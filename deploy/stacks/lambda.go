package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type GoLambdaOnDockerExampleStackProps struct {
	awscdk.StackProps
}

func LambdaStack(
	scope constructs.Construct, 
	id string, 
	props *GoLambdaOnDockerExampleStackProps) awscdk.Stack {
		
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	awslambda.NewDockerImageFunction(
		stack,
		jsii.String("GoDockerLambdaStack"),
		&awslambda.DockerImageFunctionProps{
			Description: jsii.String(
				"An example of a Go Lambda function built with Docker"),
			Code: awslambda.DockerImageCode_FromImageAsset(
				jsii.String("../bin"),
				&awslambda.AssetImageCodeProps{},
			),
			Architecture: awslambda.Architecture_ARM_64(),
			Environment: &map[string]*string{
				"CGO_ENABLED": jsii.String("0"),
				"GOOS": jsii.String("linux"),
				"GOARCH": jsii.String("arm64"),
			},
			LogRetention: awslogs.RetentionDays_ONE_DAY,
			MemorySize: jsii.Number(128),
			Timeout: awscdk.Duration_Seconds(jsii.Number(3)),
			Tracing: awslambda.Tracing_ACTIVE,
		},
	)

	return stack
}