import { TextChannel } from 'discord.js';
import { client } from '../client.js';
import { createDataset } from '../utils.js';

export async function listChannels() {
  const guilds = client.guilds.cache;
  const channels = [];

  for (const guild of guilds.values()) {
    const guildChannels = await guild.channels.fetch();
    for (const channel of guildChannels.values()) {
      if (channel instanceof TextChannel) {
        channels.push({
          id: channel.id,
          name: channel.name,
          guildId: guild.id,
          guildName: guild.name
        });
      }
    }
  }

  await createDataset(channels, 'discord_channels');
} 