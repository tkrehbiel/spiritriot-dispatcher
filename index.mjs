import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

async function process(message) {
  console.log('dispatching work for new post:', JSON.stringify(message));
}

export async function handler(event, context) {
  for (const message of event.Records) {
    await process(message.body);
  }
}
