package global

const CredentialEnv = "GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"

var (
	ReadOnlyScopes = []string{"Contacts.Read", "Contacts.Read.Shared", "User.Read", "MailboxSettings.Read"}
	AllScopes      = []string{"Contacts.Read", "Contacts.Read.Shared", "Contacts.ReadWrite", "Contacts.ReadWrite.Shared", "User.Read", "MailboxSettings.Read", "MailboxSettings.ReadWrite"}
)
