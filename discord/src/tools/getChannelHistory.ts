import { TextChannel, Attachment, Embed } from 'discord.js';
import { client } from '../client.js';
import { formatTime, createDataset } from '../utils.js';

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
  const history = await Promise.all(Array.from(messages.values())
    .sort((a, b) => a.createdTimestamp - b.createdTimestamp)
    .map(async msg => {
      const result: any = {
        id: msg.id,
        content: msg.content,
        type: msg.type,
        system: msg.system,
        permalink: `https://discord.com/channels/${guildId}/${channelId}/${msg.id}`,
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
      };

      // Add system message details if applicable
      if (msg.system) {
        if (msg.type === 7) { // Member join message
          result.systemData = {
            type: 'memberJoin',
            member: msg.member ? {
              id: msg.member.id,
              username: msg.member.user.username,
              nickname: msg.member.nickname,
              joinedAt: formatTime(msg.member.joinedTimestamp)
            } : null
          };
        }
      }

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

  await createDataset(history, 'discord_channel_history');
} 