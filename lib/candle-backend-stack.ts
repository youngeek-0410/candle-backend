import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as logs from 'aws-cdk-lib/aws-logs';

export class CandleBackendStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create API Gateway
    const api = new apigateway.RestApi(this, 'CandleBackendApi', {
      restApiName: 'CandleBackendApi',
    });

    // Create S3 bucket
    const bucket = new s3.Bucket(this, 'CandleBackendBucket');

    // Create CloudFront distribution
    const distribution = new cloudfront.CloudFrontWebDistribution(this, 'CandleBackendAPIDistribution', {
      originConfigs: [
        {
          customOriginSource: {
            domainName: api.restApiId + '.execute-api.' + this.region + '.amazonaws.com',
            originPath: '/prod',
          },
          behaviors: [{ isDefaultBehavior: true }],
        },
      ],
    });


    // Resolve requests with Lambda
    const user = api.root.addResource('user');
    user.addMethod('GET',new apigateway.LambdaIntegration(
      new lambda.Function(this, 'CandleBackendGETUserHandler', {
        runtime: lambda.Runtime.PROVIDED_AL2,
        code: lambda.Code.fromAsset('lambda/user/{user_id}/GET'),
        handler: 'index.handler',
        logRetention: logs.RetentionDays.ONE_DAY,
      })
    ));
  }
}
