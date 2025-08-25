import pytest
from unittest.mock import patch, MagicMock, Mock
from fastmcp import Client
from fastmcp.exceptions import ToolError
from googleapiclient.errors import HttpError

# Add the parent directory to the Python path so we can import from src
from app.server import mcp

# Configure pytest for async support
pytestmark = pytest.mark.asyncio


# Fixtures
@pytest.fixture
def mock_service():
    """Create a mock Google Drive service"""
    service = MagicMock()
    return service


@pytest.fixture
def mock_file_data():
    """Sample file data for testing"""
    return [
        {
            "id": "test_file_id_1",
            "name": "Test Document.docx",
            "mimeType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "parents": ["root"],
            "modifiedTime": "2024-01-01T10:00:00.000Z",
            "size": "12345",
            "webViewLink": "https://drive.google.com/file/d/test_file_id_1/view",
        },
        {
            "id": "test_file_id_2",
            "name": "Test Spreadsheet.xlsx",
            "mimeType": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            "parents": ["test_folder_id"],
            "modifiedTime": "2024-01-02T10:00:00.000Z",
            "size": "67890",
            "webViewLink": "https://drive.google.com/file/d/test_file_id_2/view",
        },
    ]


@pytest.fixture
def mock_folder_data():
    """Sample folder data for testing"""
    return {
        "id": "test_folder_id",
        "name": "Test Folder",
        "mimeType": "application/vnd.google-apps.folder",
        "parents": ["root"],
        "modifiedTime": "2024-01-01T09:00:00.000Z",
        "webViewLink": "https://drive.google.com/drive/folders/test_folder_id",
    }


@pytest.fixture
def mock_permission_data():
    """Sample permission data for testing"""
    return [
        {
            "id": "test_permission_id_1",
            "type": "user",
            "role": "reader",
            "emailAddress": "user1@example.com",
            "displayName": "Test User 1",
        },
        {
            "id": "test_permission_id_2",
            "type": "user",
            "role": "writer",
            "emailAddress": "user2@example.com",
            "displayName": "Test User 2",
        },
    ]


@pytest.fixture
def mock_shared_drive_data():
    """Sample shared drive data for testing"""
    return [
        {
            "id": "test_drive_id_1",
            "name": "Test Shared Drive 1",
            "capabilities": {
                "canAddChildren": True,
                "canChangeCopyRequiresWriterPermissionRestriction": True,
                "canComment": True,
            },
        },
        {
            "id": "test_drive_id_2",
            "name": "Test Shared Drive 2",
            "capabilities": {
                "canAddChildren": True,
                "canChangeCopyRequiresWriterPermissionRestriction": False,
                "canComment": True,
            },
        },
    ]


# Integration Tests - Testing the MCP Server
class TestMCPServer:
    """Test the FastMCP server integration"""

    async def test_list_tools(self):
        """Test that all tools are properly registered"""
        # Test that tools are registered in the MCP server
        async with Client(mcp) as client:
            tools = await client.list_tools()
            assert (
                len(tools) == 16
            )  # Total number of active tools (excluding commented out ones)

        # Verify expected tools are present
        tool_names = [tool.name for tool in tools]
        expected_tools = [
            "list_files",
            "copy_file",
            "get_file",
            "update_file",
            "create_folder",
            "delete_file",
            "transfer_ownership",
            "list_permissions",
            "get_permission",
            "create_permission",
            "update_permission",
            "delete_permission",
            "list_shared_drives",
            "create_shared_drive",
            "delete_shared_drive",
            "rename_shared_drive",
        ]
        for tool in expected_tools:
            assert tool in tool_names


