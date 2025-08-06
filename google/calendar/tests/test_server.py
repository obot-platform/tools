import pytest
from unittest.mock import patch, MagicMock, Mock
from fastmcp import Client
from fastmcp.exceptions import ToolError
from googleapiclient.errors import HttpError
import sys
import os
# Add the parent directory to the Python path so we can import from app
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from app.server import mcp

# Configure pytest for async support
pytestmark = pytest.mark.asyncio


# Fixtures
@pytest.fixture
def mock_service():
    """Create a mock Google Calendar service"""
    service = MagicMock()
    return service


@pytest.fixture
def mock_calendar_data():
    """Sample calendar data for testing"""
    return [{
        'id': 'test_calendar_id',
        'summary': 'Test Calendar',
        'description': 'Test Description',
        'timeZone': 'America/New_York'
    },
            {
        'id': 'test_calendar_id_2',
        'summary': 'Test Calendar 2',
        'description': 'Test Description 2',
        'timeZone': 'America/Los_Angeles'
    }]


@pytest.fixture
def mock_event_data():
    """Sample event data for testing"""
    return [{
        'id': 'test_event_id',
        'summary': 'Test Event',
        'description': 'Test Description',
        'start': {'dateTime': '2024-01-01T10:00:00-05:00', 'timeZone': 'America/New_York'},
        'end': {'dateTime': '2024-01-01T11:00:00-05:00', 'timeZone': 'America/New_York'},
        'eventType': 'default'
    },
        {
        'id': 'test_event_id_2',
        'summary': 'Test Event 2',
        'description': 'Test Description 2',
        'start': {'dateTime': '2024-01-01T10:00:00-05:00', 'timeZone': 'America/New_York'},
        'end': {'dateTime': '2024-01-01T11:00:00-05:00', 'timeZone': 'America/New_York'},
        'eventType': 'default'
    }]


# Integration Tests - Testing the MCP Server
class TestMCPServer:
    """Test the FastMCP server integration"""
    
    async def test_list_tools(self):
        """Test that all tools are properly registered"""
        # Test that tools are registered in the MCP server
        async with Client(mcp) as client:
            tools = await client.list_tools()
            assert len(tools) == 14
        
        # Verify expected tools are present
        tool_names = [tool.name for tool in tools]
        expected_tools = [
            'list_calendars', 'get_calendar', 'create_calendar', 
            'update_calendar', 'delete_calendar', 'list_events',
            'get_event', 'move_event', 'quick_add_event', 
            'create_event', 'update_event', 'respond_to_event',
            'delete_event', 'list_recurring_event_instances'
        ]
        for tool in expected_tools:
            assert tool in tool_names

