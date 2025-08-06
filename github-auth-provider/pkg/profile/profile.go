package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/obot-platform/tools/auth-providers-common/pkg/state"
)

var githubBaseURL = "https://api.github.com"

type githubUserProfile struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Company           string `json:"company"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
	Hireable          bool   `json:"hireable"`
	Bio               string `json:"bio"`
	TwitterUsername   string `json:"twitter_username"`
	PublicRepos       int    `json:"public_repos"`
	PublicGists       int    `json:"public_gists"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type githubOrganization struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type githubTeam struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Organization githubOrganization `json:"organization"`
}

func FetchUserProfile(ctx context.Context, accessToken string) (*githubUserProfile, error) {
	var result githubUserProfile
	err := makeGitHubRequest(ctx, accessToken, "user", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func FetchUserGroupInfos(ctx context.Context, accessToken string) (state.GroupInfoList, error) {
	var infos state.GroupInfoList

	var orgs []githubOrganization
	err := makeGitHubRequest(ctx, accessToken, "user/orgs", &orgs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user organizations: %w", err)
	}
	for _, org := range orgs {
		infos = append(infos, state.GroupInfo{
			ID:      fmt.Sprintf("github/org/%d", org.ID),
			Name:    org.Login,
			IconURL: &org.AvatarURL,
		})
	}

	var teams []githubTeam
	err = makeGitHubRequest(ctx, accessToken, "user/teams", &teams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user teams: %w", err)
	}
	for _, team := range teams {
		infos = append(infos, state.GroupInfo{
			ID:      fmt.Sprintf("github/org/%d/team/%d", team.Organization.ID, team.ID),
			Name:    team.Name,
			IconURL: &team.Organization.AvatarURL,
		})
	}

	// Sort groups by ID lexicographically
	slices.SortFunc(infos, func(a, b state.GroupInfo) int {
		return strings.Compare(a.ID, b.ID)
	})

	return infos, nil
}

func makeGitHubRequest(ctx context.Context, accessToken, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", githubBaseURL, path), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, body)
	}

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}