# Unit Tests - Testing Individual Functions
class TestFileFunctions:
    """Test individual file-related functions"""

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.list_files")
    async def test_list_files_success(
        self,
        mock_list_files,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_file_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_files.return_value = mock_file_data

        mock_service.files().list().execute.return_value = {"files": mock_file_data}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_files", arguments={"max_results": 50}
            )
            assert result.structured_content["result"] == mock_file_data

        mock_list_files.assert_called_once_with(
            mock_service,
            drive_id=None,
            parent_id=None,
            mime_type=None,
            file_name_contains=None,
            modified_time_after=None,
            max_results=50,
            trashed=False,
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.list_files")
    async def test_list_files_with_filters(
        self,
        mock_list_files,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_file_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_files.return_value = mock_file_data

        mock_service.files().list().execute.return_value = {"files": mock_file_data}

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_files",
                arguments={
                    "drive_id": "test_drive_id",
                    "parent_id": "test_parent_id",
                    "mime_type": "application/pdf",
                    "file_name_contains": "test",
                    "modified_time_after": "2024-01-01T00:00:00Z",
                    "max_results": 10,
                },
            )
            # result is already the data
            assert result.structured_content["result"] == mock_file_data

        mock_list_files.assert_called_once_with(
            mock_service,
            drive_id="test_drive_id",
            parent_id="test_parent_id",
            mime_type="application/pdf",
            file_name_contains="test",
            modified_time_after="2024-01-01T00:00:00Z",
            max_results=10,
            trashed=False,
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.get_file")
    async def test_get_file_success(
        self,
        mock_get_file,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_file_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_get_file.return_value = mock_file_data[0]

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="get_file", arguments={"file_id": "test_file_id_1"}
            )
            # result is already the data
            assert result.structured_content == mock_file_data[0]

        mock_get_file.assert_called_once_with(mock_service, "test_file_id_1")

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.copy_file")
    async def test_copy_file_success(
        self,
        mock_copy_file,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_file_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        copied_file = mock_file_data[0].copy()
        copied_file["name"] = "Copy of Test Document.docx"
        mock_copy_file.return_value = copied_file

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="copy_file",
                arguments={
                    "file_id": "test_file_id_1",
                    "new_name": "Copy of Test Document.docx",
                },
            )
            # result is already the data
            assert result.structured_content == copied_file

        mock_copy_file.assert_called_once_with(
            mock_service,
            file_id="test_file_id_1",
            new_name="Copy of Test Document.docx",
            parent_id=None,
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.update_file")
    async def test_update_file_success(
        self,
        mock_update_file,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_file_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        updated_file = mock_file_data[0].copy()
        updated_file["name"] = "Updated Document.docx"
        mock_update_file.return_value = updated_file

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="update_file",
                arguments={
                    "file_id": "test_file_id_1",
                    "new_name": "Updated Document.docx",
                    "new_parent_id": "new_parent_id",
                },
            )
            # result is already the data
            assert result.structured_content == updated_file

        mock_update_file.assert_called_once_with(
            mock_service,
            file_id="test_file_id_1",
            new_name="Updated Document.docx",
            new_content=None,
            mime_type=None,
            new_parent_id="new_parent_id",
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.create_folder")
    async def test_create_folder_success(
        self,
        mock_create_folder,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_folder_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_folder.return_value = mock_folder_data

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="create_folder",
                arguments={"folder_name": "Test Folder", "parent_id": "root"},
            )
            # result is already the data
            assert result.structured_content == mock_folder_data

        mock_create_folder.assert_called_once_with(
            mock_service, name="Test Folder", parent_id="root"
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.delete_file")
    async def test_delete_file_success(
        self, mock_delete_file, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_delete_file.return_value = True

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_file", arguments={"file_id": "test_file_id_1"}
            )
            assert (
                result.structured_content["result"]
                == "Successfully deleted file: test_file_id_1"
            )

        mock_delete_file.assert_called_once_with(mock_service, "test_file_id_1")

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.delete_file")
    async def test_delete_file_failure(
        self, mock_delete_file, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_delete_file.return_value = False

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_file", arguments={"file_id": "test_file_id_1"}
            )
            assert (
                result.structured_content["result"]
                == "Failed to delete file: test_file_id_1"
            )

        mock_delete_file.assert_called_once_with(mock_service, "test_file_id_1")


class TestPermissionFunctions:
    """Test individual permission-related functions"""

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.list_permissions")
    async def test_list_permissions_success(
        self,
        mock_list_permissions,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_permission_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_permissions.return_value = mock_permission_data

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="list_permissions", arguments={"file_id": "test_file_id_1"}
            )
            # result is already the data
            assert result.structured_content["result"] == mock_permission_data

        mock_list_permissions.assert_called_once_with(mock_service, "test_file_id_1")

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.get_permission")
    async def test_get_permission_success(
        self,
        mock_get_permission,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_permission_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_get_permission.return_value = mock_permission_data[0]

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="get_permission",
                arguments={
                    "file_id": "test_file_id_1",
                    "permission_id": "test_permission_id_1",
                },
            )
            # result is already the data
            assert result.structured_content == mock_permission_data[0]

        mock_get_permission.assert_called_once_with(
            mock_service, "test_file_id_1", "test_permission_id_1"
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.create_permission")
    async def test_create_permission_success(
        self,
        mock_create_permission,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_permission_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_permission.return_value = mock_permission_data[0]

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="create_permission",
                arguments={
                    "file_id": "test_file_id_1",
                    "role": "reader",
                    "type": "user",
                    "email_address": "user1@example.com",
                },
            )
            # result is already the data
            assert result.structured_content == mock_permission_data[0]

        mock_create_permission.assert_called_once_with(
            mock_service,
            file_id="test_file_id_1",
            role="reader",
            type="user",
            email_address="user1@example.com",
            domain=None,
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    async def test_create_permission_missing_email(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        async with Client(mcp) as client:
            with pytest.raises(Exception) as exc_info:
                await client.call_tool(
                    name="create_permission",
                    arguments={
                        "file_id": "test_file_id_1",
                        "role": "reader",
                        "type": "user",
                    },
                )
            assert "EMAIL_ADDRESS is required for user/group permission" in str(
                exc_info.value
            )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    async def test_create_permission_missing_domain(
        self, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service

        async with Client(mcp) as client:
            with pytest.raises(Exception) as exc_info:
                await client.call_tool(
                    name="create_permission",
                    arguments={
                        "file_id": "test_file_id_1",
                        "role": "reader",
                        "type": "domain",
                    },
                )
            assert "DOMAIN is required for domain permission" in str(exc_info.value)

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.update_permission")
    async def test_update_permission_success(
        self,
        mock_update_permission,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_permission_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        updated_permission = mock_permission_data[0].copy()
        updated_permission["role"] = "writer"
        mock_update_permission.return_value = updated_permission

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="update_permission",
                arguments={
                    "file_id": "test_file_id_1",
                    "permission_id": "test_permission_id_1",
                    "role": "writer",
                },
            )
            # result is already the data
            assert result.structured_content == updated_permission

        mock_update_permission.assert_called_once_with(
            mock_service, "test_file_id_1", "test_permission_id_1", "writer"
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.delete_permission")
    async def test_delete_permission_success(
        self,
        mock_delete_permission,
        mock_get_client,
        mock_get_access_token,
        mock_service,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_delete_permission.return_value = True

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_permission",
                arguments={
                    "file_id": "test_file_id_1",
                    "permission_id": "test_permission_id_1",
                },
            )
            assert (
                result.structured_content["result"]
                == "Successfully deleted permission: test_permission_id_1"
            )

        mock_delete_permission.assert_called_once_with(
            mock_service, "test_file_id_1", "test_permission_id_1"
        )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.transfer_ownership")
    async def test_transfer_ownership_success(
        self,
        mock_transfer_ownership,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_permission_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        ownership_permission = mock_permission_data[0].copy()
        ownership_permission["role"] = "owner"
        mock_transfer_ownership.return_value = ownership_permission

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="transfer_ownership",
                arguments={
                    "file_id": "test_file_id_1",
                    "new_owner_email": "newowner@example.com",
                },
            )
            # result is already the data
            assert result.structured_content == ownership_permission

        mock_transfer_ownership.assert_called_once_with(
            mock_service, "test_file_id_1", "newowner@example.com"
        )


class TestSharedDriveFunctions:
    """Test individual shared drive-related functions"""

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.list_drives")
    async def test_list_shared_drives_success(
        self,
        mock_list_drives,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_shared_drive_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_list_drives.return_value = mock_shared_drive_data

        async with Client(mcp) as client:
            result = await client.call_tool(name="list_shared_drives", arguments={})
            # result is already the data
            assert result.structured_content["result"] == mock_shared_drive_data

        mock_list_drives.assert_called_once_with(mock_service)

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.create_drive")
    async def test_create_shared_drive_success(
        self,
        mock_create_drive,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_shared_drive_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_create_drive.return_value = mock_shared_drive_data[0]

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="create_shared_drive",
                arguments={"drive_name": "Test Shared Drive 1"},
            )
            # result is already the data
            assert result.structured_content == mock_shared_drive_data[0]

        mock_create_drive.assert_called_once_with(mock_service, "Test Shared Drive 1")

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.delete_drive")
    async def test_delete_shared_drive_success(
        self, mock_delete_drive, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_delete_drive.return_value = None  # delete_drive doesn't return anything

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="delete_shared_drive", arguments={"drive_id": "test_drive_id_1"}
            )
            # result is already the data
            expected = {
                "success": True,
                "message": "Successfully deleted shared drive: test_drive_id_1",
            }
            assert result.structured_content == expected

        mock_delete_drive.assert_called_once_with(mock_service, "test_drive_id_1")

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.update_drive")
    async def test_rename_shared_drive_success(
        self,
        mock_update_drive,
        mock_get_client,
        mock_get_access_token,
        mock_service,
        mock_shared_drive_data,
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        updated_drive = mock_shared_drive_data[0].copy()
        updated_drive["name"] = "Renamed Shared Drive"
        mock_update_drive.return_value = updated_drive

        async with Client(mcp) as client:
            result = await client.call_tool(
                name="rename_shared_drive",
                arguments={
                    "drive_id": "test_drive_id_1",
                    "drive_name": "Renamed Shared Drive",
                },
            )
            # result is already the data
            assert result.structured_content == updated_drive

        mock_update_drive.assert_called_once_with(
            mock_service, "test_drive_id_1", "Renamed Shared Drive"
        )


class TestErrorHandling:
    """Test error handling scenarios"""

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.list_files")
    async def test_http_error_handling(
        self, mock_list_files, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_resp = Mock()
        mock_resp.status = 403
        mock_list_files.side_effect = HttpError(
            resp=mock_resp, content=b'{"error": {"message": "Forbidden"}}'
        )

        async with Client(mcp) as client:
            with pytest.raises(ToolError) as exc_info:
                await client.call_tool(name="list_files", arguments={})
            assert (
                "HttpError.__init__() missing 1 required positional argument: 'content'"
                in str(exc_info.value)
            )

    @patch("app.server._get_access_token")
    @patch("app.server.get_client")
    @patch("app.server.get_file")
    async def test_generic_error_handling(
        self, mock_get_file, mock_get_client, mock_get_access_token, mock_service
    ):
        mock_get_access_token.return_value = "fake_token"
        mock_get_client.return_value = mock_service
        mock_get_file.side_effect = Exception("Generic error")

        async with Client(mcp) as client:
            with pytest.raises(ToolError) as exc_info:
                await client.call_tool(
                    name="get_file", arguments={"file_id": "test_file_id"}
                )
            assert "Unexpected ToolError" in str(exc_info.value)
