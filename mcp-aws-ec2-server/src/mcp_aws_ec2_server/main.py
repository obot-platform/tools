from typing import Annotated

import boto3
from fastmcp import FastMCP

mcp = FastMCP(name="aws-ec2")


@mcp.tool()
async def list_instances(
    aws_region: Annotated[
        str, "The AWS region to list instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all instances in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)

    try:
        response = ec2_client.describe_instances()
        instances = []
        for reservation in response["Reservations"]:
            for instance in reservation["Instances"]:
                instance_name = None
                if "Tags" in instance:
                    for tag in instance["Tags"]:
                        if tag["Key"] == "Name":
                            instance_name = tag["Value"]
                            break

                instance_info = {
                    "name": instance_name or "N/A",
                    "instance_id": instance["InstanceId"],
                    "vpc_id": instance.get("VpcId", "N/A"),
                    "instance_type": instance["InstanceType"],
                    "image_id": instance["ImageId"],
                    "state": instance["State"]["Name"],
                }
                instances.append(instance_info)

    except Exception as e:
        return {"error": f"Failed to list instances: {str(e)}"}
    return {"instances": instances}


@mcp.tool()
async def start_instance(
    instance_id: Annotated[str, "The ID of the instance to start"],
    aws_region: Annotated[
        str, "The AWS region to start the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Start the specified ec2 instance in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    try:
        ec2_client.start_instances(InstanceIds=[instance_id])
    except Exception as e:
        return {"error": f"Failed to start instance: {str(e)}"}
    return {"message": f"Instance {instance_id} started"}


@mcp.tool()
async def stop_instance(
    instance_id: Annotated[str, "The ID of the instance to stop"],
    aws_region: Annotated[
        str, "The AWS region to stop the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Stop the specified ec2 instance in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    try:
        ec2_client.stop_instances(InstanceIds=[instance_id])
    except Exception as e:
        return {"error": f"Failed to stop instance: {str(e)}"}
    return {"message": f"Instance {instance_id} stopped"}


@mcp.tool()
async def terminate_instance(
    instance_id: Annotated[str, "The ID of the instance to terminate"],
    aws_region: Annotated[
        str, "The AWS region to terminate the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Terminate the specified ec2 instance in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    ec2_client.terminate_instances(InstanceIds=[instance_id])
    return {"message": f"Instance {instance_id} terminated"}


@mcp.tool()
async def reboot_instance(
    instance_id: Annotated[str, "The ID of the instance to reboot"],
    aws_region: Annotated[
        str, "The AWS region to reboot the instance in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Reboot the specified ec2 instance in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    try:
        ec2_client.reboot_instances(InstanceIds=[instance_id])
    except Exception as e:
        return {"error": f"Failed to reboot instance: {str(e)}"}
    return {"message": f"Instance {instance_id} rebooted"}


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
    ec2_client = boto3.client("ec2", region_name=aws_region)
    try:
        instance = ec2_client.run_instances(
            InstanceType=instance_type,
            ImageId=image_id,
            SubnetId=subnet_id,
            **({"SecurityGroupIds": security_group_ids} if security_group_ids else {}),
            **({"KeyName": key_name} if key_name else {}),
            **({"UserData": user_data} if user_data else {}),
            MinCount=1,
            MaxCount=1,
            **(
                {
                    "TagSpecifications": [
                        {
                            "ResourceType": "instance",
                            "Tags": [
                                *(
                                    {"Key": "Name", "Value": instance_name}
                                    if instance_name
                                    else []
                                ),
                                *(
                                    [{"Key": k, "Value": v} for k, v in tags.items()]
                                    if tags
                                    else []
                                ),
                            ],
                        }
                    ]
                }
                if instance_name or tags
                else {}
            ),
        )
    except Exception as e:
        return {"error": f"Failed to create instance: {str(e)}"}
    return {"message": f"Instance {instance['Instances'][0]['InstanceId']} created"}


@mcp.tool()
async def list_vpcs(
    aws_region: Annotated[
        str, "The AWS region to list VPCs in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all VPCs in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_vpcs()
    try:
        return {"vpcs": response["Vpcs"]}
    except Exception as e:
        return {"error": f"Failed to list VPCs: {str(e)}"}


@mcp.tool()
async def list_subnets(
    aws_region: Annotated[
        str, "The AWS region to list subnets in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all subnets in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_subnets()
    try:
        return {"subnets": response["Subnets"]}
    except Exception as e:
        return {"error": f"Failed to list subnets: {str(e)}"}


@mcp.tool()
async def list_security_groups(
    vpc_id: Annotated[str, "The ID of the VPC to list security groups in"],
    aws_region: Annotated[
        str, "The AWS region to list security groups in, defaults to us-east-1"
    ] = "us-east-1",
):
    """List all security groups in the specified AWS region and VPC."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_security_groups()
    try:
        return {"security_groups": response["SecurityGroups"]}
    except Exception as e:
        return {"error": f"Failed to list security groups: {str(e)}"}


@mcp.tool()
async def get_instance_details(
    instance_id: Annotated[str, "The ID of the instance to get details for"],
    aws_region: Annotated[
        str, "The AWS region to get instance details in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Get the details of the specified ec2 instance in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_instances(InstanceIds=[instance_id])
    try:
        return {"instance": response["Reservations"][0]["Instances"][0]}
    except Exception as e:
        return {"error": f"Failed to get instance details: {str(e)}"}


@mcp.tool()
async def search_instances_by_name(
    instance_name: Annotated[str, "The name of the instance to search for"],
    aws_region: Annotated[
        str, "The AWS region to search instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Search for instances by name in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_instances(
        Filters=[{"Name": "tag:Name", "Values": [instance_name]}]
    )
    try:
        return {"instances": response["Reservations"]}
    except Exception as e:
        return {"error": f"Failed to search instances by name: {str(e)}"}


@mcp.tool()
async def search_instances_by_tag(
    tag_key: Annotated[str, "The key of the tag to search for"],
    tag_value: Annotated[str, "The value of the tag to search for"],
    aws_region: Annotated[
        str, "The AWS region to search instances in, defaults to us-east-1"
    ] = "us-east-1",
):
    """Search for instances by tag in the specified AWS region."""
    ec2_client = boto3.client("ec2", region_name=aws_region)
    response = ec2_client.describe_instances(
        Filters=[{"Name": f"tag:{tag_key}", "Values": [tag_value]}]
    )
    try:
        return {"instances": response["Reservations"]}
    except Exception as e:
        return {"error": f"Failed to search instances by tag: {str(e)}"}


def serve():
    mcp.run()


if __name__ == "__main__":
    mcp.run()
