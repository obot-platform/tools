import axios from 'axios'
import { listJiraSites } from './src/sites.js'
import { listProjects, getProject } from './src/projects.js'
import { searchIssues, listIssues, getIssue, createIssue, editIssue } from './src/issues.js'
import { addComment, listComments } from './src/comments.js'
import { listPriorities } from './src/priorities.js'
import { listUsers, getUser, getCurrentUser } from './src/users.js'

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]

async function main() {

    try {
        if (!process.env.ATLASSIAN_OAUTH_TOKEN && (!process.env.ATLASSIAN_API_TOKEN || !process.env.ATLASSIAN_EMAIL || !process.env.ATLASSIAN_SITE_URL)) {
            throw new Error("ATLASSIAN_OAUTH_TOKEN or all of ATLASSIAN_API_TOKEN, ATLASSIAN_EMAIL, and ATLASSIAN_SITE_URL are required")
        }

        let authType = "Bearer"
        let authToken = process.env.ATLASSIAN_OAUTH_TOKEN
        if (process.env.ATLASSIAN_API_TOKEN) {
            authType = "Basic"
            authToken = Buffer.from(`${process.env.ATLASSIAN_EMAIL}:${process.env.ATLASSIAN_API_TOKEN.trim()}`).toString("base64")
        }

        let baseUrl
        if (command !== 'listJiraSites') {
            if (process.env.ATLASSIAN_SITE_URL) {
                baseUrl = process.env.ATLASSIAN_SITE_URL
                if (baseUrl.endsWith('/')) {
                    baseUrl = baseUrl.slice(0, -1)
                }
                if (!baseUrl.startsWith('http://') && !baseUrl.startsWith('https://')) {
                    baseUrl = `https://${baseUrl}`
                }
                baseUrl = `${baseUrl}/rest/api/3`
            } else if (process.env.SITE_ID) {
                baseUrl = `https://api.atlassian.com/ex/jira/${process.env.SITE_ID}/rest/api/3`
            } else {
                throw new Error('site_id argument not provided')
            }
        }
        const client = axios.create({
            baseURL: baseUrl,
            headers: {
                'Authorization': `${authType} ${authToken}`,
                'Accept': 'application/json',
            },
        })

        let result = null
        switch (command) {
            case "listJiraSites":
                if (process.env.ATLASSIAN_SITE_URL) {
                    console.log("listJiraSites is not needed when ATLASSIAN_SITE_URL is set, so you can continue without needing the site_id argument")
                    break
                }
                result = await listJiraSites(client)
                break
            case "createIssue":
                result = await createIssue(
                    client,
                    process.env.PROJECT_ID,
                    process.env.SUMMARY,
                    process.env.DESCRIPTION,
                    process.env.ISSUE_TYPE_ID,
                    process.env.PRIORITY_ID,
                    process.env.ASSIGNEE_ID,
                    process.env.REPORTER_ID
                )
                break
            case "editIssue":
                result = await editIssue(
                    client,
                    process.env.ISSUE_ID_OR_KEY,
                    process.env.NEW_SUMMARY,
                    process.env.NEW_DESCRIPTION,
                    process.env.NEW_ASSIGNEE_ID,
                    process.env.NEW_PRIORITY_ID,
                    process.env.NEW_NAME,
                    process.env.NEW_STATUS_NAME
                )
                break
            case "searchIssues":
                result = await searchIssues(client, process.env.JQL_QUERY)
                break
            case "listIssues":
                result = await listIssues(client, process.env.PROJECT_ID_OR_KEY)
                break
            case "getIssue":
                result = await getIssue(client, process.env.ISSUE_ID_OR_KEY)
                break
            case "addComment":
                result = await addComment(client, process.env.ISSUE_ID_OR_KEY, process.env.COMMENT_BODY)
                break
            case "listComments":
                result = await listComments(client, process.env.ISSUE_ID_OR_KEY)
                break
            case "listPriorities":
                result = await listPriorities(client, process.env.SCHEME_ID)
                break
            case "listUsers":
                result = await listUsers(client, process.env.INCLUDE_APP_USERS)
                break
            case "getUser":
                result = await getUser(client, process.env.ACCOUNT_ID)
                break
            case "getCurrentUser":
                result = await getCurrentUser(client)
                break
            case "listProjects":
                result = await listProjects(client)
                break
            case "getProject":
                result = await getProject(client, process.env.PROJECT_ID_OR_KEY)
                break
            default:
                throw new Error(`Unknown command: ${command}`)
        }
        console.log(JSON.stringify(result))
    } catch (error) {
        // We use console.log instead of console.error here so that it goes to stdout
        console.log(error)
        process.exit(1)
    }
}

await main()
