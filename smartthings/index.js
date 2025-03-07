const {SmartThingsClient, BearerTokenAuthenticator, DeviceListOptions} = require('@smartthings/core-sdk')
const {SceneListOptions} = require("@smartthings/core-sdk/dist/endpoint/scenes");
const {RuleRequest} = require("@smartthings/core-sdk/dist/endpoint/rules");

const command = process.argv[2]
let token = process.env.SMARTTHINGS_TOKEN
let tokenType = "bearer"
if (token === undefined || token === "") {
    token = process.env.SMARTTHINGS_API_TOKEN
    tokenType = "token"
}

const st = new SmartThingsClient(new BearerTokenAuthenticator(token))

async function main() {
    const deviceId = process.env.DEVICE_ID
    const locationId = process.env.LOCATION_ID
    const ruleId = process.env.RULE_ID
    switch (command) {
        case "listDevices":
            await listDevices(st)
            break
        case "toggleDevice":
            const state = process.env.STATE
            const capability = process.env.CAPABILITY
            await toggleDevice(st, deviceId, capability, state)
            break
        case "getDeviceInfo":
            await getDeviceInfo(st, deviceId)
            break
        case "listLocations":
            await listLocations(st)
            break
        case "listRooms":
            await listRooms(st, locationId)
            break
        case "listRules":
            await listRules(st, locationId)
            break
        case "getRule":
            await getRule(st, ruleId, locationId)
            break
        case "deleteRule":
            await deleteRule(st, ruleId, locationId)
            break
        case "createRule":
            const ruleName = process.env.RULE_NAME
            const ruleActions = process.env.RULE_ACTIONS
            const ruleSequence = process.env.RULE_SEQUENCE
            await createRule(st, locationId, ruleName, ruleActions, ruleSequence)
            break
    }
}

async function listDevices(st) {
    const devices = await st.devices.list()
    console.log(`Found ${devices.length} devices`)
    for (device of devices) {
        let roomName;
        try {
            const room = await st.rooms.get(device.roomId, device.locationId)
            roomName = room.name
        } catch (error) {
            roomName = "None"
        }
        let deviceId = "None";
        if (device.deviceId !== "undefined") {
            deviceId = device.deviceId
        }
        console.log(`Device "${device.label}" in Room "${roomName}" with deviceID "${deviceId}"`)
    }
}

async function toggleDevice(st, deviceId, capability, state) {
    const command = {
        component: 'main', capability: capability, command: state
    }
    await st.devices.executeCommand(deviceId, command)
}

async function getDeviceInfo(st, deviceId) {
    const device = await st.devices.get(deviceId)
    console.log(`Device: ${device.label} (${deviceId})`)
    console.log(`Type: ${device.name}`)
    console.log(`LocationId: ${device.locationId}`)

    console.log(`Capabilities:`)
    for (capability of device.components[0].capabilities) {
        let value = "";
        const ignoreCapability = new Set(["mediaInputSource", "ocf", "firmwareUpdate", "codeLength", "refresh", "maxCodes", "maxCodeLength", "codeChanged", "minCodeLength", "codeReport", "scanCodes", "lockCodes"])
        if (ignoreCapability.has(capability.id)) {
            continue
        }
        let capabilityStatus;
        const retries = 3
        for (let attempt = 0; attempt < retries; attempt++) {
            try {
                capabilityStatus = await st.devices.getCapabilityStatus(deviceId, 'main', capability.id)
                break
            } catch (error) {
                if (error.response.status === 429) {
                    const retryAfter = getRetryTime(error);
                    await new Promise(res => setTimeout(res, retryAfter));
                } else {
                    console.error(`  ${capability.id} = Unable to retrieve Capability Status`)
                    continue
                }
            }
            if (attempt === retries - 1) {
                value = "Unable to retrieve Capability Status"
            }
        }

        const matchingCapability = new Set(["switch", "lock", "battery"])
        if (value === "") {
            if (capability.id === "switchLevel") {
                value = capabilityStatus['level'].value
            } else if (matchingCapability.has(capability.id)) {
                value = capabilityStatus[capability.id].value
            } else {
                value = JSON.stringify(capabilityStatus)
            }
        }
        console.log(`  ${capability.id} = ${value}`)
    }
}

function getRetryTime(error) {
    const headers = error.response?.headers || {};

    // x-ratelimit-reset provides a timestamp (in seconds since epoch)
    if (headers['x-ratelimit-reset']) {
        const resetTime = parseInt(headers['x-ratelimit-reset'], 10) * 1000;
        const currentTime = Date.now();
        return Math.max(resetTime - currentTime, 1000);
    }

    // retry-after gives the wait time directly in seconds
    if (headers['retry-after']) {
        return parseInt(headers['retry-after'], 10) * 1000;
    }

    return 5000;
}

async function listLocations(st) {
    const locations = await st.locations.list()
    if (locations.length > 0) {
        console.log(`Found ${locations.length} locations`)
        for (location of locations) {
            console.log(`"${location.name}" (${location.locationId})`)
        }
        return null
    }
    console.log(`Did not find any locations!`)
}

async function listRooms(st, locationId) {
    const rooms = await st.rooms.list(locationId)
    if (rooms.length > 0) {
        console.log(`Found ${rooms.length} rooms`)
        for (room of rooms) {
            console.log(`"${room.name}" (${room.roomId})`)
        }
    }
}

async function listRules(st, locationId) {
    const rules = await st.rules.list(locationId)
    if (rules !== undefined && rules.length > 0) {
        console.log(`Found ${rules.length} rooms`)
        for (rule of rules) {
            console.log(`"${rule.name}" (${rule.roomId})`)
        }
        return null
    }
    console.log(`Did not find any rules!`)
}

async function getRule(st, ruleId, locationId) {
    const rule = await st.rules.get(ruleId, locationId)
    console.log(`${rule.name} (${rule.id})`)
    console.log('Sequence: ', rule.sequence)
    console.log(`${rule.actions}`)
}

async function createRule(st, locationId, name, actions, sequence) {
    const ruleRequest = {
        name: name, actions: actions, sequence: {
            actions: sequence,
        }
    }
    try {
        const newRule = await st.rules.create(ruleRequest, locationId)
        console.log(newRule)
    } catch (e) {
        console.log(e)
    }
}

async function deleteRule(st, locationId, ruleId) {
    try {
        await st.rules.delete(locationId, ruleId)
        console.log(`Deleted rule with Id ${ruleId}`)
    } catch (e) {
        console.log(e)
    }
}

main()
