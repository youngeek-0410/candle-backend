package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"os"
	"os/exec"
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

	// Define Lambda function
	lambdaFunction := awslambda.NewFunction(stack, jsii.String("MyLambdaFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Handler: jsii.String("bin/bootstrap"), // Assuming your binary is in the 'bin' directory
	})

	// Define API Gateway
	api := awsapigateway.NewRestApi(stack, jsii.String("MyApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("MyApi"),
		Description: jsii.String("My API Gateway"),
	})

	// Define Lambda integration
	lambdaIntegration := awsapigateway.NewLambdaIntegration(lambdaFunction, &awsapigateway.LambdaIntegrationOptions{})

	// Define API Gateway method and integrate with Lambda
	api.Root().AddMethod("GET", lambdaIntegration, nil)

	// Output the API Gateway URL
	stack.AddOutput(jsii.String("ApiUrl"), &awscdk.OutputProps{
		Value: api.Url(),
	})

	// Build the Go binary
	if err := buildGoBinary(); err != nil {
		fmt.Println("Error building Go binary:", err)
		return nil
	}

	return stack
}

func buildGoBinary() error {
	buildCmd := exec.Command("go", "build", "-o", "bin/bootstrap", "handler/lib.go")
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	return buildCmd.Run()
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
