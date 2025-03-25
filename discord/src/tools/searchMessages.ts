import { TextChannel, Attachment, Embed, Message, Collection } from 'discord.js';
import { client } from '../client';

export async function searchMessages() {
  if (!process.env.QUERY || !process.env.LIMIT) {
    throw new Error('QUERY and LIMIT environment variables are required for searchMessages');
  }

  const query = process.env.QUERY;
  const limit = parseInt(process.env.LIMIT, 10);
  const guildId = process.env.GUILDID; // Optional
  const channelId = process.env.CHANNELID; // Optional

  // If channelId is provided, guildId must also be provided
  if (channelId && !guildId) {
    throw new Error('GUILDID must be provided when CHANNELID is specified');
  }

  const results = [];

  if (guildId && channelId) {
    // Search in specific channel
    const guild = await client.guilds.fetch(guildId);
    const channel = await guild.channels.fetch(channelId);
    
    if (!(channel instanceof TextChannel)) {
      throw new Error('Specified channel is not a text channel');
    }

    try {
      const messages = await channel.messages.fetch({ 
        limit: Math.min(100, limit)
      });

      const matchingMessages = messages.filter(msg => 
        msg.content.toLowerCase().includes(query.toLowerCase())
      );

      for (const msg of matchingMessages.values()) {
        const result: any = {
          id: msg.id,
          content: msg.content,
          permalink: `https://discord.com/channels/${guild.id}/${channel.id}/${msg.id}`,
          author: {
            id: msg.author.id,
            username: msg.author.username,
            discriminator: msg.author.discriminator,
          },
          timestamp: msg.createdTimestamp,
          channel: {
            id: channel.id,
            name: channel.name,
          },
          guild: {
            id: guild.id,
            name: guild.name,
          },
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

        results.push(result);
        if (results.length >= limit) break;
      }
    } catch (error) {
      console.error(`Error searching messages in channel ${channel.name}:`, error);
    }
  } else {
    // Search across all channels in specified guild(s)
    const guilds = guildId ? [await client.guilds.fetch(guildId)] : client.guilds.cache.values();

    for (const guild of guilds) {
      const channels = await guild.channels.fetch();
      for (const channel of channels.values()) {
        if (!(channel instanceof TextChannel)) continue;

        try {
          const messages = await channel.messages.fetch({ 
            limit: Math.min(100, limit)
          });

          const matchingMessages = messages.filter(msg => 
            msg.content.toLowerCase().includes(query.toLowerCase())
          );

          for (const msg of matchingMessages.values()) {
            const result: any = {
              id: msg.id,
              content: msg.content,
              permalink: `https://discord.com/channels/${guild.id}/${channel.id}/${msg.id}`,
              author: {
                id: msg.author.id,
                username: msg.author.username,
                discriminator: msg.author.discriminator,
              },
              timestamp: msg.createdTimestamp,
              channel: {
                id: channel.id,
                name: channel.name,
              },
              guild: {
                id: guild.id,
                name: guild.name,
              },
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

            results.push(result);
            if (results.length >= limit) break;
          }
        } catch (error) {
          console.error(`Error searching messages in channel ${channel.name}:`, error);
          continue;
        }
        
        if (results.length >= limit) break;
      }
      if (results.length >= limit) break;
    }
  }

  return JSON.stringify(results.slice(0, limit));
} 