package stacks

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodecommit"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipelineactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DeployStackProps struct {
	awscdk.StackProps
}

// Creates a stack capable of automated deployment of a lambda function
func DeployStack(
	scope constructs.Construct,
	id string,
	props *DeployStackProps) awscdk.Stack {

	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	repo := awscodecommit.Repository_FromRepositoryName(
		stack,
		jsii.String("Repo"),
		jsii.String("go-lambda-on-docker-example2"),
	)

	createCodePipeline(&stack, repo)

	return stack
}

func createCodePipeline(stack *awscdk.Stack, repo awscodecommit.IRepository) {

	codeBuildRole := createCodeBuildRole(stack)

	// Create a custom log group to attach a retention policy
	logGroupName := jsii.String("CodeBuildLogGroup")
	codeBuildLogGroup := awslogs.NewLogGroup(
		*stack,
		logGroupName,
		&awslogs.LogGroupProps{
			LogGroupName:  logGroupName,
			Retention:     awslogs.RetentionDays_ONE_DAY,
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
	)

	applicationsBucket := awss3.NewBucket(
		*stack,
		jsii.String("ApplicationsBucket"),
		&awss3.BucketProps{
			BucketName:    jsii.String("applications-bucket"),
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
	)

	source := awscodebuild.Source_GitHub(
		&awscodebuild.GitHubSourceProps{
			Owner: jsii.String("jtfm"),
			Repo:  jsii.String("go-lambda-on-docker-example"),
		},
	)

	codeBuildProjectName := jsii.String("CodeBuildProject")
	codebuildProject := awscodebuild.NewProject(
		*stack,
		codeBuildProjectName,
		&awscodebuild.ProjectProps{
			ProjectName: codeBuildProjectName,
			Source:      source,
			Environment: &awscodebuild.BuildEnvironment{
				BuildImage: awscodebuild.LinuxBuildImage_AMAZON_LINUX_2_ARM_2(),
				Privileged: jsii.Bool(true)},
			BuildSpec: awscodebuild.BuildSpec_FromSourceFilename(
				jsii.String("buildspec.yml")),
			Role: codeBuildRole,
			// TODO: Add cache
			Logging: &awscodebuild.LoggingOptions{
				CloudWatch: &awscodebuild.CloudWatchLoggingOptions{
					Enabled:  jsii.Bool(true),
					LogGroup: codeBuildLogGroup,
					Prefix:   jsii.String("aws/codebuild/")}},
			Artifacts: awscodebuild.Artifacts_S3(
				&awscodebuild.S3ArtifactsProps{
					Bucket:         applicationsBucket,
					Path:           repo.RepositoryName(),
					IncludeBuildId: jsii.Bool(true),
					Name:           jsii.String("build.zip"),
					PackageZip:     jsii.Bool(true),
				},
			),
		})

	// deploy to lambda app from s3 bucket
	sourceArtifact := awscodepipeline.NewArtifact(
		jsii.String("SourceArtifact"))

	sourceAction := awscodepipelineactions.NewCodeCommitSourceAction(
		&awscodepipelineactions.CodeCommitSourceActionProps{
			ActionName: jsii.String("Source"),
			Repository: repo,
			Branch:     jsii.String("main"),
			Output:     sourceArtifact,
			Trigger: awscodepipelineactions.CodeCommitTrigger(
				awscodepipelineactions.S3Trigger_EVENTS),
		})

	pipelineName := jsii.String("Pipeline")
	codePipeline := awscodepipeline.NewPipeline(
		*stack,
		pipelineName,
		&awscodepipeline.PipelineProps{
			PipelineName: pipelineName,
			Stages: &[]*awscodepipeline.StageProps{
				{
					StageName: jsii.String("Source"),
					Actions:   &[]awscodepipeline.IAction{sourceAction},
				},
				{
					StageName: jsii.String("Build"),
					Actions: &[]awscodepipeline.IAction{
						awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
							ActionName: jsii.String("Build"),
							Project:    codebuildProject,
							Input:      sourceArtifact,
							Outputs:    &[]awscodepipeline.Artifact{},
						}),
					},
				},
			},
			ArtifactBucket: applicationsBucket,
		},
	)

	detail := map[string]interface{}{
		"referenceType": []interface{}{
			jsii.String("branch"),
		},
	}

	ruleName := jsii.String("MainBranchCommitRule")

	rule := repo.OnCommit(
		ruleName,
		&awscodecommit.OnCommitOptions{
			RuleName: ruleName,
			Branches: jsii.Strings("main"),
			EventPattern: &awsevents.EventPattern{
				Detail: &detail,
			},
		})

	codePipelineTarget := awseventstargets.NewCodePipeline(
		codePipeline,
		&awseventstargets.CodePipelineTargetOptions{
			EventRole: codeBuildRole,
		})

	snsTopicName := jsii.String("CodeCommitSnsTopic")
	snsTopic := awssns.NewTopic(
		*stack,
		snsTopicName,
		&awssns.TopicProps{
			TopicName:   snsTopicName,
			DisplayName: snsTopicName,
		})

	// Send an email to the admin when someone commits to main
	snsTopicTarget := awseventstargets.NewSnsTopic(
		snsTopic,
		&awseventstargets.SnsTopicProps{
			Message: awsevents.RuleTargetInput_FromText(
				jsii.String(fmt.Sprintf(
					"A commit was made to the main branch of the %s repository",
					*repo.RepositoryName())),
			),
		})

	//cdk.SubscribeEmailToSns(stack, "insert email here", snsTopic)

	rule.AddTarget(snsTopicTarget)
	rule.AddTarget(codePipelineTarget)
}

func createCodeBuildRole(stack *awscdk.Stack) awsiam.IRole {
	roleName := jsii.String("CodeBuildRole")
	codeBuildRole := awsiam.NewRole(
		*stack,
		roleName,
		&awsiam.RoleProps{
			RoleName: roleName,
			AssumedBy: awsiam.NewServicePrincipal(
				jsii.String("codebuild.amazonaws.com"),
				nil),
		},
	)

	codeBuildPolicyName := jsii.String("CodeBuildPolicy")
	codeBuildRole.AttachInlinePolicy(
		awsiam.NewPolicy(
			*stack,
			codeBuildPolicyName,
			&awsiam.PolicyProps{
				PolicyName: codeBuildPolicyName,
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Resources: &[]*string{
							jsii.String("arn:aws:iam::*:role/cdk-*"),
						},
						Actions: &[]*string{
							// cloudformation
							jsii.String("sts:AssumeRole"),
							jsii.String("iam:PassRole"),
						},
						Effect: awsiam.Effect_ALLOW,
						Sid:    jsii.String("assumerole"),
					}),
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Resources: &[]*string{
							jsii.String("*"),
						},
						Actions: &[]*string{
							// getparameters
							jsii.String("ssm:GetParameters"),
							jsii.String("ssm:GetParameter"),
						},
						Effect: awsiam.Effect_ALLOW,
					}),
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Resources: &[]*string{
							jsii.String("*"),
						},
						Actions: &[]*string{
							jsii.String("logs:CreateLogStream"),
							jsii.String("logs:CreateLogGroup"),
							jsii.String("logs:PutLogEvents"),
						},
						Effect: awsiam.Effect_ALLOW,
					}),
				},
			},
		),
	)

	return codeBuildRole
}
