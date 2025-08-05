from typing import Any, Dict, List, Optional

import boto3
from boto3.session import Session


class AWSService:
    def __init__(
        self, session: Optional[Session] = None, region_name: str = "us-east-1"
    ):
        """Initialize AWS service with optional session for testing."""
        self.session = session or boto3
        self.region_name = region_name
        self._ec2_client = None

    @property
    def ec2_client(self):
        """Lazy initialization of EC2 client."""
        if not self._ec2_client:
            self._ec2_client = self.session.client("ec2", region_name=self.region_name)
        return self._ec2_client

    def list_instances(self) -> Dict[str, List[Dict[str, Any]]]:
        """List all instances in the region."""
        try:
            response = self.ec2_client.describe_instances()
            instances = []
            for reservation in response["Reservations"]:
                for instance in reservation["Instances"]:
                    # Skip terminated instances
                    if instance["State"]["Name"] == "terminated":
                        continue

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
            return {"instances": instances}
        except Exception as e:
            return {"error": f"Failed to list instances: {str(e)}"}

    def start_instance(self, instance_id: str) -> Dict[str, str]:
        """Start an EC2 instance."""
        try:
            self.ec2_client.start_instances(InstanceIds=[instance_id])
            return {"message": f"Instance {instance_id} started"}
        except Exception as e:
            return {"error": f"Failed to start instance: {str(e)}"}

    def stop_instance(self, instance_id: str) -> Dict[str, str]:
        """Stop an EC2 instance."""
        try:
            self.ec2_client.stop_instances(InstanceIds=[instance_id])
            return {"message": f"Instance {instance_id} stopped"}
        except Exception as e:
            return {"error": f"Failed to stop instance: {str(e)}"}

    def terminate_instance(self, instance_id: str) -> Dict[str, str]:
        """Terminate an EC2 instance."""
        try:
            self.ec2_client.terminate_instances(InstanceIds=[instance_id])
            return {"message": f"Instance {instance_id} terminated"}
        except Exception as e:
            return {"error": f"Failed to terminate instance: {str(e)}"}

    def reboot_instance(self, instance_id: str) -> Dict[str, str]:
        """Reboot an EC2 instance."""
        try:
            self.ec2_client.reboot_instances(InstanceIds=[instance_id])
            return {"message": f"Instance {instance_id} rebooted"}
        except Exception as e:
            return {"error": f"Failed to reboot instance: {str(e)}"}

    def create_instance(
        self,
        instance_type: str,
        image_id: str,
        subnet_id: str,
        key_name: Optional[str] = None,
        instance_name: Optional[str] = None,
        tags: Optional[Dict[str, str]] = None,
        user_data: Optional[str] = None,
        security_group_ids: List[str] = None,
    ) -> Dict[str, str]:
        """Create a new EC2 instance."""
        try:
            params = {
                "InstanceType": instance_type,
                "ImageId": image_id,
                "SubnetId": subnet_id,
                "MinCount": 1,
                "MaxCount": 1,
            }

            if security_group_ids:
                params["SecurityGroupIds"] = security_group_ids
            if key_name:
                params["KeyName"] = key_name
            if user_data:
                params["UserData"] = user_data
            if instance_name or tags:
                tag_specs = []
                if instance_name:
                    tag_specs.append({"Key": "Name", "Value": instance_name})
                if tags:
                    tag_specs.extend([{"Key": k, "Value": v} for k, v in tags.items()])
                params["TagSpecifications"] = [
                    {"ResourceType": "instance", "Tags": tag_specs}
                ]

            instance = self.ec2_client.run_instances(**params)
            return {
                "message": f"Instance {instance['Instances'][0]['InstanceId']} created"
            }
        except Exception as e:
            return {"error": f"Failed to create instance: {str(e)}"}

    def list_vpcs(self) -> Dict[str, List[Dict[str, Any]]]:
        """List all VPCs in the region."""
        try:
            response = self.ec2_client.describe_vpcs()
            return {"vpcs": response["Vpcs"]}
        except Exception as e:
            return {"error": f"Failed to list VPCs: {str(e)}"}

    def list_subnets(self) -> Dict[str, List[Dict[str, Any]]]:
        """List all subnets in the region."""
        try:
            response = self.ec2_client.describe_subnets()
            return {"subnets": response["Subnets"]}
        except Exception as e:
            return {"error": f"Failed to list subnets: {str(e)}"}

    def list_security_groups(self, vpc_id: str) -> Dict[str, List[Dict[str, Any]]]:
        """List all security groups in the specified VPC."""
        try:
            response = self.ec2_client.describe_security_groups(
                Filters=[{"Name": "vpc-id", "Values": [vpc_id]}]
            )
            return {"security_groups": response["SecurityGroups"]}
        except Exception as e:
            return {"error": f"Failed to list security groups: {str(e)}"}

    def get_instance_details(self, instance_id: str) -> Dict[str, Any]:
        """Get details of an EC2 instance."""
        try:
            response = self.ec2_client.describe_instances(InstanceIds=[instance_id])
            return response["Reservations"][0]["Instances"][0]
        except Exception as e:
            return {"error": f"Failed to get instance details: {str(e)}"}

    def search_instances_by_tag(
        self, tag_key: str, tag_value: str
    ) -> Dict[str, List[Dict[str, Any]]]:
        """Search for instances by tag."""
        try:
            response = self.ec2_client.describe_instances(
                Filters=[{"Name": f"tag:{tag_key}", "Values": [tag_value]}]
            )
            return response["Reservations"]
        except Exception as e:
            return {"error": f"Failed to search instances by tag: {str(e)}"}
