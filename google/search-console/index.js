import { fetchGSCProperties } from './src/siteList.js';
import { getUrlsForProperty } from './src/urlsForProperty.js';
import { inspectUrl } from './src/urlInspection.js';
import { getTopKeywordsForUrl } from './src/keywords.js';

const oauthToken = process.env.GOOGLE_SEARCH_OAUTH_TOKEN;

async function main() {
    const command = process.argv[2];
    console.log(command);
    if (!oauthToken) {
        console.error('Error: OAUTH_TOKEN environment variable is not set.');
        return;
    }

    const propertyUrl = process.env.PROPERTY;
    const url = process.env.URL;

    switch (command) {
        case 'getUrlsForProperty':
            if (!propertyUrl) {
                console.error('Error: PROPERTY environment variable is not set.');
                return;
            }
            await handleGetUrlsForProperty(propertyUrl, oauthToken);
            break;

        case 'fetchGSCProperties':
            await handleListProperties(oauthToken);
            break;

        case 'inspectUrl':
            if (!propertyUrl) {
                console.error('Error: PROPERTY environment variable is not set.');
                return;
            }
            if (!url) {
                console.error('Error: URL environment variable is not set.');
                return;
            }
            await handleInspectUrl(propertyUrl, url, oauthToken);
            break;

        case 'getTopKeywordsForUrl':
            if (!propertyUrl) {
                console.error('Error: PROPERTY environment variable is not set.');
                return;
            }

            if (!url) {
                console.error('Error: URL environment variable is not set.');
                return;
            }

            await handleGetTopKeywordsForUrl(propertyUrl, url, oauthToken);
            break;

        default:
            console.error('Error: Unknown command. Use "getUrlsForProperty, fetchGSCProperties, or inspectUrl".');
    }
}

async function handleGetUrlsForProperty(propertyUrl, oauthToken) {
    try {
        const urls = await getUrlsForProperty(propertyUrl, oauthToken);
        console.log('Fetched URLs:', urls);
    } catch (error) {
        console.error('Error fetching URLs:', error);
    }
}

async function handleListProperties(oauthToken) {
    try {
        const properties = await fetchGSCProperties(oauthToken);
        console.log('GSC Properties:', properties);
    } catch (error) {
        console.error('Error listing properties:', error);
    }
}

async function handleInspectUrl(propertyUrl, url, oauthToken) {
    try {
        const inspectionResult = await inspectUrl(propertyUrl, url, oauthToken);
        console.log('Inspection Result:', inspectionResult);
    } catch (error) {
        console.error('Error inspecting URL:', error);
    }
}

async function handleGetTopKeywordsForUrl(propertyUrl, url, oauthToken) {
    try {
        const keywords = await getTopKeywordsForUrl(propertyUrl, url, oauthToken);
        console.log('Top Keywords:', keywords);
    } catch (error) {
        console.error('Error fetching top keywords:', error);
    }
}

main();