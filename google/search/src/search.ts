import { type BrowserContext, type Page } from '@playwright/test'
import * as cheerio from 'cheerio';
import TurndownService from 'turndown'

export interface SearchResult {
  url: string
  title?: string
  content?: string | string[]
}

export interface SearchResults {
  query: string
  results: SearchResult[]
}

export async function search (
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

  const page = await context.newPage()
  const pages = await Promise.all(
    Array.from({ length: maxResults }, () => context.newPage())
  )

  try {
    await page.goto(searchUrl)
    const content = await page.content()
    const $ = cheerio.load(content)
    const elements = $('#rso a[jsname]')

    elements.each((_, element) => {
      if (results.length >= maxResults) return false

      const url = $(element).attr('href') ?? ''
      if ((url !== '') && !url.includes('youtube.com/watch?v') && !foundURLs.has(url)) {
        foundURLs.add(url)
        results.push(getMarkdown(pages[results.length], url).then(content => {
          return (content !== '') ? { url, content } : null
        }))
      }
    })

    return {
      query,
      results: (await Promise.all(results)).filter(Boolean) as SearchResult[]
    }
  } finally {
    // Fire and forget page close so we can move on
    void page.close()
    void Promise.all(pages.map(async p => { await p.close() }))
  }
}

export async function getMarkdown (page: Page, url: string): Promise<string> {
  try {
    await page.goto(url, { timeout: 1000 })
    await page.waitForLoadState('networkidle', { timeout: 1000 })
  } catch (e) {
    console.warn('slow page:', url)
  }

  let content = ''
  while (content === '') {
    let fails = 0
    try {
      content = await page.evaluate(() => document.documentElement.outerHTML)
    } catch (e) {
      fails++
      if (fails > 2) {
        void page.close()
        console.warn('rip:', url)
        return '' // Page didn't load; just ignore.
      }
      await new Promise(resolve => setTimeout(resolve, 100)) // sleep 100ms
    }
  }
  void page.close()

  const $ = cheerio.load(content)

  // Remove common elements that are not part of the page content
  $('noscript, script, style, img, g, svg, iframe').remove();
  $('header, footer, nav, aside').remove();
  $('.sidebar, .advertisement, .promo, .related-content').remove();

  let resp = ''
  const turndownService = new TurndownService({
    headingStyle: 'atx',
    bulletListMarker: '-',
  })

  // Prioritize main content selectors, eventually falling back to the full body
  const mainSelectors = ['main', 'article', '.content', '.post-content', '.entry-content', '.main-content', 'body'];
  for (const selector of mainSelectors) {
    if ($(selector).first().length < 1) {
      continue;
    }


    $(selector).each(function () {
      resp += turndownService.turndown($.html(this))
    })
    break
  }

  return trunc(resp, 50000)
}

function trunc (text: string, max: number): string {
  return text.length > max ? text.slice(0, max) + '...' : text
}
