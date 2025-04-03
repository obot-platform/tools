from tools.helper import tool_registry, format_url, api_key_headers
import os
import requests

@tool_registry.decorator("GetCompanyProfile")
def get_company_profile():
    company_name = os.getenv("COMPANY")

    api_endpoint = 'https://nubela.co/proxycurl/api/linkedin/company'
    params = {
        'url': f'https://www.linkedin.com/company/{format_url(company_name)}/',
        'categories': 'exclude',
        'funding_data': 'exclude',
        'exit_data': 'exclude',
        'acquisitions': 'exclude',
        'extra': 'exclude',
        'use_cache': 'if-present',
        'fallback_to_cache': 'on-error',
    }

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())


@tool_registry.decorator("GetSchoolProfile")
def get_school_profile():
    school_name = os.getenv("SCHOOL")

    api_endpoint = 'https://nubela.co/proxycurl/api/linkedin/school'
    params = {
        'url': f'https://www.linkedin.com/school/{format_url(school_name)}/',
        'use_cache': 'if-present',
    }

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())


@tool_registry.decorator("GetUserProfile")
def get_user_profile():
    user = os.getenv("USER")

    api_endpoint = 'https://nubela.co/proxycurl/api/v2/linkedin'
    params = {
        'linkedin_profile_url': f'https://linkedin.com/in/{format_url(user)}/',
        'extra': 'exclude',
        'github_profile_id': 'exclude',
        'facebook_profile_id': 'exclude',
        'twitter_profile_id': 'exclude',
        'personal_contact_number': 'exclude',
        'personal_email': 'exclude',
        'inferred_salary': 'exclude',
        'skills': 'exclude',
        'use_cache': 'if-present',
        'fallback_to_cache': 'on-error',
    }

    response = requests.get(api_endpoint, params=params, headers=api_key_headers)
    print(response.json())

