import { SQSClient, SendMessageCommand } from '@aws-sdk/client-sqs';
import { fromNodeProviderChain } from '@aws-sdk/credential-providers';
import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

const sqsClient = new SQSClient({
  credentials: fromNodeProviderChain(),
});

async function sendMessage(queueName, message) {
  if (!queueName) return;
  const command = new SendMessageCommand({
    QueueUrl: queueName,
    MessageBody: JSON.stringify(message),
  });
  return sqsClient.send(command);
}

async function process(message) {
  console.log('dispatching work for new post:', JSON.stringify(message));
  const promises = [];
  promises.push(sendMessage(process.env.NOTIFIER_QUEUE, message));
  promises.push(sendMessage(process.env.WEBMENTION_QUEUE, message));
  await Promise.all(promises);
}

export async function handler(event, context) {
  for (const message of event.Records) {
    await process(message.body);
  }
}
