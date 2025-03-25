import { client } from '../client';

export async function searchUsers() {
  if (!process.env.GUILDID || !process.env.QUERY) {
    throw new Error('GUILDID and QUERY environment variables are required for searchUsers');
  }

  const guildId = process.env.GUILDID;
  const query = process.env.QUERY;
  const guild = await client.guilds.fetch(guildId);
  
  // Fetch members matching the query
  const members = await guild.members.fetch({ query });
  
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