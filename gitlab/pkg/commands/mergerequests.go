package commands

import (
	"encoding/json"
	"fmt"
	"gitlab.com/gitlab-org/api/client-go"
	"log"
	"strconv"
	"strings"
)

func buildListMergeRequestsOptions(query map[string]any) (gitlab.ListMergeRequestsOptions, error) {
	var opts gitlab.ListMergeRequestsOptions
	listOptions := gitlab.ListOptions{
		Pagination: "keyset",
		PerPage:    20,
		OrderBy:    "updated_at",
		Sort:       "desc",
	}

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return gitlab.ListMergeRequestsOptions{}, err
	}
	err = json.Unmarshal(jsonQuery, &opts)
	if err != nil {
		return gitlab.ListMergeRequestsOptions{}, err
	}
	opts.ListOptions = listOptions

	return opts, nil
}

func buildListProjectMergeRequestsOptions(query map[string]any) (gitlab.ListProjectMergeRequestsOptions, error) {
	var opts gitlab.ListProjectMergeRequestsOptions
	listOptions := gitlab.ListOptions{
		Pagination: "keyset",
		PerPage:    20,
		OrderBy:    "updated_at",
		Sort:       "desc",
	}

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return gitlab.ListProjectMergeRequestsOptions{}, err
	}
	err = json.Unmarshal(jsonQuery, &opts)
	if err != nil {
		return gitlab.ListProjectMergeRequestsOptions{}, err
	}
	opts.ListOptions = listOptions

	return opts, nil
}

func buildUpdateMergeRequestOptions(updates map[string]any) (gitlab.UpdateMergeRequestOptions, error) {
	var opts gitlab.UpdateMergeRequestOptions
	jsonUpdates, err := json.Marshal(updates)
	if err != nil {
		return gitlab.UpdateMergeRequestOptions{}, err
	}
	err = json.Unmarshal(jsonUpdates, &opts)
	if err != nil {
		return gitlab.UpdateMergeRequestOptions{}, err
	}
	return opts, nil
}

