import { AtpAgent, RichText } from "@atproto/api"
import * as cheerio from "cheerio"

type Metadata = {
  title: string
  description: string
  image: string
}

/**
 * Extract Open Graph (OG) metadata from the content of page.
 * @param url - The URL to extract metadata from
 * @returns Metadata containing title, description, and image
 */
export const getPageMetadata = async (url: string): Promise<Metadata> => {
  const response = await fetch(url, { headers: { "User-Agent": "Mozilla/5.0" } })
  const html = await response.text()
  const $ = cheerio.load(html)

  const title = $('meta[property="og:title"]').attr("content") || $("title").text() || "Untitled"
  const description = $('meta[property="og:description"]').attr("content") || $('meta[name="description"]').attr("content") || ""
  const image = $('meta[property="og:image"]').attr("content") || ""

  return { title, description, image }
}

/**
 * Make a best effort attempt to generate the Bluesky embed card for the first link in the RichText's facets.
 * Returns undefined if the facets have no links or the first link's metadata could not be retrieved.
 * @param rt - The RichText object containing potential URLs
 * @param agent - The Bluesky agent
 * @returns The embed card if successfully generated, otherwise undefined
 */
export const getFirstEmbedCard = async (rt: RichText, agent: AtpAgent) => {
  const url = rt.facets?.find(facet => facet.features.some(f => f.$type === "app.bsky.richtext.facet#link"))
    ?.features.find(f => f.$type === "app.bsky.richtext.facet#link" && 'uri' in f)?.uri

  if (!url) return

  try {
    const metadata = await getPageMetadata(url)
    let thumbBlob

    if (metadata.image) {
      const blob = await fetch(metadata.image).then(r => r.blob())
      const { data } = await agent.uploadBlob(blob, { encoding: blob.type })
      thumbBlob = data.blob
    }

    return {
      $type: "app.bsky.embed.external", 
      external: {
        uri: url,
        title: metadata.title,
        description: metadata.description,
        thumb: thumbBlob,
      },
    }
  } catch (error) {
    console.error("Error fetching embed card:", error)
    return
  }
}