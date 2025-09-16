import { AtpAgent, AppBskyActorSearchActors } from '@atproto/api'

export async function searchUsers(
    agent: AtpAgent,
    query?: string,
    limit?: string,
): Promise<any> {
    let queryParams: AppBskyActorSearchActors.QueryParams = {
        q: query ?? '',
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

    const response = await agent.app.bsky.actor.searchActors(queryParams)

    return response.data.actors
}