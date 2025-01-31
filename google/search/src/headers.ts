import { type IncomingHttpHeaders } from 'node:http'
import { createHash } from 'node:crypto'

export interface ModelProviderCredentials {
    baseUrl: string
    apiKey: string
}

export function getModelProviderCredentials(headers: IncomingHttpHeaders): ModelProviderCredentials | undefined {
  const baseUrl = getGPTScriptEnv(headers, 'OPENAI_BASE_URL')?.trim()
  if (!baseUrl) return undefined

  const apiKey = getGPTScriptEnv(headers, 'OPENAI_API_KEY')?.trim()
  if (!apiKey) return undefined

  return { baseUrl, apiKey }
}

export function getSessionId(headers: IncomingHttpHeaders): string {
  const workspaceId = getGPTScriptEnv(headers, 'GPTSCRIPT_WORKSPACE_ID')
  if (!workspaceId?.trim()) throw new Error('No GPTScript workspace ID provided')

  return createHash('sha256').update(workspaceId).digest('hex').substring(0, 16)
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