package client

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/global"
)

// StaticTokenCredential is taken from https://github.com/gptscript-ai/mail-assistant/blob/10944805801bbb6f71eccefd1bea5f114fded164/pkg/mstoken/auth.go
type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

func NewClient(scopes []string) (*msgraphsdkgo.GraphServiceClient, error) {
	return msgraphsdkgo.NewGraphServiceClientWithCredentials(StaticTokenCredential{
		token: os.Getenv(global.CredentialEnv),
	}, scopes)
}
