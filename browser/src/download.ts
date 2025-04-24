import axios from 'axios'
import { GPTScript } from '@gptscript-ai/gptscript'
import { getWorkspaceId } from './session.ts'
import { type IncomingHttpHeaders } from 'node:http'
import * as mime from 'mime-types'
import * as path from 'path'

const client = new GPTScript()

export interface DownloadInfo {
  url: string
  resolvedUrl: string
  contentType?: string
  workspaceFile: string
  downloadedAt: number
  fileContents?: string
}

export async function download(headers: IncomingHttpHeaders, url: string): Promise<DownloadInfo> {
  const downloadedAt = Date.now()

  try {
    // Configure axios to follow redirects with a limit
    const response = await axios.get(url, {
      responseType: 'arraybuffer',
      maxRedirects: 5, // Set maximum redirects to follow
      validateStatus: (status) => status >= 200 && status < 400, // Allow redirects (3xx)
    })

    const contentType = response.headers['content-type']
    
    // Extract filename from URL
    let fileName = 'download';
    try {
      const urlObj = new URL(url);
      fileName = path.basename(urlObj.pathname).split('?')[0];
    } catch (e) {
      fileName = url.split('/').pop()?.split('?')[0] || 'download';
    }
    
    const extension = mime.extension(contentType);
    if (extension && !fileName.endsWith(`.${extension}`)) {
      fileName = `${fileName}.${extension}`;
    }

    const workspaceId = getWorkspaceId(headers)
    const workspaceFile = workspaceId ? `files/${fileName}` : fileName;
    await client.writeFileInWorkspace(workspaceFile, Buffer.from(response.data), workspaceId)

    let fileContents: string | undefined
    try {
      const run = await client.run("github.com/obot-platform/tools/workspace-files/tool.gpt", {
        subTool: "workspace_read",
        workspace: workspaceId,
        input: JSON.stringify({
          filename: fileName
        })
      })

      fileContents = await run.text()
    } catch (err) {
      fileContents = undefined
    }

    return {
        url,
        resolvedUrl: response.request.res.responseUrl || url,
        contentType,
        workspaceFile,
        downloadedAt,
        fileContents
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new Error(`Error downloading content from ${url}: ${msg}`)
  }
}
