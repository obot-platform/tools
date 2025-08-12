import pytest
from unittest.mock import patch, MagicMock, Mock
from fastmcp import Client
from fastmcp.exceptions import ToolError
from googleapiclient.errors import HttpError
import json

from obot_gmail_mcp.server import mcp

# Configure pytest for async support
pytestmark = pytest.mark.asyncio


# Fixtures
@pytest.fixture
def mock_service():
    """Create a mock Google Gmail service with proper chain structure"""
    service = MagicMock()

    # Mock the chain: service.users().messages().list().execute()
    mock_users = MagicMock()
    mock_messages = MagicMock()
    mock_list = MagicMock()
    mock_get = MagicMock()
    mock_send = MagicMock()
    mock_trash = MagicMock()
    mock_modify = MagicMock()

    # Set up the chain
    service.users.return_value = mock_users
    mock_users.messages.return_value = mock_messages
    mock_messages.list.return_value = mock_list
    mock_messages.get.return_value = mock_get
    mock_messages.send.return_value = mock_send
    mock_messages.trash.return_value = mock_trash
    mock_messages.modify.return_value = mock_modify

    # Mock drafts chain
    mock_drafts = MagicMock()
    mock_users.drafts.return_value = mock_drafts
    mock_drafts_delete = MagicMock()
    mock_drafts_send = MagicMock()
    mock_drafts.delete.return_value = mock_drafts_delete
    mock_drafts.send.return_value = mock_drafts_send

    # Mock getProfile chain
    mock_get_profile = MagicMock()
    mock_users.getProfile.return_value = mock_get_profile

    return service


@pytest.fixture
def mock_email_data():
    """Sample email data for testing"""
    return [
        {
            "id": "test_email_id",
            "snippet": "Test email snippet",
            "labelIds": ["INBOX"],
            "payload": {
                "headers": [
                    {"name": "Subject", "value": "Test Subject"},
                    {"name": "From", "value": "test@example.com"},
                    {"name": "To", "value": "user@example.com"},
                    {"name": "Date", "value": "Mon, 1 Jan 2024 10:00:00 -0500"},
                ]
            },
        }
    ]


@pytest.fixture
def mock_label_data():
    """Sample label data for testing"""
    return [
        {
            "id": "test_label_id",
            "name": "Test Label",
            "type": "user",
            "labelListVisibility": "labelShow",
            "messageListVisibility": "show",
        }
    ]


@pytest.fixture
def mock_draft_data():
    """Sample draft data for testing"""
    return [
        {
            "id": "test_draft_id",
            "message": {"id": "test_message_id", "snippet": "Test draft snippet"},
        }
    ]


@pytest.fixture
def mock_profile_data():
    """Sample profile data for testing"""
    return {
        "emailAddress": "user@example.com",
        "messagesTotal": 100,
        "threadsTotal": 50,
    }


# Integration Tests - Testing the MCP Server
class TestMCPServer:
    """Test the FastMCP server integration"""

    async def test_list_tools(self):
        """Test that all tools are properly registered"""
        # Test that tools are registered in the MCP server
        async with Client(mcp) as client:
            tools = await client.list_tools()
            assert len(tools) == 16

        # Verify expected tools are present
        tool_names = [tool.name for tool in tools]
        expected_tools = [
            "list_emails",
            "list_drafts",
            "list_labels",
            "create_label",
            "update_label",
            "delete_label",
            "modify_message_labels",
            "get_current_email_address",
            "create_draft",
            "delete_draft",
            "delete_email",
            "read_email",
            "send_draft",
            "send_email",
            "update_draft",
            "list_attachments",
        ]
        for tool in expected_tools:
            assert tool in tool_names


