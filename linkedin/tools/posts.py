from tools.helper import ACCESS_TOKEN, load_from_gptscript_workspace
from linkedin_api.clients.restli.client import RestliClient
from tools.users import get_user
import os
import requests


async def create_post(client: RestliClient) -> dict:
    r = get_user(client)
    user_id = r["sub"]
    content = os.getenv("CONTENT")

    share_media_category = os.getenv("SHARE_MEDIA_CATEGORY", "NONE")
    share_media_category_enum = ["NONE", "IMAGE", "VIDEO", "ARTICLE"]
    if share_media_category not in share_media_category_enum:
        raise ValueError(
            f"Error: invalid SHARE_MEDIA_CATEGORY value: {share_media_category}. Must be one of {share_media_category_enum}"
        )

    if share_media_category == "NONE" and (content == ""):
        raise ValueError(
            "Error: CONTENT can't be empty if SHARE_MEDIA_CATEGORY is NONE."
        )

    visibility = os.getenv("VISIBILITY", "PUBLIC")
    visibility_enum = ["PUBLIC", "CONNECTIONS"]
    if visibility not in visibility_enum:
        raise ValueError(
            f"Error: invalid VISIBILITY value: {visibility}. Must be one of {visibility_enum}"
        )

    share_media_category_value = share_media_category
    if share_media_category == "VIDEO":
        share_media_category_value = "IMAGE"
    payload = {
        "author": f"urn:li:person:{user_id}",
        "lifecycleState": "PUBLISHED",  # it's always PUBLISHED
        "specificContent": {
            "com.linkedin.ugc.ShareContent": {
                "shareCommentary": {"text": content},
                "shareMediaCategory": share_media_category_value,
            }
        },
        "visibility": {"com.linkedin.ugc.MemberNetworkVisibility": visibility},
    }

    if share_media_category != "NONE":
        title = os.getenv("SHARE_MEDIA_TITLE")
        if title is None or title == "":
            raise ValueError("Error: to create a media, SHARE_MEDIA_TITLE is required.")
        description = os.getenv("SHARE_MEDIA_DESC")
        if description is None or description == "":
            raise ValueError("Error: to create a media, SHARE_MEDIA_DESC is required.")

        if share_media_category == "ARTICLE":
            if "SHARE_MEDIA_ORIGINAL_URL" not in os.environ:
                raise ValueError(
                    "Error: to create an article, SHARE_MEDIA_ORIGINAL_URL is required."
                )
            original_url = os.getenv("SHARE_MEDIA_ORIGINAL_URL")
            media_payload = [
                {
                    "status": "READY",
                    "description": {"text": description},
                    "originalUrl": original_url,
                    "title": {"text": title},
                }
            ]
        elif share_media_category == "IMAGE" or share_media_category == "VIDEO":
            file_path = os.getenv("SHARE_MEDIA_FILE_PATH")
            if file_path is None or file_path == "":
                raise ValueError(
                    "Error: to create an image or video, SHARE_MEDIA_FILE_PATH is required."
                )
            upload_url, asset = _register_upload(client, share_media_category)
            await _upload_registered_file(file_path, upload_url)
            media_payload = [
                {
                    "status": "READY",
                    "description": {"text": description},
                    "media": asset,
                    "title": {"text": title},
                }
            ]

        payload["specificContent"]["com.linkedin.ugc.ShareContent"][
            "media"
        ] = media_payload

    response = client.create(
        resource_path="/ugcPosts",
        entity=payload,
        access_token=ACCESS_TOKEN,
    )
    if response.status_code != 201:
        raise ValueError(
            f"Error: failed to create post. Status code: {response.status_code}. Response: {response.entity}"
        )

    response.entity["link"] = (
        f"https://www.linkedin.com/feed/update/{response.entity['id']}"  # craft the link manually
    )
    return response.entity


def _register_upload(client: RestliClient, category: str) -> tuple[str, dict]:
    url = "https://api.linkedin.com/v2/assets?action=registerUpload"

    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
        "Content-Type": "application/json",
    }

    if category == "IMAGE":
        recipe = "urn:li:digitalmediaRecipe:feedshare-image"
    elif category == "VIDEO":
        recipe = "urn:li:digitalmediaRecipe:feedshare-video"
    else:
        raise ValueError(f"Error: invalid SHARE_MEDIA_CATEGORY value: {category}.")

    user_id = get_user(client)["sub"]
    data = {
        "registerUploadRequest": {
            "recipes": [recipe],
            "owner": f"urn:li:person:{user_id}",
            "serviceRelationships": [
                {
                    "relationshipType": "OWNER",
                    "identifier": "urn:li:userGeneratedContent",
                }
            ],
        }
    }

    response = requests.post(url, headers=headers, json=data)

    if response.status_code == 200:
        response_json = response.json()
        return (
            response_json["value"]["uploadMechanism"][
                "com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"
            ]["uploadUrl"],
            response_json["value"]["asset"],
        )
    else:
        raise Exception(
            f"Error: failed to register upload. Status code: {response.status_code}. Response: {response.text}"
        )


async def _upload_registered_file(file_path, upload_url) -> int:
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
        "Content-Type": "application/octet-stream",
    }

    data = await load_from_gptscript_workspace(file_path)

    response = requests.post(upload_url, headers=headers, data=data)

    if response.status_code != 201:
        raise ValueError(
            f"Error: failed to upload file. Status code: {response.status_code}. Response: {response.text}. Please try to upload the file again."
        )
    return response.status_code
