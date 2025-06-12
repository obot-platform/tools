package common

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/docker/docker-credential-helpers/credentials"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	// Create mock database connection
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}

	// Create GORM database instance with mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("failed to create GORM instance: %v", err)
	}

	cleanup := func() {
		mockDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestDatabase_Add(t *testing.T) {
	tests := []struct {
		name        string
		creds       *credentials.Credentials
		expectError bool
		setupMock   func(sqlmock.Sqlmock)
	}{
		{
			name: "successful add new credential",
			creds: &credentials.Credentials{
				ServerURL: "https://example.com///test-context",
				Username:  "testuser",
				Secret:    "testsecret",
			},
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock checking for existing credential
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE server_url = \\? ORDER BY `gptscript_credentials`.`id` LIMIT \\?").
					WithArgs("https://example.com///test-context", 1).
					WillReturnError(gorm.ErrRecordNotFound)

				// Mock creating new credential
				mock.ExpectExec("INSERT INTO `gptscript_credentials` \\(`created_at`,`server_url`,`username`,`secret`,`context`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)").
					WithArgs(sqlmock.AnyArg(), "https://example.com///test-context", "testuser", "testsecret", "test-context").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "successful add replacing existing credential",
			creds: &credentials.Credentials{
				ServerURL: "https://example.com///test-context",
				Username:  "testuser",
				Secret:    "testsecret",
			},
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock finding existing credential
				rows := sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}).
					AddRow(1, time.Now(), "https://example.com///test-context", "olduser", "oldsecret", "test-context")
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE server_url = \\? ORDER BY `gptscript_credentials`.`id` LIMIT \\?").
					WithArgs("https://example.com///test-context", 1).
					WillReturnRows(rows)

				// Mock deleting existing credential
				mock.ExpectExec("DELETE FROM `gptscript_credentials` WHERE `gptscript_credentials`.`id` = \\?").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock creating new credential
				mock.ExpectExec("INSERT INTO `gptscript_credentials` \\(`created_at`,`server_url`,`username`,`secret`,`context`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)").
					WithArgs(sqlmock.AnyArg(), "https://example.com///test-context", "testuser", "testsecret", "test-context").
					WillReturnResult(sqlmock.NewResult(2, 1))
			},
		},
		{
			name: "invalid server URL format",
			creds: &credentials.Credentials{
				ServerURL: "invalid-url",
				Username:  "testuser",
				Secret:    "testsecret",
			},
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// No mock expectations since it should fail before DB operations
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			db := Database{
				db: gormDB,
			}

			err := db.Add(tt.creds)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled mock expectations: %v", err)
			}
		})
	}
}

func TestDatabase_Get(t *testing.T) {
	tests := []struct {
		name           string
		serverURL      string
		expectedUser   string
		expectedSecret string
		expectError    bool
		setupMock      func(sqlmock.Sqlmock)
	}{
		{
			name:           "successful get existing credential",
			serverURL:      "https://example.com///test-context",
			expectedUser:   "testuser",
			expectedSecret: "testsecret",
			expectError:    false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}).
					AddRow(1, time.Now(), "https://example.com///test-context", "testuser", "testsecret", "test-context")
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE server_url = \\? ORDER BY `gptscript_credentials`.`id` LIMIT \\?").
					WithArgs("https://example.com///test-context", 1).
					WillReturnRows(rows)
			},
		},
		{
			name:           "get non-existent credential",
			serverURL:      "https://nonexistent.com///test-context",
			expectedUser:   "",
			expectedSecret: "",
			expectError:    false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE server_url = \\? ORDER BY `gptscript_credentials`.`id` LIMIT \\?").
					WithArgs("https://nonexistent.com///test-context", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			db := Database{
				db: gormDB,
			}

			username, secret, err := db.Get(tt.serverURL)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if username != tt.expectedUser {
				t.Errorf("expected username %s, got %s", tt.expectedUser, username)
			}
			if secret != tt.expectedSecret {
				t.Errorf("expected secret %s, got %s", tt.expectedSecret, secret)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled mock expectations: %v", err)
			}
		})
	}
}

