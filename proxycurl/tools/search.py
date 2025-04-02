from tools.helper import tool_registry, format_url, api_key_headers
import os
import requests


@tool_registry.register("SearchCompany")
def search_company():
    ...


@tool_registry.register("SearchUser")
def search_user():
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/person'
    ...


@tool_registry.register("SearchJob")
def search_job():
    ...
