import { TextChannel } from 'discord.js';
import { client } from '../client.js';
import { createDataset } from '../utils.js';

export async function searchChannels() {
  if (!process.env.QUERY) {
    throw new Error('QUERY environment variable is required for searchChannels');
  }

  const query = process.env.QUERY;
  const results = [];

  for (const guild of client.guilds.cache.values()) {
    const guildChannels = await guild.channels.fetch();
    for (const channel of guildChannels.values()) {
      if (channel instanceof TextChannel && 
          channel.name.toLowerCase().includes(query.toLowerCase())) {
        results.push({
          id: channel.id,
          name: channel.name,
          guildId: guild.id,
          guildName: guild.name
        });
      }
    }
  }

  await createDataset(results, 'discord_channel_search');
}
