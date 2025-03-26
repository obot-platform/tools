import { TextChannel, ThreadChannel } from 'discord.js';
import { client } from '../client.js';
import { formatTime } from '../utils.js';

export async function sendMessageInThread() {
  if (!process.env.CHANNELID || !process.env.GUILDID || !process.env.THREADID || !process.env.TEXT) {
    throw new Error('CHANNELID, GUILDID, THREADID, and TEXT environment variables are required for sendMessageInThread');
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const threadId = process.env.THREADID;
  const text = process.env.TEXT;

  try {
    const guild = await client.guilds.fetch(guildId);
    const channel = await guild.channels.fetch(channelId);

    if (!(channel instanceof TextChannel)) {
      throw new Error('Specified channel is not a text channel');
    }

    const thread = await channel.threads.fetch(threadId);
    if (!thread) {
      throw new Error('Thread not found');
    }

    const message = await thread.send(text);

    return JSON.stringify({
      id: message.id,
      content: message.content,
      timestamp: formatTime(message.createdTimestamp),
      permalink: `https://discord.com/channels/${guildId}/${channelId}/${message.id}`,
      channel: {
        id: channel.id,
        name: channel.name
      },
      thread: {
        id: thread.id,
        name: thread.name
      },
      guild: {
        id: guild.id,
        name: guild.name
      }
    });
  } catch (error) {
    console.error('Error sending message in thread:', error);
    throw error;
  }
} 