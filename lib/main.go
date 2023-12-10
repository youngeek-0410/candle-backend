package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"os"
	"os/exec"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type LibStackProps struct {
	awscdk.StackProps
}

func NewLibStack(scope constructs.Construct, id string, props *LibStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	buildCmd := exec.Command("go", "build", "-tags", " lambda.norpc", "-o", "bin/bootstrap", "handler/main.go")
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	if err := buildCmd.Run(); err != nil {
		return nil
	}

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewLibStack(app, "LibStack", &LibStackProps{
		awscdk.StackProps{
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
