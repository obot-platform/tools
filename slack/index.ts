#!/usr/bin/env node

import express, { Request, Response } from "express";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";
import { WebClient } from "@slack/web-api";
import Fuse from "fuse.js";
import slackifyMarkdown from 'slackify-markdown';

// Schema definitions for all tools
const ListChannelsSchema = z.object({});

const SearchChannelsSchema = z.object({
  query: z.string().describe("The search query for channels"),
});

const GetChannelHistorySchema = z.object({
  channelId: z.string().describe("The ID of the channel to get the history for"),
  limit: z.number().optional().default(10).describe("The number of messages to return"),
});

const GetChannelHistoryByTimeSchema = z.object({
  channelId: z.string().describe("The ID of the channel to get the history for"),
  limit: z.number().optional().default(10).describe("The maximum number of messages to return"),
  start: z.string().describe("The start time in RFC 3339 format"),
  end: z.string().describe("The end time in RFC 3339 format"),
});

const GetThreadHistorySchema = z.object({
  channelId: z.string().describe("The ID of the channel containing the thread"),
  threadId: z.string().describe("The ID of the thread to get the history for"),
  limit: z.number().optional().default(10).describe("The number of messages to return"),
});

const GetThreadHistoryFromLinkSchema = z.object({
  messageLink: z.string().describe("The link to the first Slack message in the thread"),
  limit: z.number().optional().default(10).describe("The number of messages to return"),
});

const SearchMessagesSchema = z.object({
  query: z.string().describe("The search query"),
  sortByTime: z.boolean().optional().default(false).describe("Sort by timestamp rather than score"),
});

const SendMessageSchema = z.object({
  channelId: z.string().describe("The ID of the channel to send the message to"),
  text: z.string().describe("The text to send"),
});

const SendMessageInThreadSchema = z.object({
  channelId: z.string().describe("The ID of the channel containing the thread"),
  threadId: z.string().describe("The ID of the thread to send the message to"),
  text: z.string().describe("The text to send"),
});

const ListUsersSchema = z.object({});

const SearchUsersSchema = z.object({
  query: z.string().describe("The search query for users"),
});

const SendDMSchema = z.object({
  userIds: z.string().describe("Comma-separated list of user IDs to send the message to"),
  text: z.string().describe("The text to send"),
});

const SendDMInThreadSchema = z.object({
  userIds: z.string().describe("Comma-separated list of user IDs for the conversation"),
  threadId: z.string().describe("The ID of the thread to send the message to"),
  text: z.string().describe("The text to send"),
});

const GetMessageLinkSchema = z.object({
  channelId: z.string().describe("The ID of the channel containing the message"),
  messageId: z.string().describe("The ID of the message"),
});

const GetDMHistorySchema = z.object({
  userIds: z.string().describe("Comma-separated list of user IDs for the conversation"),
  limit: z.number().optional().default(10).describe("The number of messages to return"),
});

const GetDMThreadHistorySchema = z.object({
  userIds: z.string().describe("Comma-separated list of user IDs for the conversation"),
  threadId: z.string().describe("The ID of the thread to get the history for"),
  limit: z.number().optional().default(10).describe("The number of messages to return"),
});

const UserContextSchema = z.object({});

const SendTypingEventSchema = z.object({
  channelId: z.string().describe("The ID of the channel to send the typing event to"),
  threadId: z.string().optional().describe("The ID of the thread to send the typing event to"),
  status: z.string().describe("The status to set the typing event that shows in the slack thread"),
});

class SlackClient {
  private webClient: WebClient;

  constructor(botToken: string) {
    this.webClient = new WebClient(botToken);
  }

  async userContext() {
    const result = await this.webClient.auth.test({});
    const userResult = await this.webClient.users.info({ user: result.user_id! });
    return {
      name: userResult.user?.name || "",
      realName: userResult.user?.profile?.real_name || "",
      displayName: userResult.user?.profile?.display_name || "",
      userId: result.user_id || "",
    };
  }

