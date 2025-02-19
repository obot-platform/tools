async function getMe(client) {
    const resp = await client.get('/users/me');
    return resp.data;
}

module.exports = {
    getMe
} 