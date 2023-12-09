import dotenv from 'dotenv';

dotenv.config({ path: './.env.local' });

async function process(event) {
  console.log('processing', JSON.stringify(event));
}

export async function handler(event) {
  console.log('launching from handler');
  await process(event);
}

// Invoke main() if run directly on command line
if (import.meta.url === `file://${process.argv[1]}`) {
  console.log('launching from command line');
  (async () => await process({ message: 'hello world' }))();
}
