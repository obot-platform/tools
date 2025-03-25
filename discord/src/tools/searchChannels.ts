import { TextChannel } from 'discord.js';
import { client } from '../client';

export async function searchChannels() {
  if (!process.env.QUERY) {
    throw new Error('QUERY environment variable is required for searchChannels');
  }

  const query = process.env.QUERY;
  const guilds = client.guilds.cache;
  const results = [];

  for (const guild of guilds.values()) {
    const guildChannels = await guild.channels.fetch();
    for (const channel of guildChannels.values()) {
      if (channel instanceof TextChannel && 
          channel.name.toLowerCase().includes(query.toLowerCase())) {
        results.push({
          id: channel.id,
          name: channel.name,
          guildId: guild.id,
          guildName: guild.name,
        });
      }
    }
  }

  return JSON.stringify(results);
}
