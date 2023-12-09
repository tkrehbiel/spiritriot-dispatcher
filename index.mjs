import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

async function process(event) {
  console.log('processing', JSON.stringify(event));
}

export async function handler(event, context) {
  console.log('launching from handler');
  for (const message of event.Records) {
    await process(message);
  }
}
