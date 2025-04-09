import {
  Message,
  PublicThreadChannel,
  TextChannel,
  ThreadChannel,
} from "discord.js";
import { client } from "../client.js";
import { formatTime } from "../utils.js";

export async function sendMessageInThread() {
  if (
    !process.env.CHANNELID ||
    !process.env.GUILDID ||
    !process.env.MESSAGEID ||
    !process.env.TEXT
  ) {
    throw new Error(
      "CHANNELID, GUILDID, MESSAGEID, and TEXT environment variables are required for sendMessageInThread"
    );
  }

  const channelId = process.env.CHANNELID;
  const guildId = process.env.GUILDID;
  const text = process.env.TEXT;
  const messageId = process.env.MESSAGEID;
  const guild = await client.guilds.fetch(guildId);
  const channel = await guild.channels.fetch(channelId);

  let message: Message | PublicThreadChannel;
  if (channel instanceof TextChannel) {
    const existingMessage = await channel.messages.fetch(messageId ?? "");
    if (!existingMessage) {
      throw new Error("Could not find specified message");
    }
    const thread = await existingMessage.startThread({
      name: "New Thread",
    });
    message = await thread.send(text);
  } else if (channel instanceof ThreadChannel) {
    message = await channel.send(text);
  } else {
    throw new Error("Channel is not a TextChannel or ThreadChannel");
  }

  if (!message) {
    throw new Error("Could not send message");
  }

  return JSON.stringify({
    id: message.id,
    content: text,
    timestamp: formatTime(message.createdTimestamp),
    permalink: `https://discord.com/channels/${guildId}/${channelId}/${message.id}`,
    channel: {
      id: channel.id,
      name: channel.name,
    },
    guild: {
      id: guild.id,
      name: guild.name,
    },
  });
}
