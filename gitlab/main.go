package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gptscript-ai/tools/gitlab/pkg/commands"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		exit("Usage: gitlab <command>", errors.New("invalid arguments"))
	}

	command := os.Args[1]

	git, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		exit("Failed to create gitlab client", err)
	}
	user, _, err := git.Users.CurrentUser()
	if err != nil {
		exit("Error getting current user. This could potentially be an authentication issue", err)
	}

	switch command {
	// Issues
	case "queryIssues":
		rawQuery := os.Getenv("ISSUE_QUERY")
		var query map[string]any
		if rawQuery != "" {
			query, err = parseStringToMap(rawQuery)
			if err != nil {
				exit("Failed to parse expected format of issue_query parameter", err)
			}
		}

		if err := commands.QueryIssues(git, query); err != nil {
			exit("Unable to query Issues", err)
		}
	case "listProjectIssues":
		rawQuery := os.Getenv("ISSUE_QUERY")
		var query map[string]any
		if rawQuery != "" {
			query, err = parseStringToMap(rawQuery)
			if err != nil {
				exit("Failed to parse expected format of issue_query parameter", err)
			}
		}

		if err := commands.ListProjectIssues(git, os.Getenv("PROJECT_ID"), query); err != nil {
			exit("Unable to list Project Issues", err)
		}
	case "listMyAssignedIssues":
		query := map[string]any{
			"assignee_username": user.Username,
		}
		if err := commands.QueryIssues(git, query); err != nil {
			exit("Unable to query Issues", err)
		}
	case "listMyCreatedIssues":
		query := map[string]any{
			"author_id": user.ID,
		}
		if err := commands.QueryIssues(git, query); err != nil {
			exit("Unable to query Issues", err)
		}
	case "getIssueDetails":
		projectID := os.Getenv("PROJECT_ID")
		issueIID, err := strconv.Atoi(os.Getenv("ISSUE_IID"))
		if err != nil {
			exit("Failed to parse issue_iid parameter as an integer", err)
		}
		if err := commands.GetIssueDetails(git, projectID, issueIID); err != nil {
			exit("Unable to get Issue details", err)
		}
	case "createIssue":
		projectID := os.Getenv("PROJECT_ID")
		issueName := os.Getenv("ISSUE_NAME")
		issueDescription := os.Getenv("ISSUE_DESCRIPTION")
		if projectID == "" || issueName == "" {
			exit("Required parameters are not set", errors.New("project and issue_name parameters are not set"))
		}
		if err := commands.CreateIssue(git, projectID, issueName, issueDescription); err != nil {
			exit("Unable to create Issue", err)
		}
	case "updateIssue":
		projectID := os.Getenv("PROJECT_ID")
		issueID := os.Getenv("ISSUE_IID")
		updatesMap, err := parseStringToMap(os.Getenv("UPDATES"))
		if err != nil {
			exit("Failed to parse expected format of updates parameter", err)
		}

		if err := commands.UpdateIssue(git, projectID, issueID, updatesMap); err != nil {
			exit("Unable to update Issue", err)
		}
	case "deleteIssue":
		projectID := os.Getenv("PROJECT_ID")
		issueID := os.Getenv("ISSUE_IID")
		if err := commands.DeleteIssue(git, projectID, issueID); err != nil {
			exit("Unable to delete Issue", err)
		}
	case "addIssueComment":
		projectID := os.Getenv("PROJECT_ID")
		issueID := os.Getenv("ISSUE_IID")
		comment := os.Getenv("COMMENT")
		if err := commands.AddIssueComment(git, projectID, issueID, comment); err != nil {
			exit("Unable to add Issue Comment", err)
		}

	// Merge Requests
	case "queryMergeRequests":
		query, err := parseStringToMap(os.Getenv("MR_QUERY"))
		if err != nil {
			exit("Failed to parse mr_query parameter:", err)
		}
		if err := commands.QueryMergeRequests(git, query); err != nil {
			exit("Unable to query Merge Requests", err)
		}
	case "listProjectMergeRequests":
		query, err := parseStringToMap(os.Getenv("MR_QUERY"))
		if err != nil {
			exit("Failed to parse expected format of mr_query parameter", err)
		}
		if err := commands.ListProjectMergeRequests(git, os.Getenv("PROJECT_ID"), query); err != nil {
			exit("Unable to list Project Merge Requests", err)
		}
	case "listMyOpenedMergeRequests":
		query := map[string]any{
			"author_id": user.ID,
		}
		if err := commands.QueryMergeRequests(git, query); err != nil {
			exit("Unable to query Merge Requests", err)
		}
	case "listMyAssignedMergeRequests":
		query := map[string]any{
			"assignee_username": user.Username,
		}
		if err := commands.QueryMergeRequests(git, query); err != nil {
			exit("Unable to query Merge Requests", err)
		}
	case "getMergeRequestDetails":
		projectID := os.Getenv("PROJECT_ID")
		mrID, err := strconv.Atoi(os.Getenv("MR_ID"))
		if err != nil {
			exit("Failed to parse mr_id parameter as an integer", err)
		}
		if err := commands.GetMergeRequestDetails(git, projectID, mrID); err != nil {
			exit("Unable to get Merge Request details", err)
		}
	case "createMergeRequest":
		projectID := os.Getenv("PROJECT_ID")
		mrName := os.Getenv("MR_NAME")
		mrDescription := os.Getenv("MR_DESCRIPTION")
		mrSourceBranch := os.Getenv("MR_SOURCE_BRANCH")
		mrTargetBranch := os.Getenv("MR_TARGET_BRANCH")
		if projectID == "" || mrName == "" || mrSourceBranch == "" || mrTargetBranch == "" {
			exit("Parameters not set", errors.New("project, mr_name, mr_source_branch, mr_target_branch parameters are all required and must be set"))
		}
		if err := commands.CreateMergeRequest(git, projectID, mrName, mrDescription, mrSourceBranch, mrTargetBranch); err != nil {
			exit("Unable to create Merge Request", err)

		}
	case "updateMergeRequest":
		projectID := os.Getenv("PROJECT_ID")
		MRID := os.Getenv("MR_ID")
		updatesMap, err := parseStringToMap(os.Getenv("UPDATES"))
		if err != nil {
			exit("Failed to parse UPDATES", err)
		}

		if err := commands.UpdateMergeRequest(git, projectID, MRID, updatesMap); err != nil {
			exit("Failed to update Merge Request", err)
		}
	case "deleteMergeRequest":
		projectID := os.Getenv("PROJECT_ID")
		MRID := os.Getenv("MR_ID")
		if err := commands.DeleteMergeRequest(git, projectID, MRID); err != nil {
			exit("Failed to delete Merge Request", err)
		}
	case "addMergeRequestComment":
		projectID := os.Getenv("PROJECT_ID")
		MRID := os.Getenv("MR_ID")
		comment := os.Getenv("COMMENT")
		if err := commands.AddMergeRequestComment(git, projectID, MRID, comment); err != nil {
			exit("Failed to add Merge Request comment", err)
		}
	case "approveMergeRequest":
		projectID := os.Getenv("PROJECT_ID")
		MRID := os.Getenv("MR_ID")

		if err := commands.ApproveMergeRequest(git, projectID, MRID); err != nil {
			exit("Failed to approve merge request", err)
		}

	// Utility
	case "lookupUserID":
		username := os.Getenv("GITLAB_USERNAME")
		if username == "" {
			exit("Username is required", errors.New("gitlab_username parameter must be set"))
		}
		if err := commands.LookupUserID(git, username); err != nil {
			exit(fmt.Sprintf("Failed to lookup user %s", username), err)
		}
	default:
		exit(fmt.Sprintf("Unknown command"), errors.New(command))
	}
}

func exit(message string, err error) {
	log.Fatalf("%s: %v", message, err)
}

func parseStringToMap(raw string) (map[string]any, error) {
	var transformedMap map[string]any
	err := json.Unmarshal([]byte(raw), &transformedMap)
	if err != nil {
		return nil, err
	}
	return transformedMap, nil
}
