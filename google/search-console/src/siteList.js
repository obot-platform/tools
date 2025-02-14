export async function fetchGSCProperties(oauthToken) {
    // Fetch the list of GSC properties
    const sites = await getSitesListFromGSC(oauthToken);

    if (!sites || sites.length === 0) {
        console.log('No GSC properties found. Please ensure you have access to GSC properties.');
        return;
    }

    // Display the properties (for demonstration, we'll log them to the console)
    console.log('GSC Properties:');
    sites.forEach(site => {
        console.log(`- ${site.siteUrl}`);
    });
}

async function getSitesListFromGSC(oauthToken) {
    const url = 'https://www.googleapis.com/webmasters/v3/sites';

    const headers = {
        'Authorization': `Bearer ${oauthToken}`,
        'Content-Type': 'application/json'
    };

    try {
        const response = await fetch(url, {
            method: 'GET',
            headers: headers
        });

        if (!response.ok) {
            throw new Error(`Error fetching data: ${response.status} - ${response.statusText}`);
        }

        const data = await response.json();
        return data.siteEntry || [];
    } catch (error) {
        console.error(`Error: ${error.message}`);
        return [];
    }
}