  async listChannels() {
    let allChannels: any[] = [];
    let cursor;
    do {
      const result = await this.webClient.conversations.list({
        limit: 100,
        types: "public_channel,private_channel",
        cursor: cursor,
      });
      if (result.channels) {
        allChannels = allChannels.concat(result.channels);
      }
      cursor = result.response_metadata?.next_cursor;
    } while (cursor);
    return allChannels;
  }

  async searchChannels(query: string) {
    const allChannels = await this.listChannels();
    const fuse = new Fuse(allChannels, {
      keys: ["name"],
      threshold: 0.4,
      findAllMatches: true,
    });
    return fuse.search(query).map((result) => result.item);
  }

  async getChannelHistory(channelId: string, limit: number = 10) {
    const history = await this.webClient.conversations.history({
      channel: channelId,
      limit: limit,
    });
    return history.messages ?? [];
  }

  async getChannelHistoryByTime(channelId: string, limit: number, start: string, end: string) {
    const oldest = new Date(start).getTime() / 1000;
    const latest = new Date(end).getTime() / 1000;
    const history = await this.webClient.conversations.history({
      channel: channelId,
      limit: limit,
      oldest: oldest.toString(),
      latest: latest.toString(),
    });
    return history.messages ?? [];
  }

  async getThreadHistory(channelId: string, threadId: string, limit: number = 10) {
    const replies = await this.webClient.conversations.replies({
      channel: channelId,
      ts: threadId,
      limit: limit,
    });
    return replies.messages ?? [];
  }

  async getThreadHistoryFromLink(messageLink: string, limit: number = 10) {
    const matches = messageLink.match(/archives\/([A-Z0-9]+)\/p(\d+)/);
    if (!matches) {
      throw new Error("Invalid message link format");
    }

    const channelId = matches[1];
    const threadId = matches[2].slice(0, -6) + "." + matches[2].slice(-6);
    return await this.getThreadHistory(channelId, threadId, limit);
  }

  async searchMessages(query: string, sortByTime: boolean = false) {
    const result = await this.webClient.search.all({
      query: query,
      sort: sortByTime ? "timestamp" : "score",
    });
    return result.messages?.matches ?? [];
  }