func TestDatabase_Delete(t *testing.T) {
	tests := []struct {
		name        string
		serverURL   string
		expectError bool
		setupMock   func(sqlmock.Sqlmock)
	}{
		{
			name:        "successful delete",
			serverURL:   "https://example.com///test-context",
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `gptscript_credentials` WHERE server_url = \\?").
					WithArgs("https://example.com///test-context").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:        "delete non-existent credential",
			serverURL:   "https://nonexistent.com///test-context",
			expectError: false, // Delete operations typically don't error on non-existent records
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `gptscript_credentials` WHERE server_url = \\?").
					WithArgs("https://nonexistent.com///test-context").
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			db := Database{
				db: gormDB,
			}

			err := db.Delete(tt.serverURL)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled mock expectations: %v", err)
			}
		})
	}
}

func TestDatabase_List(t *testing.T) {
	tests := []struct {
		name          string
		expectedCreds map[string]string
		expectError   bool
		setupMock     func(sqlmock.Sqlmock)
	}{
		{
			name: "successful list with multiple credentials",
			expectedCreds: map[string]string{
				"https://example1.com///context1": "user1",
				"https://example2.com///context2": "user2",
			},
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}).
					AddRow(1, time.Now(), "https://example1.com///context1", "user1", "secret1", "context1").
					AddRow(2, time.Now(), "https://example2.com///context2", "user2", "secret2", "context2")
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials`").
					WillReturnRows(rows)
			},
		},
		{
			name:          "successful list with no credentials",
			expectedCreds: make(map[string]string),
			expectError:   false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials`").
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			db := Database{
				db: gormDB,
			}

			creds, err := db.List()

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(creds) != len(tt.expectedCreds) {
				t.Errorf("expected %d credentials, got %d", len(tt.expectedCreds), len(creds))
			}

			for serverURL, expectedUser := range tt.expectedCreds {
				if actualUser, exists := creds[serverURL]; !exists {
					t.Errorf("expected credential for %s not found", serverURL)
				} else if actualUser != expectedUser {
					t.Errorf("expected user %s for %s, got %s", expectedUser, serverURL, actualUser)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled mock expectations: %v", err)
			}
		})
	}
}

func TestDatabase_ListWithContexts(t *testing.T) {
	tests := []struct {
		name          string
		contexts      []string
		expectedCreds map[string]string
		expectError   bool
		setupMock     func(sqlmock.Sqlmock)
	}{
		{
			name:     "successful list with specific contexts",
			contexts: []string{"context1", "context2"},
			expectedCreds: map[string]string{
				"https://example1.com///context1": "user1",
				"https://example2.com///context2": "user2",
			},
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// First query for context1
				rows1 := sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}).
					AddRow(1, time.Now(), "https://example1.com///context1", "user1", "secret1", "context1")
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE context = \\?").
					WithArgs("context1").
					WillReturnRows(rows1)

				// Second query for context2
				rows2 := sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}).
					AddRow(2, time.Now(), "https://example2.com///context2", "user2", "secret2", "context2")
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE context = \\?").
					WithArgs("context2").
					WillReturnRows(rows2)
			},
		},
		{
			name:          "list with no matching contexts",
			contexts:      []string{"nonexistent"},
			expectedCreds: make(map[string]string),
			expectError:   false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gptscript_credentials` WHERE context = \\?").
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "server_url", "username", "secret", "context"}))
			},
		},
		{
			name:        "invalid context with triple slash",
			contexts:    []string{"invalid///context"},
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// No mock expectations since it should fail before DB operations
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			db := Database{
				db: gormDB,
			}

			creds, err := db.ListWithContexts(tt.contexts)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(creds) != len(tt.expectedCreds) {
				t.Errorf("expected %d credentials, got %d", len(tt.expectedCreds), len(creds))
			}

			for serverURL, expectedUser := range tt.expectedCreds {
				if actualUser, exists := creds[serverURL]; !exists {
					t.Errorf("expected credential for %s not found", serverURL)
				} else if actualUser != expectedUser {
					t.Errorf("expected user %s for %s, got %s", expectedUser, serverURL, actualUser)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled mock expectations: %v", err)
			}
		})
	}
}
