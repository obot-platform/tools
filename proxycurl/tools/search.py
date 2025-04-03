from tools.helper import tool_registry, format_url, api_key_headers
import os
import requests


@tool_registry.decorator("SearchCompany")
def search_company():
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/company'

    ai_params = ["COUNTRY", "REGION", "CITY", "TYPE", "FOLLOWER_COUNT_MIN", "FOLLOWER_COUNT_MAX",
                 "NAME", "INDUSTRY", "EMPLOYEE_COUNT_MIN", "EMPLOYEE_COUNT_MAX", "DESCRIPTION",
                 "FOUNDED_AFTER_YEAR", "FOUNDED_BEFORE_YEAR", "FUNDING_AMOUNT_MIN", "FUNDING_AMOUNT_MAX",
                 "FUNDING_RAISED_AFTER", "FUNDING_RAISED_BEFORE", "PUBLIC_IDENTIFIER_IN_LIST",
                 "PUBLIC_IDENTIFIER_NOT_IN_LIST"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())


@tool_registry.decorator("SearchUser")
def search_user():
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/person'

    ai_params = ["COUNTRY", "FIRST_NAME", "LAST_NAME", "EDUCATION_FIELD_OF_STUDY", "EDUCATION_DEGREE_NAME",
                 "EDUCATION_SCHOOL_NAME", "CURRENT_ROLE_TITLE", "PAST_ROLE_TITLE", "CURRENT_ROLE_BEFORE",
                 "CURRENT_ROLE_AFTER", "CURRENT_JOB_DESCRIPTION", "PAST_JOB_DESCRIPTION", "CURRENT_COMPANY_NAME",
                 "PAST_COMPANY_NAME", "LINKEDIN_GROUPS", "LANGUAGES",
                 "REGION", "CITY", "HEADLINE", "SUMMARY", "INDUSTRIES", "INTERESTS", "SKILLS",
                 "CURRENT_COMPANY_COUNTRY", "CURRENT_COMPANY_REGION", "CURRENT_COMPANY_CITY", "CURRENT_COMPANY_TYPE",
                 "CURRENT_COMPANY_FOLLOWER_COUNT_MIN", "CURRENT_COMPANY_FOLLOWER_COUNT_MAX", "CURRENT_COMPANY_INDUSTRY",
                 "CURRENT_COMPANY_EMPLOYEE_COUNT_MIN", "CURRENT_COMPANY_EMPLOYEE_COUNT_MAX",
                 "CURRENT_COMPANY_DESCRIPTION", "CURRENT_COMPANY_FOUNDED_AFTER_YEAR",
                 "CURRENT_COMPANY_FOUNDED_BEFORE_YEAR", "CURRENT_COMPANY_FUNDING_AMOUNT_MIN",
                 "CURRENT_COMPANY_FUNDING_AMOUNT_MAX", "CURRENT_COMPANY_FUNDING_RAISED_AFTER",
                 "CURRENT_COMPANY_FUNDING_RAISED_BEFORE", "PUBLIC_IDENTIFIER_IN_LIST", "PUBLIC_IDENTIFIER_NOT_IN_LIST",
                 "FOLLOWER_COUNT_MIN", "FOLLOWER_COUNT_MAX"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())


@tool_registry.decorator("SearchJob")
def search_job():
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/linkedin/company/job'

    ai_params = ["JOB_TYPE", "EXPERIENCE_LEVEL", "WHEN", "FLEXIBILITY", "GEO_ID", "KEYWORD", "SEARCH_ID"]
    
    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["page_size"] = 1

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())
