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
      timeToLiveAttribute: 'TTL',
    });

    const userTable = new cdk.aws_dynamodb.Table(this, 'CandleBackendUserTable', {
      partitionKey: { name: 'user_id', type: cdk.aws_dynamodb.AttributeType.STRING },
      tableName: 'CandleBackendUserTable',
    });


    // Resolve requests with Lambda
    //bundlingの設定を書く必要があった
    //ref:https://iret.media/79515
    const goLambdaBundleConfig ={
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
    }
    const room = api.root.addResource('room');

    //room:POST
    const roomPOSTHandler = new lambda.Function(this, 'CandleBackendRoomPOSTHandler', {
      functionName: 'RoomPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/POST',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomPOSTHandler);
    room.addMethod('POST', new apigateway.LambdaIntegration(roomPOSTHandler))
    
    //room/{room_id}:POST
    const roomId = room.addResource('{room_id}');
    const roomIdPOSTHandler = new lambda.Function(this, 'CandleBackendRoomIdPOSTHandler', {
      functionName: 'RoomIdPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/POST',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdPOSTHandler);
    userTable.grantReadWriteData(roomIdPOSTHandler);
    roomId.addMethod('POST', new apigateway.LambdaIntegration(roomIdPOSTHandler))

    //room/{room_id}/start:GET
    const start = roomId.addResource('start');
    const roomIdStartGETHandler = new lambda.Function(this, 'CandleBackendRoomIdStartGETHandler', {
      functionName: 'RoomIdStartGETHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/start/GET',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdStartGETHandler);
    userTable.grantReadWriteData(roomIdStartGETHandler);
    start.addMethod('GET', new apigateway.LambdaIntegration(roomIdStartGETHandler))
    
    //room/{room_id}/result:GET
    const result = roomId.addResource('result');
    const roomIdResultGETHandler = new lambda.Function(this, 'CandleBackendRoomIdResultGETHandler', {
      functionName: 'RoomIdResultGETHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/result/GET',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdResultGETHandler);
    userTable.grantReadWriteData(roomIdResultGETHandler);
    result.addMethod('GET', new apigateway.LambdaIntegration(roomIdResultGETHandler))

    //room/{room_id}/result:POST
    const roomIdResultPOSTHandler = new lambda.Function(this, 'CandleBackendRoomIdResultPOSTHandler', {
      functionName: 'RoomIdResultPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/result/POST',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdResultPOSTHandler);
    userTable.grantReadWriteData(roomIdResultPOSTHandler);
    result.addMethod('POST', new apigateway.LambdaIntegration(roomIdResultPOSTHandler))

  }
}
