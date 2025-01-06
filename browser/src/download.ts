import axios from 'axios'
import { GPTScript } from '@gptscript-ai/gptscript'
import { getWorkspaceId, getGPTScriptEnv } from './session.ts'
import { type IncomingHttpHeaders } from 'node:http'

const client = new GPTScript()

export interface DownloadInfo {
  url: string
  resolvedUrl: string
  contentType?: string
  workspaceFile: string
  downloadedAt: number
}

export async function download(
  headers: IncomingHttpHeaders,
  url: string,
  fileName: string
): Promise<DownloadInfo> {
  const downloadedAt = Date.now()

  try {
    const workspaceId = getWorkspaceId(headers)
    const workspaceFile = workspaceId !== undefined ? `files/${fileName}` : fileName

    // Configure axios to follow redirects with a limit
    const response = await axios.get(url, {
      responseType: 'arraybuffer',
      maxRedirects: 5, // Set maximum redirects to follow
      validateStatus: (status) => status >= 200 && status < 400, // Allow redirects (3xx)
    })

    const contentType = response.headers['content-type']
    await client.writeFileInWorkspace(workspaceFile, Buffer.from(response.data), workspaceId)

    return { 
        url,
        resolvedUrl: response.request.res.responseUrl || url,
        contentType,
        workspaceFile,
        downloadedAt
    } // Use the final URL after redirects
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new Error(`Error downloading content from ${url}: ${msg}`)
  }
}
