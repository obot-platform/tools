package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ChangeInstanceState(ctx context.Context, client *ec2.Client, instanceId, state string) error {
	switch state {
	case "start":
		params := &ec2.StartInstancesInput{
			InstanceIds: []string{instanceId},
		}
		_, err := client.StartInstances(ctx, params)
		if err != nil {
			return err
		}
		fmt.Println("Sent command to start instance ", instanceId)
	case "stop":
		params := &ec2.StopInstancesInput{
			InstanceIds: []string{instanceId},
		}
		_, err := client.StopInstances(ctx, params)
		if err != nil {
			return err
		}
		fmt.Println("Sent command to stop instance ", instanceId)
	case "reboot":
		params := &ec2.RebootInstancesInput{
			InstanceIds: []string{instanceId},
		}
		_, err := client.RebootInstances(ctx, params)
		if err != nil {
			return err
		}
		fmt.Println("Sent command to reboot instance ", instanceId)
	case "terminate":
		params := &ec2.TerminateInstancesInput{
			InstanceIds: []string{instanceId},
		}
		_, err := client.TerminateInstances(ctx, params)
		if err != nil {
			return err
		}
		fmt.Println("Sent command to terminate instance ", instanceId)
	default:
		return fmt.Errorf("Invalid state: %s", state)
	}

	return nil
}
