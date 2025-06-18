package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker-credential-helpers/credentials"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/storage/value"
)

// uid is here to fulfill the value.Context interface for the transformer.
// This is similar to authenticatedDataString from the k8s apiserver's storage interface
// for etcd: https://github.com/kubernetes/kubernetes/blob/a42f4f61c2c46553bfe338eefe9e81818c7360b4/staging/src/k8s.io/apiserver/pkg/storage/etcd3/store.go#L63
type uid string

var migrate = os.Getenv("OBOT_TOOLS_MIGRATE_DB") != "false"

func (u uid) AuthenticatedData() []byte {
	return []byte(u)
}

var groupResource = schema.GroupResource{
	Group:    "", // deliberately left empty
	Resource: "credentials",
}

type Database struct {
	db          *gorm.DB
	transformer value.Transformer
}

func NewDatabase(ctx context.Context, db *gorm.DB) (Database, error) {
	if migrate {
		if err := db.AutoMigrate(&GptscriptCredential{}); err != nil {
			return Database{}, fmt.Errorf("failed to auto migrate GptscriptCredential: %w", err)
		}
		if err := migrateContext(db); err != nil {
			return Database{}, fmt.Errorf("failed to migrate context: %w", err)
		}
	}

	encryptionConf, err := readEncryptionConfig(ctx)
	if err != nil {
		return Database{}, fmt.Errorf("failed to read encryption config: %w", err)
	} else if encryptionConf != nil {
		transformer, exists := encryptionConf.Transformers[groupResource]
		if !exists {
			return Database{}, fmt.Errorf("failed to find encryption transformer for %s", groupResource.String())
		}
		return Database{
			db:          db,
			transformer: transformer,
		}, nil
	}

	return Database{
		db: db,
	}, nil
}

// migrateContext populates the Context field of all credentials.
func migrateContext(db *gorm.DB) error {
	var creds []GptscriptCredential
	if err := db.Where("context IS NULL OR context = ?", "").Find(&creds).Error; err != nil {
		return fmt.Errorf("failed to find credentials with empty context: %w", err)
	}

	for _, cred := range creds {
		parts := strings.Split(cred.ServerURL, "///")
		if len(parts) != 2 {
			// Shouldn't happen, but just ignore it.
			continue
		}

		cred.Context = parts[1]
		if err := db.Save(&cred).Error; err != nil {
			return fmt.Errorf("failed to save credential: %w", err)
		}
	}

	return nil
}

type GptscriptCredential struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	ServerURL string `gorm:"unique"`
	Username  string
	Secret    string

	// Context is used to separately store the context of the credential.
	// The ServerURL field already contains the context, but we need to be able to query by it separately for performance reasons.
	Context string `gorm:"index"`
}

func (d Database) Add(creds *credentials.Credentials) error {
	parts := strings.Split(creds.ServerURL, "///")
	if len(parts) != 2 {
		return fmt.Errorf("invalid server URL: %s", creds.ServerURL)
	}

	cred := GptscriptCredential{
		ServerURL: creds.ServerURL,
		Username:  creds.Username,
		Secret:    creds.Secret,
		Context:   parts[1],
	}

	cred, err := d.encryptCred(context.Background(), cred)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// First, we need to check if a credential with this serverURL already exists.
	// If it does, delete it first.
	// This would normally happen during a credential refresh.
	var existing GptscriptCredential
	if err := d.db.Where("server_url = ?", cred.ServerURL).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get existing credential: %w", err)
		}
	} else {
		if err := d.db.Delete(&existing).Error; err != nil {
			return fmt.Errorf("failed to delete existing credential: %w", err)
		}
	}

	if err := d.db.Create(&cred).Error; err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	return nil
}

func (d Database) Delete(serverURL string) error {
	var (
		cred GptscriptCredential
		err  error
	)
	if err = d.db.Where("server_url = ?", serverURL).Delete(&cred).Error; err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	return nil
}

func (d Database) Get(serverURL string) (string, string, error) {
	var (
		cred GptscriptCredential
		err  error
	)
	if err = d.db.Where("server_url = ?", serverURL).First(&cred).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("failed to get credential: %w", err)
	}

	cred, err = d.decryptCred(context.Background(), cred)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return cred.Username, cred.Secret, nil
}

func (d Database) List() (map[string]string, error) {
	var (
		creds []GptscriptCredential
		err   error
	)
	if err = d.db.Find(&creds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	credMap := make(map[string]string)
	for _, cred := range creds {
		// No need to decrypt anything, since we don't need to access the secret.
		credMap[cred.ServerURL] = cred.Username
	}

	return credMap, nil
}

func (d Database) ListWithContexts(contexts []string) (map[string]string, error) {
	var (
		allCreds []GptscriptCredential
		err      error
	)

	for _, context := range contexts {
		if strings.Contains(context, "///") {
			return nil, fmt.Errorf("invalid context: %s", context)
		}

		var creds []GptscriptCredential
		if err = d.db.Where("context = ?", context).Find(&creds).Error; err != nil {
			return nil, fmt.Errorf("failed to list credentials: %w", err)
		}
		allCreds = append(allCreds, creds...)
	}

	credMap := make(map[string]string)
	for _, cred := range allCreds {
		// No need to decrypt anything, since we don't need to access the secret.
		credMap[cred.ServerURL] = cred.Username
	}

	return credMap, nil
}
