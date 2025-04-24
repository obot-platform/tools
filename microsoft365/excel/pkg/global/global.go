package global

const CredentialEnv = "GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"

var (
	ReadOnlyScopes = []string{"Files.Read", "User.Read"}
	AllScopes      = []string{"Files.Read", "Files.ReadWrite", "User.Read"}
)
