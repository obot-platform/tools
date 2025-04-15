from tools.helper import api_key_headers
import requests


def company_profile_from_url(url: str) -> requests.Response:
    api_endpoint = 'https://nubela.co/proxycurl/api/linkedin/company'
    params = {
        'url': url,
        'categories': 'exclude',
        'funding_data': 'exclude',
        'exit_data': 'exclude',
        'acquisitions': 'exclude',
        'extra': 'exclude',
        'use_cache': 'if-present',
        'fallback_to_cache': 'on-error',
    }

    return requests.get(api_endpoint, params=params, headers=api_key_headers)


def person_profile_from_url(url: str) -> requests.Response:
    api_endpoint = 'https://nubela.co/proxycurl/api/v2/linkedin'
    params = {
        'url': url,
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

    return requests.get(api_endpoint, params=params, headers=api_key_headers)


def school_profile_from_url(url: str) -> requests.Response:
    api_endpoint = 'https://nubela.co/proxycurl/api/linkedin/school'
    params = {
        'url': url,
        'use_cache': 'if-present',
    }

    return requests.get(api_endpoint, params=params, headers=api_key_headers)
