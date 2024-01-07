package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/jtfm/go-lambda-on-docker-example.git/deploy/stacks"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	stacks.DeployStack(
		app,
		"DeployStack",
		&stacks.DeployStackProps{
			StackProps: awscdk.StackProps{
				Env: env(),
			},
		},
	)

	stacks.LambdaStack(
		app,
		"LambdaStack",
		&stacks.GoLambdaOnDockerExampleStackProps{
			StackProps: awscdk.StackProps{
				Env: env(),
			},
		})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
