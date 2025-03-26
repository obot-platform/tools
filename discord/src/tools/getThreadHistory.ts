import { TextChannel, Attachment, Embed } from 'discord.js';
import { client } from '../client.js';
import { toRFC3339, createDataset } from '../utils.js';

export async function getThreadHistory() {
  if (!process.env.CHANNELID || !process.env.GUILDID || !process.env.THREADID || !process.env.LIMIT) {
    throw new Error('CHANNELID, GUILDID, THREADID, and LIMIT environment variables are required for getThreadHistory');
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const threadId = process.env.THREADID;
  const limit = parseInt(process.env.LIMIT, 10);

  const guild = await client.guilds.fetch(guildId);
  const channel = await guild.channels.fetch(channelId);

  if (!(channel instanceof TextChannel)) {
    throw new Error('Channel is not a text channel');
  }

  // Get the thread
  const thread = await channel.threads.fetch(threadId);
  if (!thread) {
    throw new Error('Thread not found');
  }

  // Fetch messages from the thread
  const messages = await thread.messages.fetch({ limit });
  const history = messages.map(msg => ({
    id: msg.id,
    content: msg.content,
    permalink: `https://discord.com/channels/${guildId}/${channelId}/${msg.id}`,
    author: {
      id: msg.author.id,
      username: msg.author.username,
      discriminator: msg.author.discriminator,
    },
    timestamp: toRFC3339(msg.createdTimestamp),
    attachments: msg.attachments.map((att: Attachment) => ({
      url: att.url,
      name: att.name,
    })),
    embeds: msg.embeds.map((embed: Embed) => ({
      title: embed.title,
      description: embed.description,
      url: embed.url,
      color: embed.color,
      fields: embed.fields,
      timestamp: embed.timestamp ? toRFC3339(new Date(embed.timestamp).getTime()) : null,
    })),
  }));

  await createDataset(history, 'discord_thread_history');
} 