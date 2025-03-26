import { TextChannel, Message, Attachment, Embed, FetchMessagesOptions } from 'discord.js';
import { client } from '../client.js';
import { formatTime, createDataset } from '../utils.js';

export async function getChannelHistoryByTime() {
  if (!process.env.CHANNELID || !process.env.GUILDID || !process.env.LIMIT || !process.env.START || !process.env.END) {
    throw new Error('CHANNELID, GUILDID, LIMIT, START, and END environment variables are required for getChannelHistoryByTime');
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const limit = parseInt(process.env.LIMIT, 10);
  const startTime = new Date(process.env.START).getTime();
  const endTime = new Date(process.env.END).getTime();

  const guild = await client.guilds.fetch(guildId);
  const channel = await guild.channels.fetch(channelId);

  if (!(channel instanceof TextChannel)) {
    throw new Error('Channel is not a text channel');
  }

  // Fetch messages until we have enough or reach the start time
  const messages: Message[] = [];
  let lastId: string | undefined;

  while (messages.length < limit) {
    const options: FetchMessagesOptions = { limit: 100 }; // Fetch 100 at a time for efficiency
    if (lastId) {
      options.before = lastId;
    }

    const batch = await channel.messages.fetch(options);
    if (batch.size === 0) break;

    for (const msg of batch.values()) {
      if (msg.createdTimestamp < startTime) {
        break;
      }
      if (msg.createdTimestamp <= endTime) {
        messages.push(msg);
        if (messages.length >= limit) break;
      }
      lastId = msg.id;
    }

    if (batch.size < 100 || messages.length >= limit) break;
  }

  // Sort messages in chronological order
  messages.sort((a, b) => a.createdTimestamp - b.createdTimestamp);

  const history = await Promise.all(messages.map(async msg => {
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

  await createDataset(history, 'discord_channel_history_by_time');
} 