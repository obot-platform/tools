import { startServer } from './server.ts'
import { startSlackBot } from './bot.ts'


try {
    const obotAPIToken = process.env.OBOT_API_TOKEN
    if (!obotAPIToken) {
        throw new Error('OBOT_API_TOKEN is required')
    }
    const obotServerUrl = process.env.OBOT_SERVER_URL
    if (!obotServerUrl) {
        throw new Error('OBOT_BASE_URL is required')
    }
    const botToken = process.env.OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_BOT_TOKEN
    if (!botToken) {
        throw new Error('OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_BOT_TOKEN is required')
    }
    const appToken = process.env.OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_APP_TOKEN
    if (!appToken) {
        throw new Error('OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_APP_TOKEN is required')
    }
    const signingSecret = process.env.OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_SIGNING_SECRET
    if (!signingSecret) {
        throw new Error('OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_SIGNING_SECRET is required')
    }

    const slackBot = await startSlackBot(
        obotAPIToken,
        obotServerUrl,
        botToken,
        appToken,
        signingSecret
    )

    const port = parseInt(process.env.PORT ?? '9888', 10)
    const server = startServer(port)

    let stopped = false
    const stop = async (): Promise<void> => {
        if (stopped) return
        stopped = true
        console.error('Daemon shutting down...')
        await slackBot.stop()
        server.close(() => process.exit(0))
    }

    process.stdin.resume()
    process.stdin.on('close', stop);
    ['SIGINT', 'SIGTERM', 'SIGHUP'].forEach(signal => process.on(signal, stop))

} catch (error) {
    console.error('Error initializing app:', error)
    process.exit(1)
}