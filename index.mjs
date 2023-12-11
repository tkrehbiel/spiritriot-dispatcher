import {
  SQSClient,
  SendMessageCommand,
  DeleteMessageCommand,
} from '@aws-sdk/client-sqs';
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

async function deleteMessage(queueName, receipt) {
  if (!queueName) return;
  const command = new DeleteMessageCommand({
    QueueUrl: queueName,
    ReceiptHandle: receipt,
  });
  return sqsClient.send(command);
}

async function process(message) {
  console.log('dispatching work for new post:', JSON.stringify(message.body));
  console.log(process.env);
  const promises = [];
  try {
    promises.push(sendMessage(process.env.NOTIFIER_QUEUE, message.body));
    promises.push(sendMessage(process.env.WEBMENTION_QUEUE, message.body));
    await Promise.all(promises).then((x) =>
      deleteMessage(process.env.INCOMING_QUEUE, message.receiptHandle),
    );
  } catch (error) {
    console.log(error);
  }
}

export async function handler(event, context) {
  console.log(process.env);
  const promises = [];
  for (const message of event.Records) {
    promises.push(process(message));
  }
  await Promise.all(promises);
}
