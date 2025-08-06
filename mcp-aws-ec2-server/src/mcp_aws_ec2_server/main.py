from typing import Annotated

from fastmcp import FastMCP

from .aws_service import AWSService

mcp = FastMCP(name="aws-ec2")
aws_service = None


def get_aws_service(region: str = "us-east-1") -> AWSService:
    """Get or create AWS service instance."""
    global aws_service
    if aws_service is None or aws_service.region_name != region:
        aws_service = AWSService(region_name=region)
    return aws_service


@mcp.tool()
async def list_instances(
    aws_region: Annotated[
        str, "The AWS region to list instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all instances in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.list_instances()


@mcp.tool()
async def start_instance(
    instance_id: Annotated[str, "The ID of the instance to start"],
    aws_region: Annotated[
        str, "The AWS region to start the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Start the specified ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.start_instance(instance_id)


@mcp.tool()
async def stop_instance(
    instance_id: Annotated[str, "The ID of the instance to stop"],
    aws_region: Annotated[
        str, "The AWS region to stop the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Stop the specified ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.stop_instance(instance_id)


@mcp.tool()
async def terminate_instance(
    instance_id: Annotated[str, "The ID of the instance to terminate"],
    aws_region: Annotated[
        str, "The AWS region to terminate the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Terminate the specified ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.terminate_instance(instance_id)


@mcp.tool()
async def reboot_instance(
    instance_id: Annotated[str, "The ID of the instance to reboot"],
    aws_region: Annotated[
        str, "The AWS region to reboot the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Reboot the specified ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.reboot_instance(instance_id)


@mcp.tool()
async def create_instance(
    instance_type: Annotated[str, "The type of instance to create"],
    image_id: Annotated[str, "The ID of the image to use"],
    subnet_id: Annotated[str, "The ID of the subnet to use"],
    key_name: Annotated[str | None, "The name of the key pair to use"] = None,
    instance_name: Annotated[str | None, "The name of the instance"] = None,
    tags: Annotated[dict | None, "The tags to apply to the instance"] = None,
    user_data: Annotated[str | None, "The user data to pass to the instance"] = None,
    security_group_ids: Annotated[
        list[str], "The IDs of the security groups to use"
    ] = [],
    aws_region: Annotated[
        str, "The AWS region to create the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Create a new ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.create_instance(
        instance_type=instance_type,
        image_id=image_id,
        subnet_id=subnet_id,
        key_name=key_name,
        instance_name=instance_name,
        tags=tags,
        user_data=user_data,
        security_group_ids=security_group_ids,
    )


@mcp.tool()
async def list_vpcs(
    aws_region: Annotated[
        str, "The AWS region to list VPCs in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all VPCs in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.list_vpcs()


@mcp.tool()
async def list_subnets(
    aws_region: Annotated[
        str, "The AWS region to list subnets in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all subnets in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.list_subnets()


@mcp.tool()
async def list_security_groups(
    vpc_id: Annotated[str, "The ID of the VPC to list security groups in"],
    aws_region: Annotated[
        str, "The AWS region to list security groups in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all security groups in the specified AWS region and VPC."""
    service = get_aws_service(aws_region)
    return service.list_security_groups(vpc_id)


@mcp.tool()
async def get_instance_details(
    instance_id: Annotated[str, "The ID of the instance to get details for"],
    aws_region: Annotated[
        str, "The AWS region to get instance details in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Get the details of the specified ec2 instance in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.get_instance_details(instance_id)


@mcp.tool()
async def search_instances_by_name(
    instance_name: Annotated[str, "The name of the instance to search for"],
    aws_region: Annotated[
        str, "The AWS region to search instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Search for instances by name in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.search_instances_by_tag(tag_key="Name", tag_value=instance_name)


@mcp.tool()
async def search_instances_by_tag(
    tag_key: Annotated[str, "The key of the tag to search for"],
    tag_value: Annotated[str, "The value of the tag to search for"],
    aws_region: Annotated[
        str, "The AWS region to search instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Search for instances by tag in the specified AWS region."""
    service = get_aws_service(aws_region)
    return service.search_instances_by_tag(tag_key, tag_value)


def serve():
    """Run the FastMCP server."""
    mcp.run()


if __name__ == "__main__":
    serve()
