import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { DockerImage } from 'aws-cdk-lib';
import * as cr from 'aws-cdk-lib/custom-resources';

export class CandleBackendStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create API Gateway
    const api = new apigateway.RestApi(this, 'CandleBackendApi', {
      restApiName: 'CandleBackendApi',
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowHeaders: apigateway.Cors.DEFAULT_HEADERS,
        allowMethods: apigateway.Cors.ALL_METHODS,
      },
    });

    const questionTable = new cdk.aws_dynamodb.Table(this, 'CandleBackendQuestionTable', {
        partitionKey: { name: 'question_id', type: cdk.aws_dynamodb.AttributeType.NUMBER },
        tableName: 'CandleBackendQuestionTable',
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
            "GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap main.go",
          ].join(" && "),
        ],
      },
    }

    const questions = api.root.addResource('questions')

    // questions:GET
    const questionsGETHandler = new lambda.Function(this, 'CandleBackendQuestionsGETHandler', {
        functionName: 'QuestionsGETHandler',
        runtime: lambda.Runtime.PROVIDED_AL2,
        handler: 'bootstrap',
        code: lambda.Code.fromAsset('lambda/questions/GET', goLambdaBundleConfig),
        environment: {
          TABLE_NAME: questionTable.tableName,
        },
    });
    questionTable.grantReadWriteData(questionsGETHandler);
    questions.addMethod('GET', new apigateway.LambdaIntegration(questionsGETHandler));
    // questions:PUT
    const questionsPUTHandler = new lambda.Function(this, 'CandleBackendQuestionsPOSTHandler', {
        functionName: 'QuestionsPUTHandler',
        runtime: lambda.Runtime.PROVIDED_AL2,
        handler: 'bootstrap',
        code: lambda.Code.fromAsset('lambda/questions/PUT', goLambdaBundleConfig),
        environment: {
          TABLE_NAME: questionTable.tableName,
        },
    });
    questionTable.grantReadWriteData(questionsPUTHandler);
    questions.addMethod('PUT', new apigateway.LambdaIntegration(questionsPUTHandler));

    const seedDataLambda = new lambda.Function(this, 'CandleBackendSeedDataLambda', {
        functionName: 'SeedDataLambda',
        runtime: lambda.Runtime.PROVIDED_AL2,
        handler: 'bootstrap',
        code: lambda.Code.fromAsset('lambda/questions/seed', goLambdaBundleConfig),
        environment: {
          TABLE_NAME: questionTable.tableName,
        },
    });
    questionTable.grantWriteData(seedDataLambda)


    const dbInitiateCR = new cr.AwsCustomResource(this, 'dbInitiateCustomResource', {
      onCreate: {
        service:'Lambda',
        action: 'invoke',
        parameters: {
          FunctionName: seedDataLambda.functionArn,
        },
        physicalResourceId: cr.PhysicalResourceId.of('dbInitiateCustomResource'),
      },
      policy: cr.AwsCustomResourcePolicy.fromSdkCalls({resources: cr.AwsCustomResourcePolicy.ANY_RESOURCE}),
    });

    seedDataLambda.grantInvoke(dbInitiateCR);
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

    //room/{room_id}/start:POST
    const start = roomId.addResource('start');
    const roomIdStartPOSTHandler = new lambda.Function(this, 'CandleBackendRoomIdStartPOSTHandler', {
      functionName: 'RoomIdStartPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/start/POST',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdStartPOSTHandler);
    userTable.grantReadWriteData(roomIdStartPOSTHandler);
    questionTable.grantReadWriteData(roomIdStartPOSTHandler);
    start.addMethod('POST', new apigateway.LambdaIntegration(roomIdStartPOSTHandler))

    //room/{room_id}/result:GET
    const result = roomId.addResource('result');
    const roomIdResultGETHandler = new lambda.Function(this, 'CandleBackendRoomIdResultGETHandler', {
      functionName: 'RoomIdResultGETHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/result/{user_id}/GET',goLambdaBundleConfig),
    });
    roomTable.grantReadWriteData(roomIdResultGETHandler);
    userTable.grantReadWriteData(roomIdResultGETHandler);
    result.addResource("{user_id}").addMethod('GET', new apigateway.LambdaIntegration(roomIdResultGETHandler))

    //room/{room_id}/result:POST
    const roomIdResultPOSTHandler = new lambda.Function(this, 'CandleBackendRoomIdResultPOSTHandler', {
      functionName: 'RoomIdResultPOSTHandler',
      runtime: lambda.Runtime.PROVIDED_AL2,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambda/room/{room_id}/result/POST',goLambdaBundleConfig),
      environment: {
        TABLE_NAME: userTable.tableName,
      },
    });
    roomTable.grantReadWriteData(roomIdResultPOSTHandler);
    userTable.grantReadWriteData(roomIdResultPOSTHandler);
    result.addMethod('POST', new apigateway.LambdaIntegration(roomIdResultPOSTHandler))

  }
}
