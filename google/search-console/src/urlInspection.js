// urlInspection.js
import fetch from 'node-fetch';

/**
 * Inspects a URL using the Google Search Console URL Inspection API.
 * @param {string} property - The GSC property URL.
 * @param {string} url - The URL to inspect.
 * @param {string} oauthToken - The OAuth token for authentication.
 * @returns {Promise<Object>} - The inspection result.
 */
export async function inspectUrl(property, url, oauthToken) {
    const apiUrl = 'https://searchconsole.googleapis.com/v1/urlInspection/index:inspect';

    const payload = {
        siteUrl: property,
        inspectionUrl: url,
        languageCode: 'en-US'
    };

    const headers = {
        Authorization: `Bearer ${oauthToken}`,
        'Content-Type': 'application/json'
    };

    const options = {
        method: 'POST',
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
            const jsonResponse = JSON.parse(responseText);
            if (jsonResponse && jsonResponse.inspectionResult) {
                return jsonResponse.inspectionResult;
            } else {
                console.error(`Unexpected API Response Structure: ${responseText}`);
                throw new Error('Unexpected API response format. "inspectionResult" field is missing.');
            }
        } else {
            console.error(`Failed API Call: ${responseText}`);
            throw new Error(`Failed to inspect URL. Response Code: ${responseCode}. Response: ${responseText}`);
        }
    } catch (error) {
        console.error(`Error during API call: ${error.message}`);
        throw new Error(`Error inspecting URL: ${error.message}`);
    }
}