  removeBoldInHeadings(markdownText: string) {
    // Both headings and bold text will get converted to bold in slack.
    // This function removes the bold markers from headings to help the parser to return the correct text
    // without this , # Welcome to *Markdown* → Slack **Test** --> *Welcome to ​_Markdown_​ → Slack *Test** , which is incorrect.
    return markdownText
      .split('\n')
      .map(line => {
        const headingMatch = line.match(/^(\s{0,3}#+\s)(.*)$/);
        if (headingMatch) {
          const prefix = headingMatch[1]; // "# ", "## ", etc.
          let content = headingMatch[2];
          content = content.replace(/\*\*(.+?)\*\*/g, '$1');
          content = content.replace(/__(.+?)__/g, '$1');
          return prefix + content;
        }
        return line;
      })
      .join('\n');
  }
  
  markdownToSlack(text: string) {
    // Convert markdown format text to slack format text. 
    text = this.removeBoldInHeadings(text)
    return slackifyMarkdown(text)
  }

  async sendMessage(channelId: string, text: string) {
    const result = await this.webClient.chat.postMessage({
      channel: channelId,
      text: this.markdownToSlack(text),
    });
    return result;
  }

  async sendMessageInThread(channelId: string, threadId: string, text: string) {
    const result = await this.webClient.chat.postMessage({
      channel: channelId,
      thread_ts: threadId,
      text: this.markdownToSlack(text),
    });
    return result;
  }

  async listUsers() {
    let allUsers: any[] = [];
    let cursor;
    do {
      const result = await this.webClient.users.list({
        limit: 100,
        cursor: cursor,
      });
      if (result.members) {
        allUsers = allUsers.concat(result.members);
      }
      cursor = result.response_metadata?.next_cursor;
    } while (cursor);
    return allUsers;
  }

  async searchUsers(query: string) {
    const allUsers = await this.listUsers();
    const fuse = new Fuse(allUsers, {
      keys: ["name", "real_name", "profile.display_name"],
      threshold: 0.4,
      findAllMatches: true,
    });
    return fuse.search(query).map((result) => result.item);
  }

  async sendDM(userIds: string, text: string) {
    const userIdArray = userIds.split(",").map((id) => id.trim());
    const result = await this.webClient.conversations.open({
      users: userIdArray.join(","),
    });

    if (!result.ok || !result.channel) {
      throw new Error(`Failed to open DM: ${result.error}`);
    }

    const messageResult = await this.webClient.chat.postMessage({
      channel: result.channel?.id || "",
      text: this.markdownToSlack(text),
    });
    return messageResult;
  }

  async sendDMInThread(userIds: string, threadId: string, text: string) {
    const userIdArray = userIds.split(",").map((id) => id.trim());
    const result = await this.webClient.conversations.open({
      users: userIdArray.join(","),
    });

    if (!result.ok || !result.channel) {
      throw new Error(`Failed to open DM: ${result.error}`);
    }

    const messageResult = await this.webClient.chat.postMessage({
      channel: result.channel?.id || "",
      thread_ts: threadId,
      text: this.markdownToSlack(text),
    });
    return messageResult;
  }

  async getMessageLink(channelId: string, messageId: string) {
    const result = await this.webClient.chat.getPermalink({
      channel: channelId,
      message_ts: messageId,
    });
    return result.permalink;
  }

  async getDMHistory(userIds: string, limit: number = 10) {
    const userIdArray = userIds.split(",").map((id) => id.trim());
    const result = await this.webClient.conversations.open({
      users: userIdArray.join(","),
    });

    if (!result.ok || !result.channel) {
      throw new Error(`Failed to open DM: ${result.error}`);
    }

    const history = await this.webClient.conversations.history({
      channel: result.channel?.id || "",
      limit: limit,
    });
    return history.messages ?? [];
  }

  async getDMThreadHistory(userIds: string, threadId: string, limit: number = 10) {
    const userIdArray = userIds.split(",").map((id) => id.trim());
    const result = await this.webClient.conversations.open({
      users: userIdArray.join(","),
    });

    if (!result.ok || !result.channel) {
      throw new Error(`Failed to open DM: ${result.error}`);
    }

    const replies = await this.webClient.conversations.replies({
      channel: result.channel?.id || "",
      ts: threadId,
      limit: limit,
    });
    return replies.messages ?? [];
  }

  async sendTypingEvent(channelId: string, threadId?: string, status?: string) {
    await this.webClient.assistant.threads.setStatus({
      thread_ts: threadId || "",
      channel_id: channelId,
      status: status || "is typing...",
    })
  }
}

// Create MCP server using the SDK
function createMcpServer(slackClient: SlackClient, token: string): McpServer {
  const server = new McpServer(
    {
      name: "slack-mcp-server",
      version: "1.0.0",
    },
    {
      capabilities: {
        tools: {},
      },
    }
  );

  const isBotToken = token.startsWith("xoxb-");

  server.tool(
    "list_channels",
    "List all channels in the Slack workspace. Returns the name and ID for each channel",
    ListChannelsSchema.shape,
    async (args, { sendNotification }): Promise<CallToolResult> => {
      const channels = await slackClient.listChannels();
      return {
        content: [{ type: "text", text: JSON.stringify(channels, null, 2) }],
      };
    }
  );

  server.tool(
    "search_channels",
    "Search for channels in the Slack workspace",
    SearchChannelsSchema.shape,
    async ({ query }, { sendNotification }): Promise<CallToolResult> => {
      const channels = await slackClient.searchChannels(query);
      return {
        content: [{ type: "text", text: JSON.stringify(channels, null, 2) }],
      };
    }
  );

  server.tool(
    "get_channel_history",
    "Get the chat history for a channel in the Slack workspace",
    GetChannelHistorySchema.shape,
    async ({ channelId, limit }, { sendNotification }): Promise<CallToolResult> => {
      const messages = await slackClient.getChannelHistory(channelId, limit);
      return {
        content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
      };
    }
  );

  server.tool(
    "get_channel_history_by_time",
    "Get the chat history for a channel in the Slack workspace within a specific time range",
    GetChannelHistoryByTimeSchema.shape,
    async ({ channelId, limit, start, end }, { sendNotification }): Promise<CallToolResult> => {
      const messages = await slackClient.getChannelHistoryByTime(channelId, limit, start, end);
      return {
        content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
      };
    }
  );

  server.tool(
    "get_thread_history",
    "Get the chat history for a particular thread",
    GetThreadHistorySchema.shape,
    async ({ channelId, threadId, limit }, { sendNotification }): Promise<CallToolResult> => {
      const messages = await slackClient.getThreadHistory(channelId, threadId, limit);
      return {
        content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
      };
    }
  );

  server.tool(
    "get_thread_history_from_link",
    "Get the chat history for a particular thread from a Slack message link",
    GetThreadHistoryFromLinkSchema.shape,
    async ({ messageLink, limit }, { sendNotification }): Promise<CallToolResult> => {
      const messages = await slackClient.getThreadHistoryFromLink(messageLink, limit);
      return {
        content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
      };
    }
  );

  if (!isBotToken) {
    server.tool(
      "search_messages",
      "Search for messages in the Slack workspace",
      SearchMessagesSchema.shape,
      async ({ query, sortByTime }, { sendNotification }): Promise<CallToolResult> => {
        const messages = await slackClient.searchMessages(query, sortByTime);
        return {
          content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
        };
      }
    );
    server.tool(
      "get_dm_history",
      "Get the chat history for a direct message conversation",
      GetDMHistorySchema.shape,
      async ({ userIds, limit }, { sendNotification }): Promise<CallToolResult> => {
        const messages = await slackClient.getDMHistory(userIds, limit);
        return {
          content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
        };
      }
    );
  
    server.tool(
      "get_dm_thread_history",
      "Get the chat history for a thread in a direct message conversation",
      GetDMThreadHistorySchema.shape,
      async ({ userIds, threadId, limit }, { sendNotification }): Promise<CallToolResult> => {
        const messages = await slackClient.getDMThreadHistory(userIds, threadId, limit);
        return {
          content: [{ type: "text", text: JSON.stringify(messages, null, 2) }],
        };
      }
    );
  }

  server.tool(
    "send_message",
    "Send a message to a channel in the Slack workspace",
    SendMessageSchema.shape,
    async ({ channelId, text }, { sendNotification }): Promise<CallToolResult> => {
      const result = await slackClient.sendMessage(channelId, text);
      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    }
  );

  server.tool(
    "send_message_in_thread",
    "Send a message in a thread in the Slack workspace",
    SendMessageInThreadSchema.shape,
    async ({ channelId, threadId, text }, { sendNotification }): Promise<CallToolResult> => {
      const result = await slackClient.sendMessageInThread(channelId, threadId, text);
      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    }
  );

  server.tool(
    "list_users",
    "List all users in the Slack workspace",
    ListUsersSchema.shape,
    async (args, { sendNotification }): Promise<CallToolResult> => {
      const users = await slackClient.listUsers();
      return {
        content: [{ type: "text", text: JSON.stringify(users, null, 2) }],
      };
    }
  );

  server.tool(
    "search_users",
    "Search for users in the Slack workspace",
    SearchUsersSchema.shape,
    async ({ query }, { sendNotification }): Promise<CallToolResult> => {
      const users = await slackClient.searchUsers(query);
      return {
        content: [{ type: "text", text: JSON.stringify(users, null, 2) }],
      };
    }
  );

  server.tool(
    "send_dm",
    "Send a direct message to a user",
    SendDMSchema.shape,
    async ({ userIds, text }, { sendNotification }): Promise<CallToolResult> => {
      const result = await slackClient.sendDM(userIds, text);
      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    }
  );

  server.tool(
    "send_dm_in_thread",
    "Send a message in a thread in a direct message conversation",
    SendDMInThreadSchema.shape,
    async ({ userIds, threadId, text }, { sendNotification }): Promise<CallToolResult> => {
      const result = await slackClient.sendDMInThread(userIds, threadId, text);
      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    }
  );

  server.tool(
    "get_message_link",
    "Get the permalink for a message",
    GetMessageLinkSchema.shape,
    async ({ channelId, messageId }, { sendNotification }): Promise<CallToolResult> => {
      const link = await slackClient.getMessageLink(channelId, messageId);
      return {
        content: [{ type: "text", text: JSON.stringify({ link }, null, 2) }],
      };
    }
  );

  server.tool(
    "user_context",
    "Get information about the logged in user",
    UserContextSchema.shape,
    async (args, { sendNotification }): Promise<CallToolResult> => {
      const userInfo = await slackClient.userContext();
      return {
        content: [{ type: "text", text: JSON.stringify(userInfo, null, 2) }],
      };
    }
  );

  server.tool(
    "send_typing_event",
    "Send a typing event to a channel in the Slack workspace",
    SendTypingEventSchema.shape,
    async ({ channelId, threadId, status }, { sendNotification }): Promise<CallToolResult> => {
      await slackClient.sendTypingEvent(channelId, threadId, status);
      return {
        content: [{ type: "text", text: "Typing event sent" }],
      };
    }
  );

  return server;
}

async function getServer() {
  try {
    const botToken = process.env.SLACK_BOT_TOKEN;
    if (!botToken) {
      throw new Error("SLACK_BOT_TOKEN is not set");
    }

    const slackClient = new SlackClient(botToken);
    const mcpServer = createMcpServer(slackClient, botToken);
    return mcpServer;
  } catch (error) {
    console.error("Failed to initialize server:", error);
    process.exit(1);
  }
}

const app = express();
app.use(express.json());

app.post("/mcp", async (req: Request, res: Response) => {
  console.log(req.headers);
  const server = await getServer();
  try {
    const transport: StreamableHTTPServerTransport =
      new StreamableHTTPServerTransport({
        sessionIdGenerator: undefined,
      });
    await server.connect(transport);
    await transport.handleRequest(req, res, req.body);
    res.on("close", () => {
      console.log("Request closed");
      transport.close();
      server.close();
    });
  } catch (error) {
    console.error("Error handling MCP request:", error);
    if (!res.headersSent) {
      res.status(500).json({
        jsonrpc: "2.0",
        error: {
          code: -32603,
          message: "Internal server error",
        },
        id: null,
      });
    }
  }
});

app.get("/mcp", async (req: Request, res: Response) => {
  console.log("Received GET MCP request");
  res.writeHead(405).end(
    JSON.stringify({
      jsonrpc: "2.0",
      error: {
        code: -32000,
        message: "Method not allowed.",
      },
      id: null,
    })
  );
});

app.delete("/mcp", async (req: Request, res: Response) => {
  console.log("Received DELETE MCP request");
  res.writeHead(405).end(
    JSON.stringify({
      jsonrpc: "2.0",
      error: {
        code: -32000,
        message: "Method not allowed.",
      },
      id: null,
    })
  );
});

// Start the server
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;
app.listen(PORT, () => {
  console.log(`Slack MCP Stateless Streamable HTTP Server listening on port ${PORT}`);
});

// Handle server shutdown
process.on("SIGINT", async () => {
  console.log("Shutting down server...");
  process.exit(0);
});
