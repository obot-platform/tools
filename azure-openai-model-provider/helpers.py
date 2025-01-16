import os
from typing import Tuple

from azure.identity import (
    CredentialUnavailableError,
    DefaultAzureCredential,
    get_bearer_token_provider,
)
from azure.mgmt.cognitiveservices import CognitiveServicesManagementClient
from openai import AsyncAzureOpenAI


def list_openai(client: CognitiveServicesManagementClient, resource_group: str):
    accounts = client.accounts.list_by_resource_group(
        resource_group_name=resource_group, api_version="2023-05-01"
    )
    deployments = []
    for account in accounts:
        if account.kind == "OpenAI":
            deployments.extend(
                client.deployments.list(
                    resource_group_name=resource_group,
                    account_name=account.name,
                    api_version="2023-05-01",
                )
            )

    return deployments


def configure() -> Tuple[AsyncAzureOpenAI, CognitiveServicesManagementClient, str, str]:
    api_version = os.environ.get(
        "OBOT_AZURE_OPENAI_MODEL_PROVIDER_API_VERSION", "2024-10-21"
    )
    subscription_id = os.environ.get("OBOT_AZURE_OPENAI_MODEL_PROVIDER_SUBSCRIPTION_ID")
    if subscription_id is None:
        raise Exception("Azure subscription ID was not configured")

    resource_group = os.environ.get("OBOT_AZURE_OPENAI_MODEL_PROVIDER_RESOURCE_GROUP")
    if resource_group is None:
        raise Exception("Azure Resource Group was not configured")

    endpoint = os.environ.get("OBOT_AZURE_OPENAI_MODEL_PROVIDER_ENDPOINT")
    if endpoint is None:
        raise Exception("Azure model endpoint was not configured")

    os.environ["AZURE_CLIENT_ID"] = os.environ.get(
        "OBOT_AZURE_OPENAI_MODEL_PROVIDER_CLIENT_ID"
    )
    os.environ["AZURE_TENANT_ID"] = os.environ.get(
        "OBOT_AZURE_OPENAI_MODEL_PROVIDER_TENANT_ID"
    )
    os.environ["AZURE_CLIENT_SECRET"] = os.environ.get(
        "OBOT_AZURE_OPENAI_MODEL_PROVIDER_CLIENT_SECRET"
    )

    try:
        token_provider = get_bearer_token_provider(
            DefaultAzureCredential(), "https://cognitiveservices.azure.com/.default"
        )

        azure_client = AsyncAzureOpenAI(
            api_version=api_version,
            azure_endpoint=endpoint,
            azure_ad_token_provider=token_provider,
        )

        cognitive_services_client = CognitiveServicesManagementClient(
            credential=DefaultAzureCredential(), subscription_id=subscription_id
        )

    except CredentialUnavailableError:
        raise Exception("Could not get Azure credentials")
    except Exception as e:
        raise e

    return azure_client, cognitive_services_client, subscription_id, resource_group
