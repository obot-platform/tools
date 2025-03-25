import { client } from '../client';

export async function listUsers() {
  if (!process.env.GUILDID) {
    throw new Error('GUILDID environment variable is required for listUsers');
  }

  const guildId = process.env.GUILDID;
  const guild = await client.guilds.fetch(guildId);
  
  // Fetch all members
  const members = await guild.members.fetch();
  
  const users = members.map(member => ({
    id: member.user.id,
    username: member.user.username,
    discriminator: member.user.discriminator,
    nickname: member.nickname,
    roles: member.roles.cache.map(role => ({
      id: role.id,
      name: role.name,
      color: role.color,
      position: role.position
    })),
    joinedAt: member.joinedTimestamp,
    isBot: member.user.bot,
    avatarUrl: member.user.displayAvatarURL()
  }));

  return JSON.stringify(users);
} 