func QueryMergeRequests(client *gitlab.Client, query map[string]any) error {
	opts, err := buildListMergeRequestsOptions(query)
	if err != nil {
		return err
	}
	options := []gitlab.RequestOptionFunc{}

	for {
		mrs, resp, err := client.MergeRequests.ListMergeRequests(&opts, options...)
		if err != nil {
			return err
		}

		for _, mr := range mrs {
			err := listMergeRequestOutput(client, mr)
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

func ListProjectMergeRequests(client *gitlab.Client, projectID string, query map[string]any) error {
	opts, err := buildListProjectMergeRequestsOptions(query)
	if err != nil {
		return err
	}
	options := []gitlab.RequestOptionFunc{}

	for {
		mrs, resp, err := client.MergeRequests.ListProjectMergeRequests(projectID, &opts, options...)
		if err != nil {
			return err
		}

		for _, mr := range mrs {
			err := listMergeRequestOutput(client, mr)
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

func GetMergeRequestDetails(client *gitlab.Client, projectID string, mrID int) error {
	this, _, err := client.MergeRequests.GetMergeRequest(projectID, mrID, nil)
	if err != nil {
		return err
	}
	if err := listMergeRequestOutput(client, this); err != nil {
		return err
	}

	notes, _, err := client.Notes.ListMergeRequestNotes(projectID, mrID, &gitlab.ListMergeRequestNotesOptions{})
	if err != nil {
		log.Fatalf("Failed to fetch merge request notes: %v", err)
	}
	if this.ClosedAt != nil {
		fmt.Printf("  Closed At: %s\n", this.ClosedAt)
		fmt.Printf("  Closed By: %s\n", this.ClosedBy)
	}
	fmt.Printf("  Locked:  %t\n", this.DiscussionLocked)
	fmt.Printf("  Downvotes: %d\n", this.Downvotes)
	if len(this.Labels) > 0 {
		labelString := strings.Join(this.Labels, ", ")
		fmt.Printf("  Labels: [%s]\n", labelString)
	}
	if len(notes) > 0 {
		fmt.Println("  Merge Request Activity:")
		for _, note := range notes {
			fmt.Printf("\n  Author: %s\n", note.Author.Username)
			fmt.Printf("  Created At: %s\n", note.CreatedAt.String())
			body := IndentString(note.Body)
			fmt.Printf("  Body:\n%s\n", body)
		}
	}
	return nil
}

func CreateMergeRequest(client *gitlab.Client, projectID, mrName, mrDescription, mrSourceBranch, mrTargetBranch string) error {
	opts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(mrName),
		Description:  gitlab.Ptr(mrDescription),
		SourceBranch: gitlab.Ptr(mrSourceBranch),
		TargetBranch: gitlab.Ptr(mrTargetBranch),
	}
	mr, _, err := client.MergeRequests.CreateMergeRequest(projectID, opts, nil)
	if err != nil {
		return err
	}
	project, _, err := client.Projects.GetProject(mr.ProjectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Created Merge Request in Project '%s' (ID: %d) with Merge Request IID %d\n", project.PathWithNamespace, mr.ProjectID, mr.IID)
	return nil
}

func UpdateMergeRequest(client *gitlab.Client, projectID, MRID string, updates map[string]any) error {
	opts, err := buildUpdateMergeRequestOptions(updates)
	if err != nil {
		return err
	}
	intMRID, err := strconv.Atoi(MRID)
	if err != nil {
		return err
	}
	mr, _, err := client.MergeRequests.UpdateMergeRequest(projectID, intMRID, &opts, nil)
	if err != nil {
		return err
	}

	project, _, err := client.Projects.GetProject(mr.ProjectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Updated Merge Request with ID %d in Project '%s' (ID: %d)\n", mr.IID, project.PathWithNamespace, mr.ProjectID)
	return nil
}

func DeleteMergeRequest(client *gitlab.Client, projectID, MRID string) error {
	intMRID, err := strconv.Atoi(MRID)
	if err != nil {
		return err
	}
	if _, err := client.MergeRequests.DeleteMergeRequest(projectID, intMRID); err != nil {
		return err
	}

	project, _, err := client.Projects.GetProject(projectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted Merge Request with ID %s in Project '%s' (ID: %d)\n", MRID, project.PathWithNamespace, project.ID)
	return nil
}

func AddMergeRequestComment(client *gitlab.Client, projectID, MRID, comment string) error {
	intIssueID, err := strconv.Atoi(MRID)
	if err != nil {
		return err
	}
	opt := &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr(comment),
	}
	note, _, err := client.Notes.CreateMergeRequestNote(projectID, intIssueID, opt, nil)
	if err != nil {
		return err
	}
	project, _, err := client.Projects.GetProject(projectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Added Merge Request comment (ID: %d) to Merge Request with ID %s in Project '%s' (ID: %d).\n", note.ID, MRID, project.PathWithNamespace, project.ID)
	return nil
}

func ApproveMergeRequest(client *gitlab.Client, projectID, MRID string) error {
	intMRID, err := strconv.Atoi(MRID)
	if err != nil {
		return err
	}
	if _, _, err := client.MergeRequestApprovals.ApproveMergeRequest(projectID, intMRID, nil, nil); err != nil {
		return err
	}

	project, _, err := client.Projects.GetProject(projectID, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Approved Merge Request with ID %s in Project '%s' (ID: %d)\n", MRID, project.PathWithNamespace, project.ID)
	return nil
}

func listMergeRequestOutput(client *gitlab.Client, mr *gitlab.MergeRequest) error {
	project, _, err := client.Projects.GetProject(mr.ProjectID, nil)
	if err != nil {
		return err
	}
	fmt.Printf("* Merge Request #%d: %s\n", mr.IID, mr.Title)
	fmt.Printf("  Merge Request ID: %d\n", mr.ID)
	fmt.Printf("  Project Name: %s\n", project.Name)
	fmt.Printf("  Project ID: %d\n", mr.ProjectID)
	fmt.Printf("  State: %s\n", mr.State)
	fmt.Printf("  Author Username: %s\n", mr.Author.Username)
	fmt.Printf("  Author ID: %d\n", mr.Author.ID)
	if len(mr.Assignees) < 1 {
		fmt.Println("  Assignee(s):")
		for _, assignee := range mr.Assignees {
			fmt.Printf("    * %s (ID: %d)\n", assignee.Username, assignee.ID)
		}
	} else if mr.Assignee != nil {
		fmt.Printf("  Assignee: %s (ID: %d)\n", mr.Assignee.Username, mr.Assignee.ID)
	} else {
		fmt.Println("  Assignee(s): N/A")
	}

	indentedDescription := IndentString(mr.Description)
	fmt.Printf("  Description:\n%s\n", indentedDescription)
	return nil
}
