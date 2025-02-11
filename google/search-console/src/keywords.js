// keywords.js
import fetch from 'node-fetch';

const USER_TIMEZONE = process.env.OBOT_USER_TIMEZONE || 'UTC';

function formatDateInUserTimezone(date) {
    return date.toLocaleDateString('en-CA', { timeZone: USER_TIMEZONE });
}

/**
 * Fetches the top keywords for a given URL from Google Search Console.
 * @param {string} property - The GSC property URL.
 * @param {string} url - The specific URL to retrieve keywords for.
 * @param {string} oauthToken - The OAuth token for authentication.
 * @returns {Promise<Array>} - A promise that resolves to an array of keyword data.
 */
export async function getTopKeywordsForUrl(property, url, oauthToken) {
    const apiUrl = `https://www.googleapis.com/webmasters/v3/sites/${encodeURIComponent(property)}/searchAnalytics/query`;

    const payload = {
        startDate: getThreeMonthsAgo(),
        endDate: getToday(),
        dimensions: ["query"],
        dimensionFilterGroups: [
            {
                filters: [
                    {
                        dimension: "page",
                        operator: "equals",
                        expression: url
                    }
                ]
            }
        ],
        rowLimit: 10,
        orderBy: [{ fieldName: "clicks", sortOrder: "DESCENDING" }]
    };

    const headers = {
        Authorization: `Bearer ${oauthToken}`,
        "Content-Type": "application/json"
    };

    const options = {
        method: "POST",
        headers: headers,
        body: JSON.stringify(payload)
    };

    try {
        const response = await fetch(apiUrl, options);
        const responseCode = response.status;
        const responseText = await response.text();

        console.log(`Response Code: ${responseCode}`);
        console.log(`Response Content: ${responseText}`);

        if (responseCode === 200) {
            const json = JSON.parse(responseText);
            return json.rows ? json.rows.map(row => ({
                query: row.keys[0],
                clicks: row.clicks,
                impressions: row.impressions,
                ctr: row.ctr
            })) : [];
        } else {
            console.error(`Failed to fetch data: ${responseCode} - ${responseText}`);
            throw new Error(`Failed to fetch data: ${responseCode} - ${responseText}`);
        }
    } catch (error) {
        console.error(`Error: ${error.message}`);
        throw new Error(`Error fetching keywords: ${error.message}`);
    }
}

function getToday() {
    return formatDateInUserTimezone(new Date());
}

function getThreeMonthsAgo() {
    const date = new Date();
    // Create date in user's timezone
    const localDate = new Date(date.toLocaleString('en-US', { timeZone: USER_TIMEZONE }));
    localDate.setMonth(localDate.getMonth() - 3);
    return formatDateInUserTimezone(localDate);
}