import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
export class CandleBackendWebSocketServerStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);
        const vpc = new cdk.aws_ec2.Vpc(this, 'CandleBackendWSSVpc', {
            maxAzs: 2,
        });
        const cluster = new cdk.aws_ecs.Cluster(this, 'CandleBackendWSSCluster', {
            vpc: vpc,
            capacity: {
                instanceType: new cdk.aws_ec2.InstanceType('t2.nano'),
            },
        });

        const task = new cdk.aws_ecs.FargateTaskDefinition(this, 'CandleBackendWSSContainer', {
            memoryLimitMiB: 512,
            cpu: 256,
        });
        const image = cdk.aws_ecs.ContainerImage.fromAsset('./lib/webSocket');

        const container = task.addContainer('CandleBackendWSSContainer', {
            image: image,
        });


        container.addPortMappings({
            containerPort: 80,
            protocol: cdk.aws_ecs.Protocol.TCP,
        });

        const securityGroup = new cdk.aws_ec2.SecurityGroup(this, 'CandleBackendWSSSecurityGroup', {
            vpc: vpc,
            allowAllOutbound: true,
        });

        securityGroup.addIngressRule(cdk.aws_ec2.Peer.anyIpv4(), cdk.aws_ec2.Port.tcp(80));
        const service = new cdk.aws_ecs.FargateService(this, 'CandleBackendWSSService', {
            cluster: cluster,
            taskDefinition: task,
            securityGroups: [securityGroup],
            assignPublicIp: true,
        });
    }
}