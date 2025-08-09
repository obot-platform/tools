import { FastMCP, UserError } from 'fastmcp'
import { AtpAgent } from '@atproto/api'
import { tools } from './tools.js'

export function createServer() {
  const server = new FastMCP({
    name: 'Bluesky',
    version: '1.0.0',
  })

  tools.forEach(tool => server.addTool(tool as any))
  return server
}

async function createAgent() {
  const handle = process.env.BLUESKY_HANDLE
  const password = process.env.BLUESKY_APP_PASSWORD

  if (!handle) {
    throw new UserError('BLUESKY_HANDLE environment variable is required')
  }

  if (!password) {
    throw new UserError('BLUESKY_APP_PASSWORD environment variable is required')
  }

  const agent = new AtpAgent({
    service: 'https://bsky.social'
  })

  try {
    await agent.login({
      identifier: handle,
      password: password
    })
  } catch (error: any) {
    throw new UserError(`Bluesky login failed: ${error.message}`)
  }

  return {
    agent
  }
}

// Server startup code: only run if this file is executed directly
if (process.env.NODE_ENV !== 'test' && !process.env.VITEST) {
  global.agent = await createAgent()
  const server = createServer();
  server.start({
    transportType: 'stdio'
  }).then(() => {
    // Vitest has issues capturing stdout, but stderr works.
    console.error('server ready')
  })
}