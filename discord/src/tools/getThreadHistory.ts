import { TextChannel, Attachment, Embed } from 'discord.js';
import { client } from '../client.js';
import { formatTime, createDataset } from '../utils.js';

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
  const history = Array.from(messages.values())
    .sort((a, b) => a.createdTimestamp - b.createdTimestamp)
    .map(msg => ({
      id: msg.id,
      content: msg.content,
      type: msg.type,
      system: msg.system,
      permalink: `https://discord.com/channels/${guildId}/${channelId}/${threadId}/${msg.id}`,
      author: {
        id: msg.author.id,
        username: msg.author.username,
        discriminator: msg.author.discriminator,
      },
      timestamp: formatTime(msg.createdTimestamp),
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
        timestamp: embed.timestamp ? formatTime(new Date(embed.timestamp).getTime()) : null,
      })),
      systemData: msg.system && msg.type === 7 ? {
        type: 'memberJoin',
        member: msg.member ? {
          id: msg.member.id,
          username: msg.member.user.username,
          nickname: msg.member.nickname,
          joinedAt: formatTime(msg.member.joinedTimestamp)
        } : null
      } : undefined
    }));

  await createDataset(history, 'discord_thread_history');
} 