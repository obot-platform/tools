from tools.helper import tool_registry, api_key_headers
import os
import requests
from tools.api import company_profile_from_url, user_profile_from_url


@tool_registry.decorator("SearchCompany")
def search_company() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/company'

    ai_params = ["COUNTRY", "REGION", "CITY", "TYPE", "FOLLOWER_COUNT_MIN", "FOLLOWER_COUNT_MAX",
                 "NAME", "INDUSTRY", "EMPLOYEE_COUNT_MIN", "EMPLOYEE_COUNT_MAX", "DESCRIPTION",
                 "FOUNDED_AFTER_YEAR", "FOUNDED_BEFORE_YEAR"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1  # limit results to 1

    search_results = requests.get(api_endpoint, params=params, headers=api_key_headers).json()

    if search_results.get("results"):
        company_url = search_results["results"][0].get("linkedin_profile_url")
        print("Found URL: ", company_url)

        if company_url:
            return company_profile_from_url(company_url).json()

    return {"Err": "Could not find company profile with given search criteria"}


@tool_registry.decorator("SearchUser")
def search_user() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/person'

    ai_params = ["COUNTRY", "FIRST_NAME", "LAST_NAME", "EDUCATION_FIELD_OF_STUDY", "EDUCATION_DEGREE_NAME",
                 "EDUCATION_SCHOOL_NAME", "CURRENT_ROLE_TITLE", "PAST_ROLE_TITLE", "CURRENT_ROLE_BEFORE",
                 "CURRENT_ROLE_AFTER", "CURRENT_JOB_DESCRIPTION", "PAST_JOB_DESCRIPTION", "CURRENT_COMPANY_NAME",
                 "PAST_COMPANY_NAME", "LINKEDIN_GROUPS", "LANGUAGES",
                 "REGION", "CITY", "HEADLINE", "SUMMARY", "INDUSTRIES", "INTERESTS", "SKILLS",
                 "CURRENT_COMPANY_COUNTRY", "CURRENT_COMPANY_REGION", "CURRENT_COMPANY_CITY", "CURRENT_COMPANY_TYPE"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1  # limit results to 1

    search_results = requests.get(api_endpoint, params=params, headers=api_key_headers).json()

    if search_results.get("results"):
        user_url = search_results["results"][0].get("linkedin_profile_url")
        print("Found URL: ", user_url)

        if user_url:
            return user_profile_from_url(user_url).json()

    return {"Err": "Could not find user profile with given search criteria"}


@tool_registry.decorator("SearchJob")
def search_job() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/linkedin/company/job'

    ai_params = ["JOB_TYPE", "EXPERIENCE_LEVEL", "WHEN", "FLEXIBILITY", "GEO_ID", "KEYWORD", "SEARCH_ID"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1  # limit results to 1

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)

    return response.json()
