from tools.helper import tool_registry, format_url, remove_images_from_profile
from tools.search import search_company
from tools.api import school_profile_from_url, person_profile_from_url, company_profile_from_url
import os


@tool_registry.decorator("GetCompanyProfile")
def get_company_profile():
    env_url = os.getenv("URL")
    company_name = os.getenv("COMPANY")
    url = env_url if env_url else f'https://www.linkedin.com/company/{format_url(company_name)}/'

    response = company_profile_from_url(url)

    # Search for URL with company name
    if response.status_code == 404 and company_name:
        print("Initial request 404. Searching for profile URL...")

        os.environ["NAME"] = company_name
        os.environ["PAGE_SIZE"] = "1"
        return search_company()

    return remove_images_from_profile(response.json())


@tool_registry.decorator("GetSchoolProfile")
def get_school_profile():
    env_url = os.getenv("URL")
    school_name = os.getenv("SCHOOL")
    url = env_url if env_url else f'https://www.linkedin.com/school/{format_url(school_name)}/'

    return remove_images_from_profile(school_profile_from_url(url).json())


@tool_registry.decorator("GetPersonProfile")
def get_person_profile():
    env_url = os.getenv("URL")

    response = person_profile_from_url(env_url)

    return remove_images_from_profile(response.json())