# Unit Tests - Testing Individual Functions
class TestCalendarFunctions:
    """Test individual calendar-related functions"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_list_calendars_success(self, mock_get_client, mock_get_token, mock_service, mock_calendar_data):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendarList().list().execute.return_value = {'items': mock_calendar_data}
        
        async with Client(mcp) as client:
            # Call the tool function directly
            result = await client.call_tool(
                name='list_calendars',
                arguments={}
            )
            assert result.data == mock_calendar_data
        mock_service.calendarList().list().execute.assert_called_once()

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_list_calendars_http_error(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendarList().list().execute.side_effect = HttpError(
            Mock(status=403), b'Forbidden'
        )
        
        with pytest.raises(ToolError, match="Failed to list google calendars"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='list_calendars',
                    arguments={}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_get_calendar_success(self, mock_get_client, mock_get_token, mock_service, mock_calendar_data):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        # Use the first calendar from mock_calendar_data (it's a list)
        mock_service.calendars().get().execute.return_value = mock_calendar_data[0]
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='get_calendar',
                arguments={'calendar_id': 'test_id'}
            )
            res_json = result.data
            assert res_json == mock_calendar_data[0]
        mock_service.calendars().get.assert_called_with(calendarId='test_id')

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_get_calendar_empty_id(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendars().get().execute.side_effect = ValueError("argument `calendar_id` can't be empty")
        with pytest.raises(ToolError, match="argument `calendar_id` can't be empty"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='get_calendar',
                    arguments={'calendar_id': ''}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    @patch('app.server.get_user_timezone')
    async def test_create_calendar_success(self, mock_get_timezone, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_get_timezone.return_value = 'America/New_York'
        mock_service.calendars().insert().execute.return_value = {'id': 'new_id'}
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='create_calendar',
                arguments={'summary': 'Test Calendar'}
            )
            res_json = result.data
            assert res_json == {'id': 'new_id'}

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_create_calendar_empty_summary(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendars().insert().execute.side_effect = ValueError("argument `summary` can't be empty")
        with pytest.raises(ToolError, match="argument `summary` can't be empty"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='create_calendar',
                    arguments={'summary': ''}
                )


class TestEventFunctions:
    """Test individual event-related functions"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_list_events_success(self, mock_get_client, mock_get_token, mock_service, mock_event_data):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.events().list().execute.return_value = {'items': mock_event_data}
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='list_events',
                arguments={'calendar_id': 'calendar_id'}
            )
            res_json = result.data
            assert res_json == mock_event_data

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_list_events_empty_calendar_id(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.events().list().execute.side_effect = ValueError("argument `calendar_id` can't be empty")
        with pytest.raises(ToolError, match="argument `calendar_id` can't be empty"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='list_events',
                    arguments={'calendar_id': ''}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_get_event_success(self, mock_get_client, mock_get_token, mock_service, mock_event_data):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        # Use the first event from mock_event_data (it's a list)
        mock_service.events().get().execute.return_value = mock_event_data[0]
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='get_event',
                arguments={'calendar_id': 'calendar_id', 'event_id': 'event_id'}
            )
            res_json = result.data
            assert res_json == mock_event_data[0]

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_get_event_empty_params(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.events().get().execute.side_effect = ValueError("argument `calendar_id` can't be empty")
        with pytest.raises(ToolError, match="argument `calendar_id` can't be empty"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='get_event',
                    arguments={'calendar_id': '', 'event_id': 'event_id'}
                )
        
        mock_service.events().get().execute.side_effect = ValueError("argument `event_id` can't be empty")
        with pytest.raises(ToolError, match="argument `event_id` can't be empty"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='get_event',
                    arguments={'calendar_id': 'calendar_id', 'event_id': ''}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    @patch('app.server.get_user_timezone')
    async def test_create_event_success(self, mock_get_timezone, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_get_timezone.return_value = 'America/New_York'
        mock_service.events().insert().execute.return_value = {'id': 'new_event_id'}
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='create_event',
                arguments={'calendar_id': 'calendar_id', 'summary': 'Test Event', 'start_datetime': '2024-01-01T10:00:00-05:00', 'end_datetime': '2024-01-01T11:00:00-05:00'}
            )
            res_json = result.data
            assert res_json == {'id': 'new_event_id'}

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_create_event_missing_time(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        
        with pytest.raises(ToolError, match="Either start_date or start_datetime must be provided"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='create_event',
                    arguments={'calendar_id': 'calendar_id'}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_create_event_invalid_datetime(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        
        with pytest.raises(ToolError, match="Invalid start_datetime"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='create_event',
                    arguments={'calendar_id': 'calendar_id', 'start_datetime': 'invalid-datetime', 'end_datetime': '2024-01-01T11:00:00-05:00'}
                )


class TestValidationFunctions:
    """Test input validation and edge cases"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_rfc3339_validation(self, mock_get_client, mock_get_token):
        """Test that valid RFC3339 timestamps are accepted"""
        mock_get_token.return_value = 'fake_token'
        mock_service = MagicMock()
        mock_get_client.return_value = mock_service
        mock_service.events().list().execute.return_value = {'items': []}
        
        # Valid RFC3339 timestamps should not raise errors
        valid_timestamps = [
            '2024-01-01T10:00:00Z',
            '2024-01-01T10:00:00-05:00',
            '2024-01-01T10:00:00+00:00'
        ]
        
        for timestamp in valid_timestamps:
            # Should not raise ValueError for valid timestamps
            try:
                async with Client(mcp) as client:
                    await client.call_tool(
                        name='list_events',
                        arguments={'calendar_id': 'calendar_id', 'time_min': timestamp}
                )
            except ValueError as e:
                if "Invalid time_min" in str(e):
                    pytest.fail(f"Valid timestamp {timestamp} was rejected")

    async def test_timezone_validation(self):
        """Test timezone validation logic"""
        from app.tools.event import is_valid_iana_timezone
        assert is_valid_iana_timezone('America/New_York')
        assert is_valid_iana_timezone('UTC')
        assert not is_valid_iana_timezone('Invalid/Timezone')


class TestErrorHandling:
    """Test error handling across different scenarios"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_http_error_handling(self, mock_get_client, mock_get_token, mock_service):
        """Test that HttpErrors are properly converted to ToolErrors"""
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendarList().list().execute.side_effect = HttpError(
            Mock(status=500), b'Internal Server Error'
        )
        
        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='list_calendars',
                    arguments={}
                )

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_general_exception_handling(self, mock_get_client, mock_get_token, mock_service):
        """Test that general exceptions are properly converted to ToolErrors"""
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        mock_service.calendarList().list().execute.side_effect = Exception('Unexpected error')
        
        with pytest.raises(ToolError):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='list_calendars',
                    arguments={}
                )


class TestComplexScenarios:
    """Test complex business logic scenarios"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    @patch('app.server.get_current_user_email')
    async def test_respond_to_event_success(self, mock_get_email, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_email.return_value = 'user@example.com'
        mock_get_client.return_value = mock_service
        
        event_with_attendees_before = {
            'id': 'event_id',
            'attendees': [
                {'email': 'user@example.com', 'responseStatus': 'needsAction'},
            ]
        }
        
        event_with_attendees = {
            'id': 'event_id',
            'attendees': [
                {'email': 'user@example.com', 'responseStatus': 'accepted'},
            ]
        }
        
        mock_service.events().get().execute.return_value = event_with_attendees_before
        mock_service.events().patch().execute.return_value = event_with_attendees
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='respond_to_event',
                arguments={'calendar_id': 'calendar_id', 'event_id': 'event_id', 'response': 'accepted'}
            )
            res_json = result.data
            assert res_json == event_with_attendees
        
        # Verify the patch was called with updated attendees
        mock_service.events().patch().execute.assert_called_once()

    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_move_event_invalid_type(self, mock_get_client, mock_get_token, mock_service):
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        
        # Mock an event type that cannot be moved
        unmovable_event = {
            'id': 'event_id',
            'eventType': 'birthday'  # Not in MOVABLE_EVENT_TYPES
        }
        
        mock_service.events().get().execute.return_value = unmovable_event
        
        with pytest.raises(ToolError, match="Events with type 'birthday' can not be moved"):
            async with Client(mcp) as client:
                await client.call_tool(
                    name='move_event',
                    arguments={'calendar_id': 'calendar_id', 'event_id': 'event_id', 'new_calendar_id': 'new_calendar_id'}
                )


# Performance and Load Testing (optional)
class TestPerformance:
    """Test performance characteristics"""
    
    @patch('app.server._get_access_token')
    @patch('app.server.get_client')
    async def test_large_event_list_pagination(self, mock_get_client, mock_get_token, mock_service):
        """Test that pagination works correctly for large result sets"""
        mock_get_token.return_value = 'fake_token'
        mock_get_client.return_value = mock_service
        
        # Mock paginated responses
        first_page = {
            'items': [{'id': f'event_{i}'} for i in range(250)],
            'nextPageToken': 'page2'
        }
        second_page = {
            'items': [{'id': f'event_{i}'} for i in range(250, 300)]
        }
        
        mock_service.events().list().execute.side_effect = [first_page, second_page]
        
        async with Client(mcp) as client:
            result = await client.call_tool(
                name='list_events',
                arguments={'calendar_id': 'calendar_id', 'max_results': 300}
            )
        
        # Should return exactly 300 events (250 + 50)
        res_json = result.data
        assert len(res_json) == 300
        # Should have made two API calls due to pagination
        assert mock_service.events().list().execute.call_count == 2


if __name__ == '__main__':
    pytest.main(['-v'])