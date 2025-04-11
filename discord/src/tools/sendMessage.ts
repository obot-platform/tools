import { TextChannel } from 'discord.js';
import { client } from '../client.js';
import { formatTime } from '../utils.js';

export async function sendMessage() {
  if (!process.env.CHANNELID || !process.env.GUILDID || !process.env.TEXT) {
    throw new Error('CHANNELID, GUILDID, and TEXT environment variables are required for sendMessage');
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const text = process.env.TEXT;

  try {
    const guild = await client.guilds.fetch(guildId);
    const channel = await guild.channels.fetch(channelId);

    if (!(channel instanceof TextChannel)) {
      throw new Error('Specified channel is not a text channel');
    }

    const message = await channel.send(text);

    return JSON.stringify({
      id: message.id,
      content: message.content,
      timestamp: formatTime(message.createdTimestamp),
      permalink: `https://discord.com/channels/${guildId}/${channelId}/${message.id}`,
      channel: {
        id: channel.id,
        name: channel.name
      },
      guild: {
        id: guild.id,
        name: guild.name
      }
    });
  } catch (error) {
    console.error('Error sending message:', error);
    throw error;
  }
} 