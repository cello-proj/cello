import * as sns from '@aws-cdk/aws-sns';
import * as subs from '@aws-cdk/aws-sns-subscriptions';
import * as sqs from '@aws-cdk/aws-sqs';
import * as cdk from '@aws-cdk/core';

export class AppCdkTypescriptStack extends cdk.Stack {
  constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const queue = new sqs.Queue(this, 'AppCdkTypescriptQueue', {
      visibilityTimeout: cdk.Duration.seconds(60)
    });

    const topic = new sns.Topic(this, 'AppCdkTypescriptTopic');

    topic.addSubscription(new subs.SqsSubscription(queue));
  }
}
