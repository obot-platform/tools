import { client } from '../client.js';
import { formatTime, createDataset } from '../utils.js';

export async function listUsers() {
  if (!process.env.GUILDID) {
    throw new Error('GUILDID environment variable is required for listUsers');
  }

  const guildId = process.env.GUILDID;
  const guild = await client.guilds.fetch(guildId);
  
  try {
    // Fetch all members
    const members = await guild.members.fetch();
    
    const users = members.map(member => ({
      id: member.user.id,
      name: member.user.username,
      username: member.user.username,
      discriminator: member.user.discriminator,
      nickname: member.nickname,
      roles: member.roles.cache.map(role => ({
        id: role.id,
        name: role.name,
        color: role.color,
        position: role.position
      })),
      joinedAt: formatTime(member.joinedTimestamp),
      isBot: member.user.bot,
      avatarUrl: member.user.displayAvatarURL()
    }));

    await createDataset(users, 'discord_users');
  } catch (error) {
    console.error('Error fetching members:', error);
    throw new Error('Failed to fetch members. Make sure the bot has the GuildMembers intent and proper permissions.');
  }
} 