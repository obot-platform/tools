from tools.helper import tool_registry, api_key_headers, remove_images_from_profile, remove_images_from_search_result
import os
import requests


@tool_registry.decorator("SearchCompany")
def search_company() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/company'

    ai_params = ["COUNTRY", "REGION", "CITY", "TYPE",
                 "NAME", "INDUSTRY", "EMPLOYEE_COUNT_MIN", "EMPLOYEE_COUNT_MAX", "DESCRIPTION",
                 "FOUNDED_AFTER_YEAR", "FOUNDED_BEFORE_YEAR"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["enrich_profiles"] = "enrich"
    params["page_size"] = os.getenv("PAGE_SIZE") if os.getenv("PAGE_SIZE") else 1  # default to 1 result

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)

    return remove_images_from_search_result(response.json())


@tool_registry.decorator("SearchPerson")
def search_person() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/search/person'

    ai_params = ["COUNTRY", "FIRST_NAME", "LAST_NAME", "EDUCATION_FIELD_OF_STUDY", "EDUCATION_DEGREE_NAME",
                 "EDUCATION_SCHOOL_NAME", "CURRENT_ROLE_TITLE", "PAST_ROLE_TITLE", "CURRENT_ROLE_BEFORE",
                 "CURRENT_ROLE_AFTER", "CURRENT_COMPANY_NAME", "PAST_COMPANY_NAME", "LINKEDIN_GROUPS",
                 "LANGUAGES", "REGION", "CITY", "HEADLINE", "SUMMARY", "INDUSTRIES", "INTERESTS", "SKILLS",
                 "CURRENT_COMPANY_COUNTRY", "CURRENT_COMPANY_REGION", "CURRENT_COMPANY_CITY", "CURRENT_COMPANY_TYPE"]

    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}
    params["enrich_profiles"] = "enrich"
    params["page_size"] = os.getenv("PAGE_SIZE") if os.getenv("PAGE_SIZE") else 1  # default to 1 result

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)

    return remove_images_from_search_result(response.json())


@tool_registry.decorator("SearchJob")
def search_job() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/linkedin/company/job'
    ai_params = ["JOB_TYPE", "EXPERIENCE_LEVEL", "WHEN", "FLEXIBILITY", "GEO_ID", "KEYWORD", "SEARCH_ID"]
    params = {k.lower(): os.getenv(k) for k in ai_params if os.getenv(k) is not None}

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)

    return response.json()


@tool_registry.decorator("SearchRole")
def search_role() -> dict:
    api_endpoint = 'https://nubela.co/proxycurl/api/find/company/role/'
    role = os.getenv("ROLE")
    company = os.getenv("COMPANY")

    params = {
        'role': role,
        'company_name': company,
        'enrich_profile': 'enrich'
    }

    result = requests.get(api_endpoint, params=params, headers=api_key_headers).json()
    result["profile"] = remove_images_from_profile(result["profile"])

    return result
