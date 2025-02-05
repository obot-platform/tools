import fs from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import { type IncomingHttpHeaders } from 'node:http'
import { createHash } from 'node:crypto'
import { type BrowserContext } from 'playwright'
import { newBrowserContext } from './context.ts'
import TTLCache from '@isaacs/ttlcache'
import AsyncLock from 'async-lock'

const APP_CACHE_DIR = (() => {
  const homeDir = os.homedir()
  const appPath = path.join('obot', 'tools', 'google', 'search')

  switch (os.platform()) {
    case 'win32':
      return path.join(process.env.APPDATA ?? path.join(homeDir, 'AppData', 'Roaming'), appPath)
    case 'darwin':
      return path.join(homeDir, 'Library', 'Caches', appPath)
    default:
      return path.join(process.env.XDG_CACHE_HOME ?? path.join(homeDir, '.cache'), appPath)
  }
})()

async function clearAppCacheDir (): Promise<void> {
  try {
    await fs.rm(APP_CACHE_DIR, { recursive: true, force: true })
    console.log(`Cleared APP_CACHE_DIR at startup: ${APP_CACHE_DIR}`)
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    console.error(`Failed to clear APP_CACHE_DIR: ${msg}`)
  }
}

// Call the function at startup
await clearAppCacheDir()

const SESSION_TTL = 5 * 60 * 1000 // 5 minutes
const LOCK_TIMEOUT = 10000 // 10 seconds timeout for locks

interface Session {
  sessionId: string
  browserContext: BrowserContext
  active: number
}

export class SessionManager {
  private static instance: SessionManager | null = null
  private sessionCache: TTLCache<string, Session>
  private lock: AsyncLock
  private isDestroyed = false

  private constructor() {
    this.sessionCache = new TTLCache<string, Session>({
      ttl: SESSION_TTL,
      max: 10000,
      updateAgeOnGet: true,
      noDisposeOnSet: true,
      dispose: async (session: Session, sessionId: string, reason: string) => {
        console.info(`[Session ${sessionId}] Disposing session due to ${reason}.`)
        await this.destroySession(session)
      },
    })

    this.lock = new AsyncLock({ timeout: LOCK_TIMEOUT })
  }

  public static create(): SessionManager {
    if (!this.instance) {
      this.instance = new SessionManager()
    }
    return this.instance
  }

  private async acquireSession(sessionId: string): Promise<Session> {
    return await this.lock.acquire(sessionId, async () => {
      if (this.isDestroyed) throw new Error('SessionManager is destroyed.')

      let session = this.sessionCache.get(sessionId)
      if (!session) {
        console.info(`[SessionManager] Creating new session for ${sessionId}.`)
        session = await this.createSession(sessionId)
        this.sessionCache.set(sessionId, session)
      }

      session.active++
      console.info(`[Session ${session.sessionId}] Acquired active=${session.active}.`)
      return session
    })
  }

  private async releaseSession(sessionId: string): Promise<void> {
    await this.lock.acquire(sessionId, async () => {
      const session = this.sessionCache.get(sessionId)
      if (!session) {
        console.warn(`[Session ${sessionId}] Release attempted but session does not exist.`)
        return
      }

      session.active--
      console.info(`[Session ${session.sessionId}] Released active=${session.active}.`)
    })
  }

  private async createSession(sessionId: string): Promise<Session> {
    const sessionDir = path.resolve(APP_CACHE_DIR, 'browser_sessions', sessionId)
    await fs.mkdir(sessionDir, { recursive: true })
    const browserContext = await newBrowserContext(sessionDir)
    console.info(`[Session ${sessionId}] Created session (dir: ${sessionDir}).`)
    return { sessionId, browserContext, active: 0 }
  }

  private async destroySession(session: Session, force: boolean = false): Promise<void> {
    await this.lock.acquire(session.sessionId, async () => {
      if (force) {
        console.warn(`[Session ${session.sessionId}] Force destroying session.`)
      } else if (session.active > 0) {
        console.warn(`[Session ${session.sessionId}] Dispose attempted while active=${session.active}.`)
        return
      }

      console.info(`[Session ${session.sessionId}] Finalizing session.`)
      try {
        await session.browserContext.close()
        console.info(`[Session ${session.sessionId}] Browser context closed.`)
      } catch (err) {
        console.error(`Error closing browser context for session ${session.sessionId}: ${err}`)
      }

      try {
        const sessionDir = path.resolve(APP_CACHE_DIR, 'browser_sessions', session.sessionId)
        await fs.rm(sessionDir, { recursive: true, force: true })
        console.info(`[Session ${session.sessionId}] Session directory removed.`)
      } catch (err) {
        console.error(`Error removing session directory for session ${session.sessionId}: ${err}`)
      }
    })
  }

  public async withSession<T>(sessionId: string, fn: (ctx: BrowserContext) => Promise<T>): Promise<T> {
    const session = await this.acquireSession(sessionId)
    try {
      console.info(`[SessionManager] Running work on session ${session.sessionId}.`)
      return await fn(session.browserContext)
    } catch (err) {
      console.error(`Error during work on session ${session.sessionId}: ${err}`)
      throw err
    } finally {
      await this.releaseSession(sessionId)
    }
  }

  public async destroy(): Promise<void> {
    console.info('[SessionManager] Destroying all sessions...')
    this.isDestroyed = true

    const sessionIds = Array.from(this.sessionCache.keys())
    for (const sessionId of sessionIds) {
      await this.lock.acquire(sessionId, async () => {
        const session = this.sessionCache.get(sessionId)
        if (session) {
          console.info(`[Session ${sessionId}] Closing session as part of destroy().`)
          await this.destroySession(session, true)
        }
      })
    }

    this.sessionCache.clear()
    SessionManager.instance = null
    console.info('[SessionManager] Destroy complete.')
  }
}

export function getSessionId(headers: IncomingHttpHeaders): string {
  const workspaceId = getGPTScriptEnv(headers, 'GPTSCRIPT_WORKSPACE_ID')
  if (workspaceId == null) throw new Error('No GPTScript workspace ID provided')

  return createHash('sha256').update(workspaceId).digest('hex').substring(0, 16)
}

export function getWorkspaceId(headers: IncomingHttpHeaders): string | undefined {
  return getGPTScriptEnv(headers, 'GPTSCRIPT_WORKSPACE_ID')
}

export function getGPTScriptEnv(headers: IncomingHttpHeaders, envKey: string): string | undefined {
  const envHeader = headers?.['x-gptscript-env']
  const envArray = Array.isArray(envHeader) ? envHeader : [envHeader]

  for (const env of envArray) {
    if (env == null) continue
    for (const pair of env.split(',')) {
      const [key, value] = pair.split('=').map((part) => part.trim())
      if (key === envKey) return value
    }
  }
  return undefined
}
