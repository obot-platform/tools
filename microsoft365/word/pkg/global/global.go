package global

const CredentialEnv = "GPTSCRIPT_MICROSOFT_WORD_TOKEN"

var (
	ReadOnlyScopes  = []string{"Files.Read", "Files.Read.All", "User.Read"}
	ReadWriteScopes = []string{"Files.ReadWrite", "Files.ReadWrite.All", "User.Read", "Sites.ReadWrite.All"}
)
