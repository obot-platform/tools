package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/gitlab-org/api/client-go"
	"strconv"
	"strings"
)

func buildListIssuesOptions(query map[string]any) (gitlab.ListIssuesOptions, error) {
	var opts gitlab.ListIssuesOptions
	listOptions := gitlab.ListOptions{
		Pagination: "keyset",
		PerPage:    20,
		OrderBy:    "updated_at",
		Sort:       "desc",
	}

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return gitlab.ListIssuesOptions{}, err
	}
	err = json.Unmarshal(jsonQuery, &opts)
	if err != nil {
		return gitlab.ListIssuesOptions{}, err
	}
	opts.ListOptions = listOptions

	return opts, nil
}

func buildListProjectIssuesOptions(query map[string]any) (gitlab.ListProjectIssuesOptions, error) {
	var opts gitlab.ListProjectIssuesOptions
	listOptions := gitlab.ListOptions{
		Pagination: "keyset",
		PerPage:    20,
		OrderBy:    "updated_at",
		Sort:       "desc",
	}

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return gitlab.ListProjectIssuesOptions{}, err
	}
	err = json.Unmarshal(jsonQuery, &opts)
	if err != nil {
		return gitlab.ListProjectIssuesOptions{}, err
	}
	opts.ListOptions = listOptions

	return opts, nil
}

func buildUpdateIssueOptions(updates map[string]any) (gitlab.UpdateIssueOptions, error) {
	var opts gitlab.UpdateIssueOptions
	jsonUpdates, err := json.Marshal(updates)
	if err != nil {
		return gitlab.UpdateIssueOptions{}, err
	}
	err = json.Unmarshal(jsonUpdates, &opts)
	if err != nil {
		return gitlab.UpdateIssueOptions{}, err
	}
	return opts, nil
}

func QueryIssues(client *gitlab.Client, query map[string]any) error {
	opts, err := buildListIssuesOptions(query)
	if err != nil {
		return err
	}
	options := []gitlab.RequestOptionFunc{}

	for {
		issues, resp, err := client.Issues.ListIssues(&opts, options...)
		if err != nil {
			return err
		}

		if len(issues) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		for _, issue := range issues {
			err := listIssueOutput(client, issue)
			if err != nil {
				return err
			}
		}

		if resp.NextLink == "" {
			break
		}

		options = []gitlab.RequestOptionFunc{
			gitlab.WithKeysetPaginationParameters(resp.NextLink),
		}

	}
	return nil
}

func ListProjectIssues(client *gitlab.Client, projectID string, query map[string]any) error {
	opts, err := buildListProjectIssuesOptions(query)
	if err != nil {
		return err
	}
	options := []gitlab.RequestOptionFunc{}

	for {
		issues, resp, err := client.Issues.ListProjectIssues(projectID, &opts, options...)
		if err != nil {
			return err
		}

		for _, issue := range issues {
			err := listIssueOutput(client, issue)
			if err != nil {
				return err
			}
		}

		if resp.NextLink == "" {
			break
		}

		options = []gitlab.RequestOptionFunc{
			gitlab.WithKeysetPaginationParameters(resp.NextLink),
		}

	}
	return nil
}

func GetIssueDetails(client *gitlab.Client, projectID string, issueIID int) error {
	this, _, err := client.Issues.GetIssue(projectID, issueIID)
	if err != nil {
		return err
	}
	if err := listIssueOutput(client, this); err != nil {
		return err
	}

	notes, _, err := client.Notes.ListIssueNotes(projectID, issueIID, &gitlab.ListIssueNotesOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to fetch issue notes: %v", err))
	}
	if this.ClosedAt != nil {
		fmt.Printf("  Closed At: %s\n", this.ClosedAt)
		fmt.Printf("  Closed By: %s\n", this.ClosedBy)
	}
	fmt.Printf("  Locked:  %t\n", this.DiscussionLocked)
	fmt.Printf("  Confidential: %t\n", this.Confidential)
	fmt.Printf("  Downvotes: %d\n", this.Downvotes)
	if this.DueDate != nil {
		fmt.Printf("  Due Date: %s\n", this.DueDate.String())
	}
	if len(this.Labels) > 0 {
		labelString := strings.Join(this.Labels, ", ")
		fmt.Printf("  Labels: [%s]\n", labelString)
	}
	if len(notes) > 0 {
		fmt.Println("  Issue Activity:")
		for _, note := range notes {
			fmt.Printf("\n  Author: %s\n", note.Author.Username)
			fmt.Printf("  Created At: %s\n", note.CreatedAt.String())
			body := IndentString(note.Body)
			fmt.Printf("  Body:\n%s\n", body)
		}
	}
	return nil
}

