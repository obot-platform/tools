package commands

import (
	"encoding/json"
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

func ListUserProjects(client *gitlab.Client) error {
	opts := &gitlab.ListProjectsOptions{
		Membership: gitlab.Ptr(true),
		OrderBy:    gitlab.Ptr("name"),
		Sort:       gitlab.Ptr("asc"),
	}

	projects, _, err := client.Projects.ListProjects(opts)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("No projects found for the current user.")
		return nil
	}

	// Convert projects to JSON for output
	projectsOutput := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		projectData := map[string]interface{}{
			"id":                  project.ID,
			"name":                project.Name,
			"path_with_namespace": project.PathWithNamespace,
			"description":         project.Description,
			"web_url":             project.WebURL,
			"visibility":          project.Visibility,
			"created_at":          project.CreatedAt,
			"last_activity_at":    project.LastActivityAt,
		}
		projectsOutput = append(projectsOutput, projectData)
	}

	outputJSON, err := json.MarshalIndent(projectsOutput, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(outputJSON))
	return nil
}
