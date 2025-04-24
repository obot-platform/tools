package global

const CredentialEnv = "GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"

var (
	ReadOnlyScopes = []string{"Mail.Read", "User.Read", "MailboxSettings.Read", "Groups.Read.All"}
	AllScopes      = []string{"Mail.Read", "Mail.ReadWrite", "Mail.Send", "User.Read", "MailboxSettings.Read", "Groups.ReadWrite.All"}
)
