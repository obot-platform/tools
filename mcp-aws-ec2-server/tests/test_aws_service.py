import boto3
import pytest
from moto import mock_aws

from mcp_aws_ec2_server.aws_service import AWSService


@pytest.fixture
def aws_service():
    """Create a mock AWS service."""
    with mock_aws(
        config={"ec2": {"use_default_vpc": False, "use_default_subnet": False}}
    ):
        # Create a mock session
        session = boto3.Session()
        service = AWSService(session=session)
        yield service


@pytest.fixture
def mock_vpc(aws_service):
    """Create a mock VPC."""
    ec2 = aws_service.ec2_client
    vpc = ec2.create_vpc(CidrBlock="10.0.0.0/16")
    return vpc["Vpc"]


@pytest.fixture
def mock_subnet(aws_service, mock_vpc):
    """Create a mock subnet."""
    ec2 = aws_service.ec2_client
    subnet = ec2.create_subnet(
        VpcId=mock_vpc["VpcId"],
        CidrBlock="10.0.0.0/24",
    )
    return subnet["Subnet"]


@pytest.fixture
def mock_security_group(aws_service, mock_vpc):
    """Create a mock security group."""
    ec2 = aws_service.ec2_client
    sg = ec2.create_security_group(
        GroupName="test-sg",
        Description="Test security group",
        VpcId=mock_vpc["VpcId"],
    )
    return sg


@pytest.fixture
def mock_instance(aws_service, mock_subnet, mock_security_group):
    """Create a mock EC2 instance."""
    instance = aws_service.create_instance(
        instance_type="t2.micro",
        image_id="ami-12345678",
        subnet_id=mock_subnet["SubnetId"],
        security_group_ids=[mock_security_group["GroupId"]],
        instance_name="test-instance",
        tags={"Environment": "test"},
    )
    return instance


def test_list_instances_empty(aws_service):
    """Test listing instances when none exist."""
    result = aws_service.list_instances()
    assert "instances" in result
    assert len(result["instances"]) == 0


def test_list_instances(aws_service, mock_instance):
    """Test listing instances with one instance."""
    result = aws_service.list_instances()
    assert "instances" in result
    assert len(result["instances"]) == 1
    instance = result["instances"][0]
    assert instance["name"] == "test-instance"
    assert instance["instance_type"] == "t2.micro"


def test_create_instance(aws_service, mock_subnet, mock_security_group):
    """Test creating an instance."""
    result = aws_service.create_instance(
        instance_type="t2.micro",
        image_id="ami-12345678",
        subnet_id=mock_subnet["SubnetId"],
        security_group_ids=[mock_security_group["GroupId"]],
        instance_name="test-instance",
        tags={"Environment": "test"},
    )
    assert "message" in result
    assert "Instance" in result["message"]
    assert "created" in result["message"]


def test_start_stop_instance(aws_service, mock_instance):
    """Test starting and stopping an instance."""
    # Get the instance ID from the mock_instance fixture
    instances = aws_service.list_instances()["instances"]
    instance_id = instances[0]["instance_id"]

    # Stop the instance
    result = aws_service.stop_instance(instance_id)
    assert "message" in result
    assert instance_id in result["message"]
    assert "stopped" in result["message"]

    # Start the instance
    result = aws_service.start_instance(instance_id)
    assert "message" in result
    assert instance_id in result["message"]
    assert "started" in result["message"]


def test_terminate_instance(aws_service, mock_instance):
    """Test terminating an instance."""
    # Get the instance ID from the mock_instance fixture
    instances = aws_service.list_instances()["instances"]
    instance_id = instances[0]["instance_id"]

    # Terminate the instance
    result = aws_service.terminate_instance(instance_id)
    assert "message" in result
    assert instance_id in result["message"]
    assert "terminated" in result["message"]

    # Wait a moment for the instance to be fully terminated
    import time

    time.sleep(0.1)

    # Verify the instance is gone
    instances = aws_service.list_instances()["instances"]
    assert len(instances) == 0


def test_list_vpcs(aws_service, mock_vpc):
    """Test listing VPCs."""
    result = aws_service.list_vpcs()
    assert "vpcs" in result
    # Filter to only include our test VPC
    test_vpcs = [vpc for vpc in result["vpcs"] if vpc["VpcId"] == mock_vpc["VpcId"]]
    assert len(test_vpcs) == 1
    assert test_vpcs[0]["VpcId"] == mock_vpc["VpcId"]


def test_list_subnets(aws_service, mock_subnet):
    """Test listing subnets."""
    result = aws_service.list_subnets()
    assert "subnets" in result
    # Filter to only include our test subnet
    test_subnets = [
        subnet
        for subnet in result["subnets"]
        if subnet["SubnetId"] == mock_subnet["SubnetId"]
    ]
    assert len(test_subnets) == 1
    assert test_subnets[0]["SubnetId"] == mock_subnet["SubnetId"]


def test_list_security_groups(aws_service, mock_vpc, mock_security_group):
    """Test listing security groups."""
    result = aws_service.list_security_groups(mock_vpc["VpcId"])
    assert "security_groups" in result
    # Filter to only include our test security group
    test_sgs = [
        sg
        for sg in result["security_groups"]
        if sg["GroupId"] == mock_security_group["GroupId"]
    ]
    assert len(test_sgs) == 1
    assert test_sgs[0]["GroupId"] == mock_security_group["GroupId"]
