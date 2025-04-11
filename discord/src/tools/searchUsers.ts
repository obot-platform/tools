import { client } from '../client.js';
import { formatTime, createDataset } from '../utils.js';

export async function searchUsers() {
  if (!process.env.QUERY || !process.env.GUILDID) {
    throw new Error('QUERY and GUILDID environment variables are required for searchUsers');
  }

  const query = process.env.QUERY.toLowerCase();
  const guildId = process.env.GUILDID;

  const guild = await client.guilds.fetch(guildId);
  const members = await guild.members.fetch();

  const results = members.filter(member => 
    member.user.username.toLowerCase().includes(query) ||
    (member.nickname && member.nickname.toLowerCase().includes(query))
  ).map(member => ({
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

  await createDataset(results, 'discord_user_search');
} 