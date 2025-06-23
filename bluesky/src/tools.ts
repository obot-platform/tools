import { z } from 'zod'
import { createPost, deletePost, searchPosts } from './posts.js'
import { searchUsers } from './users.js'
import { UserError } from 'fastmcp'
import type { AtpAgent } from '@atproto/api'

// Add type declaration for global.agent
// eslint-disable-next-line no-var
declare global {
  // eslint-disable-next-line no-var
  var agent: { agent: AtpAgent, [key: string]: unknown } | undefined;
}

// Define tools
export const tools = [
  {
    name: 'getProfile',
    description: 'Get the profile of the authenticated user',
    parameters: z.object({}),
    execute: async (args: any) => {
      if (!global.agent) {
        throw new UserError('No authenticated session found!')
      }
      try {
        if (!global.agent.agent.session!.handle) {
          throw new UserError('No authenticated handle found')
        }
        const response = await global.agent.agent.getProfile({ actor: global.agent.agent.session!.handle })
        return JSON.stringify(response.data)
      } catch (error: any) {
        if (error instanceof UserError) {
          throw error
        }
        throw new UserError(`Failed to get profile: ${error.message}`)
      }
    }
  },
  {
    name: 'createPost',
    description: 'Create a new post',
    parameters: z.object({
      text: z.string().describe('The text content of the post')
    }),
    execute: async (args: any) => {
      if (!global.agent?.agent.session) {
        throw new UserError('No authenticated session found')
      }
      try {
        const response = await createPost(global.agent.agent, args.text)
        return JSON.stringify(response)
      } catch (error: any) {
        throw new UserError(`Failed to get profile: ${error.message}`)
      }
    }
  },
  {
    name: 'deletePost',
    description: 'Delete a post',
    parameters: z.object({
      postUri: z.string().describe('The URI of the post to delete')
    }),
    execute: async (args: any) => {
      if (!global.agent?.agent.session) {
        throw new UserError('No authenticated session found')
      }
      try {
        await deletePost(global.agent.agent, args.postUri)
        return 'Post deleted'
      } catch (error: any) {
        throw new UserError(`Failed to get profile: ${error.message}`)
      }
    }
  },
  {
    name: 'searchPosts',
    description: 'Search for posts',
    parameters: z.object({
      query: z.string().describe('The search query'),
      since: z.string().optional().describe('Filter posts since this date (ISO 8601)'),
      until: z.string().optional().describe('Filter posts until this date (ISO 8601)'),
      limit: z.string().optional().describe('The maximum number of results to return'),
      tags: z.string().optional().describe('A comma-separated list of tags to search for')
    }),
    execute: async (args: any) => {
      if (!global.agent?.agent.session) {
        throw new UserError('No authenticated session found')
      }
      try {
        const response = await searchPosts(global.agent.agent, args.query, args.since, args.until, args.limit, args.tags)
        return JSON.stringify(response)
      } catch (error: any) {
        throw new UserError(`Failed to get profile: ${error.message}`)
      }
    }
  },
  {
    name: 'searchUsers',
    description: 'Search for users',
    parameters: z.object({
      query: z.string().describe('The search query'),
      limit: z.string().optional().describe('The maximum number of results to return')
    }),
    execute: async (args: any) => {
      if (!global.agent?.agent.session) {
        throw new UserError('No authenticated session found')
      }
      try {
        const response = await searchUsers(global.agent.agent, args.query, args.limit)
        return JSON.stringify(response)
      } catch (error: any) {
        throw new UserError(`Failed to get profile: ${error.message}`)
      }
    }
  }
] 