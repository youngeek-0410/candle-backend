import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
export class CandleBackendWebSocketServerStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);
        const vpc = new cdk.aws_ec2.Vpc(this, 'CandleBackendWSSVpc', {
            maxAzs: 2,
        });
        const cluster = new cdk.aws_ecs.Cluster(this, 'CandleBackendWSSCluster', {
            vpc: vpc
        });
        const wsImage = cdk.aws_ecs.ContainerImage.fromAsset('./lib/webSocket');

        const service = new cdk.aws_ecs_patterns.ApplicationLoadBalancedFargateService(this, 'CandleBackendWSService', {
            cluster: cluster,
            cpu: 256,
            desiredCount: 1,
            taskImageOptions: {
                image: wsImage,
                containerPort: 80,
            },
            memoryLimitMiB: 512,
            publicLoadBalancer: true,
        });

        //HTTPヘルスチェック用の設定
        const container = service.taskDefinition.defaultContainer!;
        container.addPortMappings({
            containerPort: 8000,
            hostPort: 8000,
        });
        
        //デフォだとヘルスチェック数分かかってめんどい
        service.targetGroup.configureHealthCheck({
            interval: cdk.Duration.seconds(7),//if HTTP,must be grater than 6 by default
            timeout: cdk.Duration.seconds(5),
            unhealthyThresholdCount: 2,
            healthyThresholdCount: 2,
            path: '/',
            port: '8000',
        });


        new cdk.CfnOutput(this, 'ALBDNSName', {
            value: service.loadBalancer.loadBalancerDnsName,
        });
    }
}