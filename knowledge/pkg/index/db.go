package index

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/tools/knowledge/pkg/index/postgres"
	"github.com/obot-platform/tools/knowledge/pkg/index/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(ctx context.Context, dsn string, autoMigrate bool) (Index, error) {
	gormLogLevel := logger.Silent
	switch os.Getenv("GORM_LOG_LEVEL") {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	}

	var (
		indexDB Index
		err     error
		gormCfg = &gorm.Config{
			Logger: logger.New(log.Default(), logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				Colorful:      true,
				LogLevel:      gormLogLevel,
			}),
			TranslateError: true,
		}
	)

	dialect := strings.Split(dsn, "://")[0]

	slog.Debug("indexdb", "dialect", dialect, "dsn", dsn)

	switch dialect {
	case "sqlite":
		indexDB, err = sqlite.New(ctx, dsn, gormCfg, autoMigrate)
	case "postgres":
		indexDB, err = postgres.New(ctx, dsn, gormCfg, autoMigrate)
	default:
		err = fmt.Errorf("unsupported dialect: %q", dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open index DB: %w", err)
	}

	return indexDB, nil
}
