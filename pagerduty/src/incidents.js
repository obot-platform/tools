function getEmail() {
    return process.env.USER_EMAIL;
}

async function listIncidents(client) {
    const resp = await client.get('/incidents');
    return resp.resource;
}

async function getIncident(client, id) {
    const resp = await client.get(`/incidents/${id}`);
    return resp.data;
}

async function updateIncidentStatus(client, id, status) {
    const orig = await getIncident(client, id);
    console.log("Original: ", orig.incident.id);
    const resp = await client.put(`/incidents/${orig.incident.id}`, {
        headers: {
            Accept: 'application/vnd.pagerduty+json;version=2',
            From: getEmail(),
        },
        data: {
            incident: {
                type: 'incident_reference',
                status: status
            }
        },
    });
    return resp.data;
}

async function listIncidentNotes(client, id) {
    const resp = await client.get(`/incidents/${id}/notes`);
    return JSON.stringify(resp.data.notes);
}

async function addIncidentNote(client, id, contents) {
    const resp = await client.post(`/incidents/${id}/notes`, {
        headers: {
            From: getEmail(),
        },
        data: {
            note: {
                content: contents
            }
        }
    });
    return resp.data;
}

async function listIncidentAlerts(client, id) {
    const resp = await client.get(`/incidents/${id}/alerts`);
    return resp.data;
}

module.exports = {
    listIncidents,
    getIncident,
    updateIncidentStatus,
    addIncidentNote,
    listIncidentNotes,
    listIncidentAlerts
}