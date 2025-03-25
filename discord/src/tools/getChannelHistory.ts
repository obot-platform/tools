import { TextChannel, Attachment, Embed } from 'discord.js';
import { client } from '../client';

export async function getChannelHistory() {
  if (!process.env.CHANNELID || !process.env.GUILDID || !process.env.LIMIT) {
    throw new Error('CHANNELID, GUILDID, and LIMIT environment variables are required for getChannelHistory');
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const limit = parseInt(process.env.LIMIT, 10);

  const guild = await client.guilds.fetch(guildId);
  const channel = await guild.channels.fetch(channelId);

  if (!(channel instanceof TextChannel)) {
    throw new Error('Channel is not a text channel');
  }

  const messages = await channel.messages.fetch({ limit });
  const history = await Promise.all(messages.map(async msg => {
    const result: any = {
      id: msg.id,
      content: msg.content,
      permalink: `https://discord.com/channels/${guildId}/${channelId}/${msg.id}`,
      author: {
        id: msg.author.id,
        username: msg.author.username,
        discriminator: msg.author.discriminator,
      },
      timestamp: msg.createdTimestamp,
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
        timestamp: embed.timestamp,
      })),
    };

    // Check if message is part of a thread
    if (msg.thread) {
      const thread = await channel.threads.fetch(msg.thread.id);
      if (thread) {
        result.thread = {
          id: thread.id,
          name: thread.name,
          messageCount: thread.messageCount,
          memberCount: thread.memberCount,
          isLocked: thread.locked,
          isArchived: thread.archived,
        };
      }
    }

    return result;
  }));

  return JSON.stringify(history);
} 