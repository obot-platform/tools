import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock AtpAgent and its methods before importing the tools
vi.mock('@atproto/api', () => {
  return {
    AtpAgent: vi.fn().mockImplementation(() => ({
      login: vi.fn().mockResolvedValue(undefined),
      session: { handle: 'test.bsky.social' },
      getProfile: vi.fn().mockResolvedValue({
        data: {
          handle: 'test.bsky.social',
          displayName: 'Test User'
        }
      })
    }))
  }
})

// Mock posts and users modules
vi.mock('../src/posts', () => ({
  createPost: vi.fn().mockResolvedValue({ uri: 'mock:uri', cid: 'mockcid' }),
  deletePost: vi.fn().mockResolvedValue(undefined),
  searchPosts: vi.fn().mockResolvedValue([{ uri: 'mock:uri', text: 'Mock post' }]),
}))
vi.mock('../src/users', () => ({
  searchUsers: vi.fn().mockResolvedValue([{ handle: 'mockuser', displayName: 'Mock User' }]),
}))

import { tools } from '../src/tools.js'
import { UserError } from 'fastmcp'
import * as posts from '../src/posts'
import * as users from '../src/users'

const { createPost, deletePost, searchPosts } = vi.mocked(posts)
const { searchUsers } = vi.mocked(users)

// Set up global.agent with the correct structure
const baseMockAgent = {
  session: { handle: 'test.bsky.social' } as any,
  getProfile: vi.fn().mockResolvedValue({
    data: {
      handle: 'test.bsky.social',
      displayName: 'Test User'
    }
  })
} as any

global.agent = { agent: { ...baseMockAgent } } as any

const getProfileTool = tools.find(t => t.name === 'getProfile')

describe('Bluesky MCP Server Tools', () => {
  beforeEach(() => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
        getProfile: vi.fn().mockResolvedValue({
          data: {
            handle: 'test.bsky.social',
            displayName: 'Test User'
          }
        })
      }
    } as any
  })

  afterEach(() => {
    delete global.agent
  })

  it('should have a getProfile tool', () => {
    expect(getProfileTool).toBeDefined()
  })

  it('should execute getProfile tool successfully', async () => {
    const result = await getProfileTool!.execute({})
    const data = JSON.parse(result as string)
    expect(data.handle).toBe('test.bsky.social')
    expect(data.displayName).toBe('Test User')
  })

  it('should throw a UserError if getProfile fails', async () => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
        getProfile: vi.fn().mockRejectedValue(new Error('Network error'))
      }
    } as any

    await expect(getProfileTool!.execute({}))
      .rejects.toThrow(new UserError('Failed to get profile: Network error'))
  })

  it('should throw a UserError if session is missing', async () => {
    global.agent = undefined // No agent at all

    await expect(getProfileTool!.execute({}))
      .rejects.toThrow(new UserError('No authenticated session found!'))
  })
})

const createPostTool = tools.find(t => t.name === 'createPost')
const deletePostTool = tools.find(t => t.name === 'deletePost')
const searchPostsTool = tools.find(t => t.name === 'searchPosts')
const searchUsersTool = tools.find(t => t.name === 'searchUsers')

describe('Bluesky MCP Server Tools - createPost', () => {
  beforeEach(() => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
      }
    } as any
    createPost.mockReset()
  })
  afterEach(() => { delete global.agent })

  it('should execute createPost tool successfully', async () => {
    const mockResult = { uri: 'post:uri', cid: 'postcid' }
    createPost.mockResolvedValueOnce(mockResult)
    const result = await createPostTool!.execute({ text: 'Hello world' })
    expect(JSON.parse(result)).toEqual(mockResult)
  })

  it('should throw a UserError if session is missing', async () => {
    global.agent = undefined
    await expect(createPostTool!.execute({ text: 'Hello world' }))
      .rejects.toThrow(new UserError('No authenticated session found'))
  })

  it('should throw a UserError if createPost fails', async () => {
    createPost.mockRejectedValueOnce(new Error('Network error'))
    await expect(createPostTool!.execute({ text: 'Hello world' }))
      .rejects.toThrow(new UserError('Failed to get profile: Network error'))
  })
})

describe('Bluesky MCP Server Tools - deletePost', () => {
  beforeEach(() => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
      }
    } as any
    deletePost.mockReset()
  })
  afterEach(() => { delete global.agent })

  it('should execute deletePost tool successfully', async () => {
    deletePost.mockResolvedValueOnce(undefined)
    const result = await deletePostTool!.execute({ postUri: 'post:uri' })
    expect(result).toBe('Post deleted')
  })

  it('should throw a UserError if session is missing', async () => {
    global.agent = undefined
    await expect(deletePostTool!.execute({ postUri: 'post:uri' }))
      .rejects.toThrow(new UserError('No authenticated session found'))
  })

  it('should throw a UserError if deletePost fails', async () => {
    deletePost.mockRejectedValueOnce(new Error('Network error'))
    await expect(deletePostTool!.execute({ postUri: 'post:uri' }))
      .rejects.toThrow(new UserError('Failed to get profile: Network error'))
  })
})

describe('Bluesky MCP Server Tools - searchPosts', () => {
  beforeEach(() => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
      }
    } as any
    searchPosts.mockReset()
  })
  afterEach(() => { delete global.agent })

  it('should execute searchPosts tool successfully', async () => {
    const mockResult = [{ uri: 'post:uri', text: 'Hello' }]
    searchPosts.mockResolvedValueOnce(mockResult)
    const result = await searchPostsTool!.execute({ query: 'Hello' })
    expect(JSON.parse(result)).toEqual(mockResult)
  })

  it('should throw a UserError if session is missing', async () => {
    global.agent = undefined
    await expect(searchPostsTool!.execute({ query: 'Hello' }))
      .rejects.toThrow(new UserError('No authenticated session found'))
  })

  it('should throw a UserError if searchPosts fails', async () => {
    searchPosts.mockRejectedValueOnce(new Error('Network error'))
    await expect(searchPostsTool!.execute({ query: 'Hello' }))
      .rejects.toThrow(new UserError('Failed to get profile: Network error'))
  })
})

describe('Bluesky MCP Server Tools - searchUsers', () => {
  beforeEach(() => {
    global.agent = {
      agent: {
        session: { handle: 'test.bsky.social' } as any,
      }
    } as any
    searchUsers.mockReset()
  })
  afterEach(() => { delete global.agent })

  it('should execute searchUsers tool successfully', async () => {
    const mockResult = [{ handle: 'user1' }]
    searchUsers.mockResolvedValueOnce(mockResult)
    const result = await searchUsersTool!.execute({ query: 'user' })
    expect(JSON.parse(result)).toEqual(mockResult)
  })

  it('should throw a UserError if session is missing', async () => {
    global.agent = undefined
    await expect(searchUsersTool!.execute({ query: 'user' }))
      .rejects.toThrow(new UserError('No authenticated session found'))
  })

  it('should throw a UserError if searchUsers fails', async () => {
    searchUsers.mockRejectedValueOnce(new Error('Network error'))
    await expect(searchUsersTool!.execute({ query: 'user' }))
      .rejects.toThrow(new UserError('Failed to get profile: Network error'))
  })
}) 