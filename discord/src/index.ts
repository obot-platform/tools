import { client } from './client';
import { listChannels } from './tools/listChannels';
import { searchChannels } from './tools/searchChannels';
import { getChannelHistory } from './tools/getChannelHistory';
import { getChannelHistoryByTime } from './tools/getChannelHistoryByTime';
import { getThreadHistory } from './tools/getThreadHistory';
import { searchMessages } from './tools/searchMessages';
import { listUsers } from './tools/listUsers';
import { searchUsers } from './tools/searchUsers';
import { sendMessage } from './tools/sendMessage';
import { sendMessageInThread } from './tools/sendMessageInThread';

async function login(): Promise<void> {
  try {
    await client.login(process.env.DISCORD_TOKEN);
  } catch (error: any) {
    console.log('Discord login failed:', error);
    process.exit(1);
  }
}

async function main() {
  try {
    await login();

    const command = process.argv[2];

    switch (command) {
      case 'listChannels':
        console.log(await listChannels());
        break;
      case 'searchChannels':
        console.log(await searchChannels());
        break;
      case 'getChannelHistory':
        console.log(await getChannelHistory());
        break;
      case 'getChannelHistoryByTime':
        console.log(await getChannelHistoryByTime());
        break;
      case 'getThreadHistory':
        console.log(await getThreadHistory());
        break;
      case 'searchMessages':
        console.log(await searchMessages());
        break;
      case 'listUsers':
        console.log(await listUsers());
        break;
      case 'searchUsers':
        console.log(await searchUsers());
        break;
      case 'sendMessage':
        console.log(await sendMessage());
        break;
      case 'sendMessageInThread':
        console.log(await sendMessageInThread());
        break;
      case 'login':
        console.log('Successfully logged in to Discord'); // login happens earlier in this function
        break;
      default:
        throw new Error(`Unknown command: ${command}`);
    }

    process.exit(0);
  } catch (error) {
    console.error('Error:', error);
    process.exit(1);
  }
}

main(); 