func CreateIssue(client *gitlab.Client, projectID, issueName, issueDescription string) error {
	opts := &gitlab.CreateIssueOptions{
		Title:       gitlab.Ptr(issueName),
		Description: gitlab.Ptr(issueDescription),
	}
	issue, _, err := client.Issues.CreateIssue(projectID, opts, nil)
	if err != nil {
		return err
	}
	project, _, err := client.Projects.GetProject(issue.ProjectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Created issue in Project '%s' (ID: %d) with Issue IID %d\n", project.PathWithNamespace, issue.ProjectID, issue.IID)
	return nil
}

func UpdateIssue(client *gitlab.Client, projectID, issueID string, updates map[string]any) error {
	opts, err := buildUpdateIssueOptions(updates)
	if err != nil {
		return err
	}
	intIssueID, err := strconv.Atoi(issueID)
	if err != nil {
		return err
	}
	issue, _, err := client.Issues.UpdateIssue(projectID, intIssueID, &opts, nil)
	if err != nil {
		return err
	}

	project, _, err := client.Projects.GetProject(issue.ProjectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Updated issue with ID %d in Project '%s' (ID: %d)\n", issue.IID, project.PathWithNamespace, issue.ProjectID)
	return nil
}

func DeleteIssue(client *gitlab.Client, projectID, issueID string) error {
	intIssueID, err := strconv.Atoi(issueID)
	if err != nil {
		return err
	}
	if _, err := client.Issues.DeleteIssue(projectID, intIssueID); err != nil {
		return err
	}

	project, _, err := client.Projects.GetProject(projectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted issue with ID %s in Project '%s' (ID: %d)\n", issueID, project.PathWithNamespace, project.ID)
	return nil
}

func AddIssueComment(client *gitlab.Client, projectID, issueID, comment string) error {
	intIssueID, err := strconv.Atoi(issueID)
	if err != nil {
		return err
	}
	opt := &gitlab.CreateIssueNoteOptions{
		Body: gitlab.Ptr(comment),
	}
	note, _, err := client.Notes.CreateIssueNote(projectID, intIssueID, opt, nil)
	if err != nil {
		return err
	}
	project, _, err := client.Projects.GetProject(projectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Added issue comment (ID: %d) to Issue with ID %s in Project '%s' (ID: %d).\n", note.ID, issueID, project.PathWithNamespace, project.ID)
	return nil
}

func listIssueOutput(client *gitlab.Client, issue *gitlab.Issue) error {
	project, _, err := client.Projects.GetProject(issue.ProjectID, nil)
	if err != nil {
		return err
	}
	fmt.Printf("* Issue #%d: %s\n", issue.IID, issue.Title)
	fmt.Printf("  Issue ID: %d\n", issue.ID)
	fmt.Printf("  Project Name: %s\n", project.Name)
	fmt.Printf("  Project ID: %d\n", issue.ProjectID)
	fmt.Printf("  State: %s\n", issue.State)
	fmt.Printf("  Author Username: %s\n", issue.Author.Username)
	fmt.Printf("  Author ID: %d\n", issue.Author.ID)
	if len(issue.Assignees) < 1 {
		fmt.Println("  Assignee(s):")
		for _, assignee := range issue.Assignees {
			fmt.Printf("    * %s (ID: %d)\n", assignee.Username, assignee.ID)
		}
	} else if issue.Assignee != nil {
		fmt.Printf("  Assignee: %s (ID: %d)\n", issue.Assignee.Username, issue.Assignee.ID)
	} else {
		fmt.Println("  Assignee(s): N/A")
	}

	indentedDescription := IndentString(issue.Description)
	fmt.Printf("  Description:\n%s\n", indentedDescription)
	return nil
}
