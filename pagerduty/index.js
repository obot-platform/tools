const { api } = require('@pagerduty/pdjs');
const {
    listIncidents,
    getIncident,
    updateIncidentStatus,
    addIncidentNote,
    listIncidentNotes,
    listIncidentAlerts,
} = require('./src/incidents.js');
const { getMe } = require('./src/users.js');

if (process.argv.length !== 3) {
    console.error('Usage: node index.js <command>')
    process.exit(1)
}

const command = process.argv[2]
let token = process.env.PAGERDUTY_BEARER_TOKEN
let tokenType = "bearer"
if (token === undefined || token === "") {
    token = process.env.PAGERDUTY_API_TOKEN
    tokenType = "token"
}

if (token === undefined || token === "") {
    console.error('Please set the PAGERDUTY_BEARER_TOKEN or PAGERDUTY_API_TOKEN environment variable')
    process.exit(1)
}

const pd = api({ token: token, tokenType: tokenType });

async function main() {
    try {
        let incidentId = "";
        switch (command) {
            case "listIncidents":
                const things = await listIncidents(pd);
                console.log("INCIDENTS: ", things);
                break
            case "getIncident":
                incidentId = getIncidentId()
                const incident = await getIncident(pd, incidentId);
                console.log("INCIDENT: ", incident);
                break
            case "acknowledgeIncident":
                incidentId = getIncidentId();
                if (incidentId === undefined || incidentId === "") {
                    console.error('Please set the INCIDENT_ID environment variable')
                    process.exit(1)
                }
                const ackIncident = await updateIncidentStatus(pd, incidentId, 'acknowledged');
                console.log("ACKNOWLEDGED INCIDENT: ", ackIncident);
                break
            case "resolveIncident":
                incidentId = getIncidentId();
                const resolvedIncident = await updateIncidentStatus(pd, incidentId, 'resolved');
                console.log("RESOLVED INCIDENT: ", resolvedIncident);
                break
            case "addIncidentNote":
                incidentId = getIncidentId();
                const note = process.env.NOTE
                if (note === undefined || note === "") {
                    console.error('Please set the NOTE_CONTENT environment variable')
                    process.exit(1)
                }
                const noteResp = await addIncidentNote(pd, incidentId, note);
                console.log("NOTE ADDED: ", noteResp);
                break
            case "listIncidentNotes":
                incidentId = getIncidentId();
                const noteList = await listIncidentNotes(pd, incidentId);
                console.log("NOTES: ", noteList);
                break
            case "listIncidentAlerts":
                incidentId = getIncidentId();
                const alerts = await listIncidentAlerts(pd, incidentId);
                console.log(JSON.stringify(alerts.alerts));
                break
            case "getMe":
                const user = await getMe(pd);
                console.log("USER: ", user);
                break;
            default:
                console.log(`Unknown command: ${command}`)
                process.exit(1)
        }
    } catch (error) {
        // We use console.log instead of console.error here so that it goes to stdout
        console.log("Got the following error: ", error);
        process.exit(1)
    }
}

function getIncidentId() {
    incidentId = process.env.INCIDENT_ID
    if (incidentId === undefined || incidentId === "") {
        console.error('Please set the INCIDENT_ID environment variable')
        process.exit(1)
    }
    return incidentId
}

main()
