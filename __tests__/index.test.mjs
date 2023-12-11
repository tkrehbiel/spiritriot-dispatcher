import { mockClient } from 'aws-sdk-client-mock';
import 'aws-sdk-client-mock-jest';
import { SQSClient, SendMessageCommand } from '@aws-sdk/client-sqs';
import { handler } from '../index';
import { showvars } from '../index.mjs';

const sqsClient = mockClient(SQSClient);

describe('test suite', () => {
  test('mock index', () => {
    console.log('test started');
  
    // TODO: Why is process.env undefined in the handler???
    process.env.NOTIFIER_QUEUE = 'queue1';
    process.env.WEBMENTION_QUEUE = 'queue2';
    process.env.INCOMING_QUEUE = 'queue3';
  
    const expected = {
      Records: [
        {
          receiptHandle: '1',
          body: {
            url: 'https://anyurl',
            published: '2023-12-09T17:00:00Z',
            detected: '2023-12-10T15:00:00Z',
          },
        },
      ],
    };
  
    console.log(process.env);
    handler(expected, {});
  
    expect(sqsClient).toHaveReceivedCommandWith(SendMessageCommand, {
      QueueUrl: 'queue1',
      MessageBody: JSON.stringify(expected.Records[0].body),
    });
    expect(sqsClient).toHaveReceivedCommandWith(SendMessageCommand, {
      QueueUrl: 'queue2',
      MessageBody: JSON.stringify(expected.Records[0].body),
    });
  });  
});
