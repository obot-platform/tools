import { encoding_for_model } from "tiktoken"
import {GPTScript, type ToolDef} from "@gptscript-ai/gptscript"
import {type SearchResult, type SearchResults} from "./search.ts"
import {type ModelProviderCredentials} from "./headers.ts"

// Max number of tokens in the search results
const MAX_RESULTS_TOKENS = 50000

const gptscript = new GPTScript()

export async function refine (creds: ModelProviderCredentials | undefined, unrefined: SearchResults): Promise<SearchResults> {
  const totalUnrefinedTokens = tokenCount(unrefined.results.reduce((acc, result) => acc + result.content, ''))
  if (totalUnrefinedTokens <= MAX_RESULTS_TOKENS) {
    console.info(`Total tokens (${totalUnrefinedTokens}) are within the limit (${MAX_RESULTS_TOKENS}), skipping refinement`)
    return unrefined
  }

  if (!creds) {
    console.warn('No model provider credentials provided, skipping refinement')
    return unrefined
  }

  console.info(`Total tokens (${totalUnrefinedTokens}) are above the limit (${MAX_RESULTS_TOKENS}), calling GPTScript to refine results`)

  const now = userDateTime()
  let refined = await Promise.all(
    unrefined.results.map(async (result) => {
      const refinedContent = await refineContent(creds, now, unrefined.query, result)
      const refinedTokens = tokenCount(refinedContent.content)
      return {
        ...result,
        ...refinedContent,
        refinedTokens 
      }
    })
  )

  const totalRefinedTokens = refined.reduce((sum, r) => sum + r.refinedTokens, 0)
  if (totalRefinedTokens <= MAX_RESULTS_TOKENS) {
    // If the refined tokens already fit the limit, return as is.
    return { query: unrefined.query, results: refined }
  }

  // Filter zero score or zero tokens
  refined = refined.filter(r => r.score > 0 && r.refinedTokens > 0)

  // Sort by "value density" = score / tokens (descending)
  refined.sort((a, b) => (b.score / b.refinedTokens) - (a.score / a.refinedTokens))

  const pruned: SearchResult[] = []
  let tokenBudget = MAX_RESULTS_TOKENS

  for (const r of refined) {
    if (tokenBudget < 1) break

    if (r.refinedTokens >= tokenBudget) {
      // If the result is too long, truncate it to fit the budget
      const truncated = truncateContent(r.content, tokenBudget)
      pruned.push({
        ...r,
        content: truncated.content,
      })

      // Consume the tokens from the budget
      tokenBudget -= truncated.tokenCount
      continue
    }

    // The entire result fits in the budget, so add it to the pruned results
    pruned.push(r)
    tokenBudget -= r.refinedTokens
  }

  return { query: unrefined.query, results: pruned }
}

function tokenCount (content?: string): number {
  if (!content || content.length === 0) {
    return 0
  }

  const enc = encoding_for_model('gpt-4o-mini');
  try {
    return enc.encode(content).length;
  } catch (e) {
    console.warn('Error encoding content', e);
  } finally {
    // Free encoding resources when done
    enc.free()
  }

  return 0
}


function truncateContent (content: string, maxTokens: number): {
  content: string,
  tokenCount: number
} {
  const codec = encoding_for_model('gpt-4o-mini');
  try {
    const tokens = codec.encode(content)
    const truncated = tokens.slice(0, maxTokens)
    return {
      content: new TextDecoder().decode(truncated),
      tokenCount: truncated.length
    }
  } finally {
    codec.free()
  }
}


function userDateTime (): string {
  const tz = process.env.TIMEZONE || 'UTC';
  try {
    new Intl.DateTimeFormat('en-US', { timeZone: tz });
  } catch {
    return new Date().toLocaleString('en-US', { timeZone: 'UTC', timeZoneName: 'short' });
  }
  return new Date().toLocaleString('en-US', { timeZone: tz, timeZoneName: 'short' });
}


async function refineContent (
  creds: ModelProviderCredentials,
  time: string,
  query: string,
  result: SearchResult): Promise<{
    content: string,
    score: number
  }> {

  const tool: ToolDef = {
    chat: false,
    jsonResponse: false,
    modelName: process.env.OBOT_DEFAULT_LLM_MINI_MODEL ?? 'gpt-4o-mini',
    temperature: 0.0,
    arguments: {
      type: 'object',
      properties: {
        time: {
          type: 'string',
          description: 'Current date and time that the search was requested at'
        },
        topic: {
          type: 'string',
          description: 'Topic to extract excerpts for'
        },
        url: {
          type: 'string',
          description: 'URL that the markdown content was sourced from'
        },
        content: {
          type: 'string',
          description: 'Markdown document created by exporting an HTML web page to markdown'
        }
      },
      required: ['time', 'topic', 'url', 'content']
    },
    instructions: refineInstructions
  }

  const run = await gptscript.evaluate(tool, {
    BaseURL: creds.baseUrl,
    APIKey: creds.apiKey,
    input: minify({
      time,
      topic: query,
      url: result.url,
      content: result.content
    })
  })

  // Parse the output into a score and content
  const output = await run.text()
  const [firstLine, ...restLines] = output?.split('\n') ?? []
  const score = Math.max(1, Math.min(10, parseInt(firstLine, 10))) || 0
  const content = restLines.join('\n')

  return { score, content }
}

// Note: Tools can't introspect their parameters schema, so we provide it in the instructions as well
const refineInstructions = `
Do not respond with any additional dialog or commentary.

You are a research assistant tasked with extracting excerpts from a markdown document that will
be used as notes to conduct detailed research about a given topic.

The document is the result of exporting an HTML webpage to markdown.

When given an object with the following JSON schema:

${minify({
  type: 'object',
  properties: {
    time: {
      type: 'string',
      description: 'Current date and time that the search was requested at'
    },
    topic: {
      type: 'string',
      description: 'Topic to extract excerpts for'
    },
    url: {
      type: 'string',
      description: 'URL that the markdown content was sourced from'
    },
    content: {
      type: 'string',
      description: 'Markdown document created by exporting an HTML web page to markdown'
    }
  },
  required: ['time', 'topic', 'url', 'content', 'time']
})}

Perform the following steps in order:
1. Refine the markdown content by removing all:
  - boilerplate and unintelligable text
  - unrelated advertisements, links, and web page structure
2. Select excerpts from the refined content that you think would make good notes for conducting detailed research about the topic
3. Compose a concise markdown document containing the excerpts organized in descending order of importance to understanding the topic. Do not paraphrase, summarize, or reword the excerpts. The goal is to preserve as much of the original content as possible.
4. Grade the corpus of excerpts as a whole based how well it covers the topic on a scale of 0-10, where high scores are good and low scores contain no relevant information

Afterwards, respond with the grade followed by the markdown document on a new line.

EXAMPLE
5
<content of markdown document>
`

function minify (obj: object): string {
  return JSON.stringify(obj).replace(/\n/g, '')
}