# Unit Tests - Testing Individual Functions
class TestEmailFunctions:
    """Test individual email-related functions"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_messages")
    @patch("obot_gmail_mcp.server.message_to_string")
    async def test_list_emails_success(
        self,
        mock_message_to_string,
        mock_list_messages,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_email_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_messages.return_value = mock_email_data
        mock_message_to_string.return_value = ("mockid", "Formatted Email")

        # Set up the mock service chain to avoid credential errors
        mock_service.users().messages().list().execute.return_value = {
            "messages": mock_email_data
        }

        async with Client(mcp) as client:
            result = await client.call_tool(name="list_emails", arguments={})
            res_json = result[0].text
            assert res_json == "Formatted Email"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_messages")
    async def test_list_emails_no_results(
        self, mock_list_messages, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_messages.return_value = []

        # Set up the mock service chain to avoid credential errors
        mock_service.users().messages().list().execute.return_value = {"messages": []}

        async with Client(mcp) as client:
            result = await client.call_tool(name="list_emails", arguments={})
            res_json = result[0].text
            assert res_json == "No emails found"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_messages")
    async def test_list_emails_with_query_containing_date_filters(
        self, mock_list_messages, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Set up the mock service chain to avoid credential errors
        mock_service.users().messages().list().execute.return_value = {"messages": []}

        with pytest.raises(
            ToolError, match="Please use the parameters `after` and `before`"
        ):
            async with Client(mcp) as client:
                await client.call_tool(
                    name="list_emails", arguments={"query": "after:2024-01-01"}
                )

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.fetch_email_or_draft")
    @patch("obot_gmail_mcp.server.get_email_body")
    @patch("obot_gmail_mcp.server.has_attachment")
    @patch("obot_gmail_mcp.server.format_message_metadata")
    async def test_read_email_success(
        self,
        mock_format_metadata,
        mock_has_attachment,
        mock_get_body,
        mock_fetch_email,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_fetch_email.return_value = {"id": "test_id"}
        mock_get_body.return_value = "Test email body"
        mock_has_attachment.return_value = False
        mock_format_metadata.return_value = (None, "Test metadata")

        # Set up the mock service chain to avoid credential errors
        mock_service.users().messages().get().execute.return_value = {"id": "test_id"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="read_email", arguments={"email_id": "test_id"}
            )
            res_json = json.loads(result[0].text)
            assert res_json["body"] == "Test email body"
            assert res_json["metadata"] == "Test metadata"
            assert res_json["has_attachment"] == False

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_read_email_missing_params(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Set up the mock service chain to avoid credential errors
        mock_service.users().messages().get().execute.return_value = {"id": "test_id"}

        with pytest.raises(
            ToolError, match="Either email_id or email_subject must be set"
        ):
            async with Client(mcp) as client:
                await client.call_tool(name="read_email", arguments={})

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_delete_email_success(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Set up the mock service chain properly
        mock_service.users().messages().trash().execute.return_value = {}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_email", arguments={"email_id": "test_id"}
            )
            res_json = result[0].text
            assert "deleted successfully" in res_json

        mock_service.users().messages().trash.assert_called_with(
            userId="me", id="test_id"
        )

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.create_message_data")
    async def test_send_email_success(
        self, mock_create_message, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_message.return_value = {"raw": "encoded_message"}

        # Set up the mock service chain properly
        mock_service.users().messages().send().execute.return_value = {"id": "sent_id"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="send_email",
                arguments={
                    "to_emails": "test@example.com",
                    "subject": "Test Subject",
                    "message": "Test message",
                },
            )
            res_json = result[0].text
            assert "Message sent successfully" in res_json

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_get_current_email_address_success(
        self, mock_get_client, mock_get_access_token, mock_service, mock_profile_data
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Set up the mock service chain properly
        mock_service.users().getProfile().execute.return_value = mock_profile_data

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="get_current_email_address", arguments={}
            )
            res_json = result[0].text
            assert res_json == "user@example.com"

        mock_service.users().getProfile.assert_called_with(userId="me")

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.modify_message_labels")
    async def test_modify_message_labels_success(
        self, mock_modify_labels, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_modify_labels.return_value = {
            "id": "test_id",
            "labelIds": ["INBOX", "STARRED"],
        }

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="modify_message_labels",
                arguments={"email_id": "test_id", "add_label_ids": ["STARRED"]},
            )
            res_json = json.loads(result[0].text)
            assert res_json["id"] == "test_id"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.fetch_email_or_draft")
    async def test_list_attachments_success(
        self, mock_fetch_email, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_email_with_attachments = {
            "payload": {
                "parts": [
                    {"filename": "test.pdf", "body": {"attachmentId": "attachment_id"}}
                ]
            }
        }
        mock_fetch_email.return_value = mock_email_with_attachments

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_attachments", arguments={"email_id": "test_id"}
            )
            res_json = json.loads(result[0].text)
            assert res_json["filename"] == "test.pdf"
            assert res_json["id"] == "attachment_id"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.fetch_email_or_draft")
    async def test_list_attachments_no_attachments(
        self, mock_fetch_email, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_email_no_attachments = {"payload": {}}
        mock_fetch_email.return_value = mock_email_no_attachments

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_attachments", arguments={"email_id": "test_id"}
            )
            assert len(result) == 0


class TestDraftFunctions:
    """Test individual draft-related functions"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_drafts")
    async def test_list_drafts_success(
        self,
        mock_list_drafts_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_draft_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_drafts_func.return_value = mock_draft_data

        async with Client(mcp) as client:
            result = await client.call_tool(name="list_drafts", arguments={})
            res_json = json.loads(result[0].text)
            assert res_json == mock_draft_data[0]

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.create_message_data")
    async def test_create_draft_success(
        self, mock_create_message, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_message.return_value = {"id": "draft_id"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="create_draft",
                arguments={
                    "to_emails": "test@example.com",
                    "subject": "Test Subject",
                    "message": "Test message",
                },
            )
            res_json = result[0].text
            assert "Draft created successfully" in res_json

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_delete_draft_success(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().drafts().delete().execute.return_value = {}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_draft", arguments={"draft_id": "test_id"}
            )
            res_json = result[0].text
            assert "deleted successfully" in res_json

        mock_service.users().drafts().delete.assert_called_with(
            userId="me", id="test_id"
        )

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_send_draft_success(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().drafts().send().execute.return_value = {"id": "sent_id"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="send_draft", arguments={"draft_id": "test_id"}
            )
            res_json = result[0].text
            assert "sent successfully" in res_json

        mock_service.users().drafts().send.assert_called_with(
            userId="me", body={"id": "test_id"}
        )

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.update_draft")
    async def test_update_draft_success(
        self,
        mock_update_draft_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_update_draft_func.return_value = {"id": "test_id"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="update_draft",
                arguments={
                    "draft_id": "test_id",
                    "to_emails": "test@example.com",
                    "subject": "Updated Subject",
                    "message": "Updated message",
                },
            )
            res_json = result[0].text
            assert "updated successfully" in res_json


class TestLabelFunctions:
    """Test individual label-related functions"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_labels")
    async def test_list_labels_success(
        self,
        mock_list_labels_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_label_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_labels_func.return_value = mock_label_data

        async with Client(mcp) as client:
            result = await client.call_tool(name="list_labels", arguments={})
            res_json = json.loads(result[0].text)
            assert res_json == mock_label_data[0]

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.get_label")
    async def test_list_labels_with_id(
        self,
        mock_get_label_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_label_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_get_label_func.return_value = mock_label_data[0]

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_labels", arguments={"label_id": "test_label_id"}
            )
            res_json = json.loads(result[0].text)
            assert res_json == mock_label_data[0]

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.create_label")
    async def test_create_label_success(
        self,
        mock_create_label_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_label_func.return_value = {
            "id": "new_label_id",
            "name": "New Label",
        }

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="create_label", arguments={"label_name": "New Label"}
            )
            res_json = json.loads(result[0].text)
            assert res_json["name"] == "New Label"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.update_label")
    async def test_update_label_success(
        self,
        mock_update_label_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_update_label_func.return_value = {"id": "test_id", "name": "Updated Label"}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="update_label",
                arguments={"label_id": "test_id", "label_name": "Updated Label"},
            )
            res_json = json.loads(result[0].text)
            assert res_json["name"] == "Updated Label"

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.delete_label")
    async def test_delete_label_success(
        self,
        mock_delete_label_func,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_delete_label_func.return_value = "Label deleted successfully"

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_label", arguments={"label_id": "test_id"}
            )
            res_json = result[0].text
            assert "deleted successfully" in res_json


class TestErrorHandling:
    """Test error handling across different scenarios"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_http_error_handling(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        """Test that HttpErrors are properly converted to ToolErrors"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().getProfile().execute.side_effect = HttpError(
            Mock(status=500), b"Internal Server Error"
        )

        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(name="get_current_email_address", arguments={})

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_general_exception_handling(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        """Test that general exceptions are properly converted to ToolErrors"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().getProfile().execute.side_effect = Exception(
            "Unexpected error"
        )

        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(name="get_current_email_address", arguments={})

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_send_email_http_error(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        """Test HttpError handling in send_email"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().messages().send().execute.side_effect = HttpError(
            Mock(status=403), b"Forbidden"
        )

        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(
                    name="send_email",
                    arguments={
                        "to_emails": "test@example.com",
                        "subject": "Test",
                        "message": "Test",
                    },
                )

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_delete_draft_http_error(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        """Test HttpError handling in delete_draft"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_service.users().drafts().delete().execute.side_effect = HttpError(
            Mock(status=404), b"Not Found"
        )

        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(
                    name="delete_draft", arguments={"draft_id": "nonexistent_id"}
                )


class TestComplexScenarios:
    """Test complex business logic scenarios"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_messages")
    @patch("obot_gmail_mcp.server.message_to_string")
    async def test_list_emails_with_category_fallback(
        self,
        mock_message_to_string,
        mock_list_messages,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        """Test the complex category fallback logic in list_emails"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # First call returns empty (no primary category emails)
        # Second call returns empty (estimate check)
        # Third call returns emails (fallback without category)
        mock_list_messages.side_effect = [[], [], [{"id": "email1"}]]
        mock_message_to_string.return_value = (None, "Formatted Email")

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_emails",
                arguments={"category": "primary", "label_ids": "INBOX"},
            )
            res_json = result[0].text
            assert res_json == "Formatted Email"

        # Should have made 3 calls to list_messages due to fallback logic
        assert mock_list_messages.call_count == 3

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.fetch_email_or_draft")
    @patch("obot_gmail_mcp.server.get_email_body")
    @patch("obot_gmail_mcp.server.has_attachment")
    @patch("obot_gmail_mcp.server.format_message_metadata")
    async def test_read_email_with_attachment(
        self,
        mock_format_metadata,
        mock_has_attachment,
        mock_get_body,
        mock_fetch_email,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        """Test reading email with attachment includes link"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_fetch_email.return_value = {"id": "test_id"}
        mock_get_body.return_value = "Test email body"
        mock_has_attachment.return_value = True  # Email has attachment
        mock_format_metadata.return_value = (None, "Test metadata")

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="read_email", arguments={"email_id": "test_id"}
            )
            res_json = json.loads(result[0].text)
            assert res_json["has_attachment"] == True
            assert "link" in res_json
            assert "mail.google.com" in res_json["link"]

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    async def test_read_email_by_subject(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        """Test reading email by subject instead of ID"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Mock the search response
        mock_service.users().messages().list().execute.return_value = {
            "messages": [{"id": "found_email_id"}]
        }

        # Mock the rest of the email reading process
        with (
            patch("obot_gmail_mcp.server.fetch_email_or_draft") as mock_fetch,
            patch("obot_gmail_mcp.server.get_email_body") as mock_get_body,
            patch("obot_gmail_mcp.server.has_attachment") as mock_has_attachment,
            patch(
                "obot_gmail_mcp.server.format_message_metadata"
            ) as mock_format_metadata,
        ):
            mock_fetch.return_value = {"id": "found_email_id"}
            mock_get_body.return_value = "Email body"
            mock_has_attachment.return_value = False
            mock_format_metadata.return_value = (None, "Metadata")

            async with Client(mcp) as client:
                result = await client.call_tool(
                    name="read_email", arguments={"email_subject": "Test Subject"}
                )
                res_json = json.loads(result[0].text)
                assert res_json["body"] == "Email body"

            # Verify search was called with correct query
            mock_service.users().messages().list.assert_called_with(
                userId="me", q='subject:"Test Subject"'
            )


# Performance and Load Testing (optional)
class TestPerformance:
    """Test performance characteristics"""

    @patch("obot_gmail_mcp.server._get_access_token")
    @patch("obot_gmail_mcp.server.get_client")
    @patch("obot_gmail_mcp.server.list_messages")
    @patch("obot_gmail_mcp.server.message_to_string")
    async def test_large_email_list(
        self,
        mock_message_to_string,
        mock_list_messages,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        """Test handling of large email lists"""
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        # Create a large list of mock emails
        large_email_list = [{"id": f"email_{i}"} for i in range(1000)]
        mock_list_messages.return_value = large_email_list
        mock_message_to_string.return_value = (None, "Formatted Email")

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_emails", arguments={"max_results": 1000}
            )
            res_json = json.loads(result[0].text)
            assert len(res_json) == 1000
            assert all(email == "Formatted Email" for email in res_json)


if __name__ == "__main__":
    pytest.main(["-v"])
