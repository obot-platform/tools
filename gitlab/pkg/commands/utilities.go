package commands

import (
	"fmt"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func LookupUserID(client *gitlab.Client, username string) error {
	opts := &gitlab.ListUsersOptions{
		Username: gitlab.Ptr(username),
	}
	users, _, err := client.Users.ListUsers(opts, nil)
	if err != nil {
		return err
	}
	if len(users) == 1 {
		user := users[0]
		fmt.Printf("User found with ID %d\n", user.ID)
		return nil
	}

	fmt.Println("No matching users found.")
	return nil
}
