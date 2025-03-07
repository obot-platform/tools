import { RichText, AtpAgent, AppBskyFeedSearchPosts } from '@atproto/api'
import { getFirstEmbedCard } from "./embed.ts"

export async function searchPosts (
    agent: AtpAgent,
    query?: string,
    since?: string,
    until?: string,
    limit?: string,
    tags?: string,
): Promise<void> {
    let queryParams: AppBskyFeedSearchPosts.QueryParams = {
        q: query ?? '',
        sort: 'latest',
        limit: 25
    }

    if (!query) {
        throw new Error('Query is required')
    }

    if (!!limit) {
        try {
            queryParams.limit = parseInt(limit, 10)
        } catch (error: unknown) {
            throw new Error(`Invalid limit format: ${String(error)}`)
        }
    }

    if (!!until) {
        try {
            queryParams.until = new Date(until).toISOString()
        } catch (error: unknown) {
            throw new Error(`Invalid until date format: ${String(error)}`)
        }
    }

    if (!!since) {
        try {
            queryParams.since = new Date(since).toISOString()
        } catch (error: unknown) {
            throw new Error(`Invalid since date format: ${String(error)}`)
        }
    }

    if (!!tags) {
        queryParams.tag = tags
            .split(',')
            .map(tag => tag.trim().replace(/^#/, ''))
    }

    const response = await agent.app.bsky.feed.searchPosts(queryParams)

    console.log(JSON.stringify(response.data.posts))
}

export async function createPost(agent: AtpAgent, text?: string): Promise<void> {
    if (!text) {
        throw new Error('Text is required')
    }

    // Replace all instances of \\n with \n
    // The LLM sometimes double escapes newlines which will make them visible on bsky.app.
    text = text.replace(/\\n/g, '\n')

    const rt = new RichText({ text })
    await rt.detectFacets(agent)

    await agent.post({
        text: rt.text,
        facets: rt.facets,
        // Attempt to get the embed card for the first link in the text.
        // This renders a preview of the page content on bsky.app.
        embed: await getFirstEmbedCard(rt, agent),
    })

    console.log('Post created')
}

export async function deletePost(agent: AtpAgent, postUri?: string): Promise<void> {
    if (!postUri) {
        throw new Error('Post URI is required')
    }

    await agent.deletePost(postUri)

    console.log('Post deleted')
}
