from tools.helper import WORDPRESS_API_URL, tool_registry


@tool_registry.register("GetSiteSettings")
def get_site_settings(client):
    url = f"{WORDPRESS_API_URL}/settings"
    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return response.json()
    else:
        print(f"Error: {response.status_code}, {response.text}")
