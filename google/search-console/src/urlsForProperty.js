import fetch from 'node-fetch';

export async function getUrlsForProperty(property, oauthToken) {
    const apiUrl = `https://www.googleapis.com/webmasters/v3/sites/${encodeURIComponent(property)}/searchAnalytics/query`;

    const payload = {
        startDate: getThreeMonthsAgo(),
        endDate: getToday(),
        dimensions: ["page"],
        rowLimit: 100,
        orderBy: [{ fieldName: "clicks", sortOrder: "DESCENDING" }]
    };

    const headers = {
        "Authorization": `Bearer ${oauthToken}`,
        "Content-Type": "application/json"
    };

    try {
        const response = await fetch(apiUrl, {
            method: "POST",
            headers: headers,
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch data: ${response.status} - ${response.statusText}`);
        }

        const data = await response.json();
        return data.rows ? data.rows.map(row => row.keys[0]) : [];
    } catch (error) {
        console.error(`Error: ${error.message}`);
        return [];
    }
}

function getToday() {
    const today = new Date();
    return today.toISOString().split("T")[0];
}

function getThreeMonthsAgo() {
    const date = new Date();
    date.setMonth(date.getMonth() - 3);
    return date.toISOString().split("T")[0];
}