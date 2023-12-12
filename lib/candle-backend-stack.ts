import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { DockerImage } from 'aws-cdk-lib';

export class CandleBackendStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create API Gateway
    const api = new apigateway.RestApi(this, 'CandleBackendApi', {
      restApiName: 'CandleBackendApi',
    });

    const roomTable = new cdk.aws_dynamodb.Table(this, 'CandleBackendRoomTable', {
      partitionKey: { name: 'room_id', type: cdk.aws_dynamodb.AttributeType.STRING },
      tableName: 'CandleBackendRoomTable',
    });

    const userTable = new cdk.aws_dynamodb.Table(this, 'CandleBackendUserTable', {
      partitionKey: { name: 'user_id', type: cdk.aws_dynamodb.AttributeType.STRING },
      tableName: 'CandleBackendUserTable',
    });


    // Resolve requests with Lambda
    const room = api.root.addResource('room');

    const roomPOSTHandler = new lambda.Function(this, 'CandleBackendRoomPOSTHandler', {
      functionName: 'RoomPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/POST', {
        bundling: {
          image: DockerImage.fromRegistry("golang:1.21"),
          command: [
            'bash', '-c', [
              "export GOCACHE=/tmp/go-cache",
              "export GOPATH=/tmp/go-path",
              "CGO_ENABLED=0 GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap main.go",
            ].join(" && "),
          ],
        },
      }),
    });
    roomTable.grantReadWriteData(roomPOSTHandler);

    room.addMethod('POST', new apigateway.LambdaIntegration(roomPOSTHandler))
  }
}
