import { fileTypeFromBuffer } from 'file-type'
import { resolve } from 'path'
import { readFile } from 'fs/promises'
import OpenAI from 'openai'
import { type ChatCompletionContentPartImage } from 'openai/resources/chat/completions'
import { GPTScript } from '@gptscript-ai/gptscript'

export async function analyzeImages (
  prompt: string = '',
  images: string = ''
): Promise<void> {
  if (prompt === '') {
    prompt = 'Provide a brief description of each image'
  }

  const imageArray = images
    .split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0)

  if (imageArray.length === 0) {
    throw new Error('No images provided. Please provide a list of images to send to the vision model.')
  }

  const content = await Promise.all(
    imageArray.map(async (image: string) => ({
      type: 'image_url',
      image_url: {
        detail: 'auto',
        url: await resolveImageURL(image)
      }
    }))
  ) as ChatCompletionContentPartImage[]

  const openai = new OpenAI()
  const response = await openai.chat.completions.create({
    model: process.env.OBOT_DEFAULT_VISION_MODEL ?? 'gpt-4o',
    stream: true,
    messages: [
      {
        role: 'user',
        content: [{ type: 'text', text: prompt }, ...content]
      }
    ]
  })

  for await (const part of response) {
    const { choices } = part
    const text = choices?.[0]?.delta?.content
    if (text !== null && text !== undefined) {
      process.stdout.write(text)
    }
  }
}

const supportedMimeTypes = ['image/jpeg', 'image/png', 'image/webp']
const threadId = process.env.OBOT_THREAD_ID
const obotServerUrl = process.env.OBOT_SERVER_URL
const imageGenBaseUrl = (threadId !== undefined && obotServerUrl !== undefined) ? `${obotServerUrl}/api/threads/${threadId}/file/` : null

async function resolveImageURL (image: string): Promise<string> {
  // If the image is a URL, return it as is
  if (image.includes('://')) {
    const url = new URL(image)
    switch (url.protocol) {
      case 'http:':
      case 'https:':
        if (imageGenBaseUrl == null || !image.startsWith(imageGenBaseUrl)) {
          return image
        }
        // This is a generated image download link, strip the base URL and retrieve the file from the workspace
        image = image.replace(imageGenBaseUrl, '')
        break
      default:
        throw new Error(`Unsupported image URL protocol: ${url.protocol}`)
    }
  }

  // Read the image file from the workspace and check its MIME type
  const data = await readImageFile(image)
  const mime = (await fileTypeFromBuffer(data))?.mime
  if (mime === undefined || !supportedMimeTypes.includes(mime)) {
    throw new Error(`Unsupported image file type ${mime}, expected one of ${supportedMimeTypes.join(', ')}`)
  }

  // Encode the image file as a base64 string and return it as a data URL
  const base64 = data.toString('base64')
  return `data:${mime};base64,${base64}`
}

async function readImageFile (path: string): Promise<Buffer> {
  if (threadId === undefined) {
    // Not running in Obot, just read the file
    return await readFile(resolve(path))
  }

  // The Generate Images tool returns file paths with a special prefix
  // so that they can be rendered in the Obot UI.
  // e.g. /api/threads/<thread-id>/file/generated_image_<hash>.webp
  // It must be stripped before reading the file from the workspace
  path = path.replace(/^\/?api\/threads\/[a-z0-9]+\/file\//, '')

  const client = new GPTScript()
  return Buffer.from(await client.readFileInWorkspace(`files/${path}`))
}
