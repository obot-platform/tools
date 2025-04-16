import requests


# Get Keyword Suggestions from Google
def get_google_suggestions(seed_keyword, lang="en", country="us"):
    url = "https://suggestqueries.google.com/complete/search"
    params = {"client": "firefox", "hl": lang, "gl": country, "q": seed_keyword}
    response = requests.get(url, params=params)
    response.raise_for_status()
    return response.json()[1]


def google_search_suggestions_tool(seed_keyword, num_suggestions=5):

    suggestions = get_google_suggestions(seed_keyword)

    return suggestions[:num_suggestions]


