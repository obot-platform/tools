import slack from '@slack/bolt'
import axios from 'axios'
import { StaticSelectAction } from '@slack/bolt'
import { KnownBlock } from '@slack/types'
import { createParser, EventSourceMessage } from 'eventsource-parser'

// TODO(njhale): Remove this after local testing
const envRegex = /^(OBOT_SERVER_URL|OBOT_API_TOKEN|OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_BOT_TOKEN|OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_APP_TOKEN|OBOT_SLACK_DAEMON_TRIGGER_PROVIDER_SIGNING_SECRET|PORT|TEST_DISABLE)/
const filteredEnv = Object.entries(process.env)
    .filter(([key]) => envRegex.test(key))
    .map(([key, value]) => `export ${key}='${value}'`)
    .join("\n")

const surround = (str: string, delim: string) => `\n${delim}\n${str}\n${delim}\n`
console.warn(surround(filteredEnv, '*'.repeat(64)))

export const startSlackBot = async (
    obotAPIToken: string,
    obotServerUrl: string,
    botToken: string,
    appToken: string,
    signingSecret: string
) => {
    const obot = axios.create({
        baseURL: obotServerUrl + '/api',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${obotAPIToken}`,
        }
    })

    const bot = new slack.App({
        token: botToken,
        appToken: appToken,
        signingSecret: signingSecret,
        deferInitialization: true,
        socketMode: true,
        logLevel: slack.LogLevel.INFO,
    })

    // TODO(njhale): Remove this after local testing
    bot.use(async ({ context, body, next }) => {
        console.log('---- Incoming Slack Event -----')
        console.log('Context:', context)
        console.log('Body:', JSON.stringify(body, null, 2))
        console.log('-------------------------------')
        await next()
    })

    bot.command('/obot-run-workflow', async ({ command, client, ack, respond }) => {
        console.warn('Command received:', command);
        await ack();

        try {
            console.warn('Fetching daemon triggers');
            const response = await obot.get('/daemon-triggers?provider=slack-daemon-trigger-provider&withExecutions=true');
            const daemonTriggers = response.data.items;

            if (daemonTriggers.length === 0) {
                await respond({ text: 'No available workflows to run.' });
                return;
            }
            const channelId = command.channel_id!

            if (command.text !== '') {
                console.warn('Searching for workflow:', command.text);
                const selectedWorkflow = daemonTriggers.find((daemonTrigger: any) =>
                    daemonTrigger.workflow === command.text
                );

                if (selectedWorkflow) {
                    await runWorkflow(selectedWorkflow.workflow, client, channelId);
                    return;
                }

                await respond({ text: `Workflow *${command.text}* not available.` });
                return;
            }

            const options = daemonTriggers.map((daemonTrigger: any) => ({
                text: {
                    type: 'plain_text',
                    text: `${daemonTrigger.workflowExecutions?.find((exec: any) => exec?.workflow?.name)?.workflow?.name || daemonTrigger.workflow} (${daemonTrigger.workflow})`,
                },
                value: daemonTrigger.workflow
            }));

            await respond({
                text: 'Select a workflow to run:',
                blocks: [
                    {
                        type: 'section',
                        text: { type: 'mrkdwn', text: 'Please select a workflow to run:' },
                        accessory: {
                            type: 'static_select',
                            action_id: 'select_obot_workflow',
                            placeholder: { type: 'plain_text', text: 'Select a workflow' },
                            options: options,
                            confirm: {
                                style: 'primary',
                                title: {
                                    type: 'plain_text',
                                    text: 'Confirm workflow run'
                                },
                                text: {
                                    type: 'plain_text',
                                    text: 'Are you sure you want to run this workflow?'
                                },
                                confirm: {
                                    type: 'plain_text',
                                    text: 'Run workflow'
                                },
                                deny: {
                                    type: 'plain_text',
                                    text: 'Cancel'
                                }
                            }
                        }
                    }
                ]
            });
        } catch (error) {
            console.error('Error fetching daemon triggers:', error);
            await respond({ text: 'Failed to fetch workflows. Please try again later.' });
        }
    });

    bot.action('select_obot_workflow', async ({ context, body, action, client, ack, respond }) => {
        console.warn('Action received:', action);
        await ack();

        if ('selected_option' in action) {
            const selectedWorkflow = (action as StaticSelectAction).selected_option?.value;
            console.warn('Selected workflow:', selectedWorkflow);

            const channel = body.channel?.id
            const user = context.userId

            try {
                await runWorkflow(selectedWorkflow, client, channel!, user);
            } catch (error) {
                const message = error instanceof Error ? error.message : 'Failed to invoke workflow. Please try again later.';
                console.error('Error invoking workflow:', message);

                const blocks: KnownBlock[] = [
                    {
                        type: 'section',
                        text: { type: 'mrkdwn', text: message }
                    }
                ];

                if (error instanceof Error && error.cause) {
                    blocks.push({
                        type: 'context',
                        elements: [
                            { type: 'mrkdwn', text: `> ${error.cause}` }
                        ]
                    });
                }

                await respond({ blocks });
            }
        } else {
            await respond({ text: 'Invalid action received.' });
        }
    });

    async function runWorkflow(
        workflowId: string,
        client: any,
        channel: string,
        user?: string
    ) {
        const messageContext: KnownBlock[] = user ? [
            {
                type: 'context',
                elements: [
                    { type: 'mrkdwn', text: `Started by: <@${user}>` }
                ]
            },
            {
                type: 'context',
                elements: [
                    { type: 'mrkdwn', text: `Workflow ID: <${obotServerUrl}/admin/workflows/${workflowId}|${workflowId}>` }
                ]
            }
        ] : [];

        try {
            const workflow = await obot.get(`/workflows/${workflowId}`);
            const workflowName = workflow.data.name;

            const initialMessage = await client.chat.postMessage({
                channel,
                blocks: [
                    { type: 'header', text: { type: 'plain_text', text: `🔄 Running: ${workflowName}` } },
                    ...messageContext
                ]
            });

            const invokeResponse = await obot.post(`/invoke/${workflowId}`);
            const systemThreadId = invokeResponse.data.threadID;
            messageContext.push({
                type: 'context',
                elements: [{ type: 'mrkdwn', text: `Thread: <${obotServerUrl}/admin/threads/${systemThreadId}|${systemThreadId}>` }]
            });

            const eventsResponse = await obot.get(
                `/threads/${systemThreadId}/events?follow=true&waitForThread=true`,
                {
                    headers: { Accept: 'text/event-stream' },
                    responseType: 'stream'
                }
            );

            const parser = createParser({
                onEvent: async (event: EventSourceMessage) => {
                    if (event.event === 'close') return;

                    const data = JSON.parse(event.data || '{}');
                    const { runID, parentRunID, time, content, contentID, waitingOnModel, step, runComplete } = data;
                    const timestamp = time ? new Date(time).toLocaleString() : new Date().toLocaleString();

                    // Log start of step
                    if (step) {
                        await client.chat.postMessage({
                            channel,
                            thread_ts: initialMessage.ts,
                            blocks: [
                                {
                                    type: 'section',
                                    text: {
                                        type: 'mrkdwn',
                                        text: [
                                            `🔄 *Starting Step*: \`${step.step}\``,
                                            `*Run ID*: \`${runID}\`${parentRunID ? ` (Parent: \`${parentRunID}\`)` : ''}`,
                                            `*Step ID*: \`${step.id}\``,
                                            `*Time*: ${timestamp}`
                                        ].join('\n')
                                    }
                                }
                            ]
                        });
                    }

                    // Log model waiting state
                    if (waitingOnModel) {
                        await client.chat.postMessage({
                            channel,
                            thread_ts: initialMessage.ts,
                            text: `⏳ Waiting for model response... (Content ID: \`${contentID}\`)`
                        });
                    }

                    // Log step output
                    if (contentID && content && !waitingOnModel) {
                        await client.chat.postMessage({
                            channel,
                            thread_ts: initialMessage.ts,
                            blocks: [
                                {
                                    type: 'section',
                                    text: {
                                        type: 'mrkdwn',
                                        text: [
                                            `📝 *Output* (Content ID: \`${contentID}\`)`,
                                            '```',
                                            content,
                                            '```'
                                        ].join('\n')
                                    }
                                }
                            ]
                        });
                    }

                    // Log run completion
                    if (runComplete) {
                        const message = parentRunID ?
                            `✅ Run \`${runID}\` completed` :
                            `✅ Workflow *${workflowName}* completed successfully!`;

                        await client.chat.postMessage({
                            channel,
                            thread_ts: initialMessage.ts,
                            blocks: [
                                {
                                    type: 'section',
                                    text: {
                                        type: 'mrkdwn',
                                        text: `${message}\n*Time*: ${timestamp}`
                                    }
                                }
                            ]
                        });

                        if (!parentRunID) {
                            await client.chat.update({
                                channel,
                                ts: initialMessage.ts,
                                blocks: [
                                    { type: 'header', text: { type: 'plain_text', text: `✅ Complete: ${workflowName}` } },
                                    ...messageContext
                                ]
                            });
                        }
                    }
                },
                onError: (err) => console.error('Error parsing event:', err)
            });

            for await (const chunk of eventsResponse.data) {
                parser.feed(new TextDecoder().decode(chunk));
            }
        } catch (error) {
            console.error('Error invoking workflow:', error);
            throw new Error(`❌ Failed to invoke workflow *${workflowId}*`, { cause: error });
        }
    }

    bot.command('/obot-list-workflows', async ({ command, ack, respond }) => {
        console.warn('/obot-list-workflows:', command)
        await ack()

        try {
            const response = await obot.get('/daemon-triggers?provider=slack-daemon-trigger-provider&withExecutions=true');
            const daemonTriggers = response.data.items;

            const blocks = daemonTriggers.flatMap((daemonTrigger: any) => {
                const workflowExecutions = daemonTrigger.workflowExecutions || [];
                const latestExecution = workflowExecutions[workflowExecutions.length - 1] || {};
                const executionCount = workflowExecutions.length;
                const latestState = latestExecution.state || 'No executions';
                const latestError = latestExecution.error || 'No error';
                const workflowName = latestExecution.workflow?.name || daemonTrigger.workflow.name;

                return [
                    { type: 'divider' },
                    {
                        type: 'section',
                        text: {
                            type: 'mrkdwn',
                            text: `*${workflowName}*\n*State:* ${latestState}\n*Error:* ${latestError}\n*Runs:* ${executionCount}`
                        }
                    },
                    { type: 'divider' }
                ];
            });

            if (blocks.length === 0) {
                blocks.push({
                    type: 'section',
                    text: {
                        type: 'mrkdwn',
                        text: 'No workflows found.'
                    }
                });
            }

            await respond({
                text: `The available workflows are listed below`,
                blocks: [
                    {
                        type: 'header',
                        text: {
                            type: 'plain_text',
                            text: 'Workflows'
                        }
                    },
                    ...blocks
                ]
            });
        } catch (error) {
            console.error('Error fetching workflows:', error);
            await respond({
                text: `Hello <@${command.user_id}>!
                I'm sorry, but I couldn't fetch the workflows. Please try again later.`,
            });
        }

        console.warn('Responded to command:', command.text)
    })

    bot.command('/obot-get-workflow', async ({ command, ack, respond }) => {
        console.warn('Command received:', command)
        await ack()

        try {
            const workflowId = command.text.trim()
            if (!workflowId) {
                await respond({
                    text: `Please provide a workflow ID. Usage: /obot-get-workflow <workflow_id>`
                })
                return
            }

            const { data: workflow } = await obot.get(`/workflows/${workflowId}`)
            const { data: executions } = await obot.get(`/workflows/${workflowId}/executions`)

            const latestExecution = executions[0] || {}
            const latestState = latestExecution.state || 'No state'
            const latestError = latestExecution.error || 'No error'
            const executionCount = executions.length

            await respond({
                blocks: [
                    {
                        type: 'header',
                        text: {
                            type: 'plain_text',
                            text: 'Workflow Details'
                        }
                    },
                    { type: 'divider' },
                    {
                        type: 'section',
                        text: {
                            type: 'mrkdwn',
                            text: `*${workflow.name}*\n*State:* ${latestState}\n*Error:* ${latestError}\n*Runs:* ${executionCount}`
                        }
                    },
                    { type: 'divider' }
                ]
            })
        } catch (error) {
            console.error('Error fetching workflow:', error)
            await respond({
                text: `Hello <@${command.user_id}>! I'm sorry, but I couldn't fetch the workflow. Please check the ID and try again.`
            })
        }

        console.warn('Responded to command:', command.text)
    })

    await bot.init()
    await bot.start()

    return bot
}
