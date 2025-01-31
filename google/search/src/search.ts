import { type BrowserContext, type Page } from '@playwright/test'
import * as cheerio from 'cheerio'
import TurndownService from 'turndown'
import prettier from 'prettier'

export interface SearchResult {
  url: string
  content?: string
  debug?: string[]
}

export interface SearchResults {
  query: string
  results: SearchResult[]
}


export async function search(
  context: BrowserContext,
  query: string,
  maxResults: number
): Promise<SearchResults> {
  if (query === '') {
    throw new Error('No query provided')
  }

  const encodedQuery = encodeURIComponent(query)
  const searchUrl = `https://www.google.com/search?q=${encodedQuery}&udm=14`

  const foundURLs = new Set<string>()
  const results: Array<Promise<SearchResult | null>> = []
  const searchPage = await context.newPage()

  try {
    await searchPage.goto(searchUrl)
    const content = await searchPage.content()
    const $ = cheerio.load(content)
    const elements = $('#rso a[jsname]')

    elements.each((_, element) => {
      if (results.length >= maxResults) return false

      const url = $(element).attr('href') ?? ''
      if (url && !url.includes('youtube.com/watch?v') && !foundURLs.has(url)) {
        foundURLs.add(url)
        // Create a new page per result and process it concurrently.
        results.push(
          (async () => {
            const childPage = await context.newPage()
            try {
              return await getMarkdown(childPage, url)
            } finally {
              await childPage.close().catch(() => {})
            }
          })()
        )
      }
    })

    const resolvedResults = (await Promise.all(results)).filter(Boolean) as SearchResult[]
    return { query, results: resolvedResults }
  } finally {
    await searchPage.close().catch(() => {})
  }
}


export async function getMarkdown (page: Page, url: string): Promise<SearchResult> {
    const result: SearchResult = { url, debug: [] }
    const pageContent = await stableContent(url, page, 500, 2000, 2000)
    const $ = cheerio.load(pageContent)

    // Remove scripts, styles, and iframes outright.
    $('noscript, script, style, iframe').remove()

    // Remove elements that are clearly ads based on class or id.
    $('[class*="ad-"], [id*="ad-"], [class*="advert"], [id*="advert"], .ads, .sponsored').remove()

    // Remove header, footer, nav, and aside elements.
    $('header, footer, nav, aside').remove()

    // Remove other known boilerplate selectors.
    $('.sidebar, .advertisement, .promo, .related-content').remove()

    // Try common selectors in order of preference.
    let content = ''
    const mainSelectors = [
      'main',
      'article',
      '.content',
      '.post-content',
      '.entry-content',
      '.main-content',
      'body'
    ]

    for (const selector of mainSelectors) {
      const section = $(selector)
      if (!section.length) continue

      let selected = ''
      for (const el of section) {
        selected += await toMarkdown($.html(el))
      }

      if (selected.length < 200) {
        result.debug?.push(`Selector ${selector} found but extracted no content, skipping...`)
        continue
      }

      content = selected
      result.debug?.push(`Extracted ${content.length} characters with selector: ${selector}`)
      break
    }

    result.content = content

    return {
      ...result,
      content: content,
    }
}

async function stableContent (
  url: string,
  page: Page,
  quietPeriod = 500,
  navigateTimeout = 2000,
  stablizeTimeout = 2000
): Promise<string> {
  try {
    // Wait up to 2s to navigate to the result URL.
    // Note: This handles redirects.
    await page.goto(url, { timeout: 1000 })
  } catch (e) {
    console.warn('Page :', url, e)
  }

  return await page.evaluate(
    ({ quietPeriod, stablizeTimeout }) => {
      return new Promise<string>((resolve) => {
        let quietTimer: number
        const observer = new MutationObserver(() => {
          clearTimeout(quietTimer)
          quietTimer = window.setTimeout(() => {
            observer.disconnect()
            // Capture and return the content when stability is reached
            resolve(document.documentElement.outerHTML)
          }, quietPeriod)
        })
        observer.observe(document.body, {
          childList: true,
          subtree: true,
          characterData: true
        })
        // Start a quiet timer in case no mutations occur
        quietTimer = window.setTimeout(() => {
          observer.disconnect()
          resolve(document.documentElement.outerHTML)
        }, quietPeriod)
        // Fallback: resolve after maxWait even if mutations continue
        window.setTimeout(() => {
          observer.disconnect()
          resolve(document.documentElement.outerHTML)
        }, stablizeTimeout)
      })
    },
    { quietPeriod, stablizeTimeout }
  )
}

// Create a TurndownService instance with compact options
const turndownService = new TurndownService({
  headingStyle: 'atx',      // One-line headings, e.g. "# Heading"
  bulletListMarker: '-',    // Use '-' for list items
  codeBlockStyle: 'fenced', // Use fenced code blocks (```)
  fence: '```',
  emDelimiter: '*',         // Use asterisk for emphasis
  strongDelimiter: '**',    // Use double asterisk for strong text
  linkStyle: 'inlined',     // User referenced style to reduce the number of links
})

// Configure Prettier to produce compact markdown
const prettierOptions: prettier.Options = {
  parser: 'markdown',
  printWidth: 9999,      // Set very high to avoid wrapping lines
  proseWrap: 'never',    // Don't force wrapping of prose
  tabWidth: 1,           // Use a single space for indentation (minimum available)
  useTabs: false,
  trailingComma: 'none'
};

async function toMarkdown (html: string): Promise<string> {
    let md = turndownService.turndown(html)
    md = await prettier.format(md, prettierOptions)
    return md.replace(/\n{3,}/g, '\n\n').trim()
}
