package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	mclickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

var ErrDirtyMigrations = errors.New("dirty migrations")

type Clickhouse struct {
	conn  *sql.DB
	cfg   Config
	dbURI string
}

type Config struct {
	BindAddr string
	Database string
	Username string
	Password string
}

func New(ctx context.Context, cfg Config) (*Clickhouse, error) {
	urlScheme := url.URL{
		Scheme: "clickhouse",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   cfg.BindAddr,
		Path:   cfg.Database,
	}

	store := Clickhouse{
		cfg:   cfg,
		dbURI: urlScheme.String(),
	}

	zap.L().Debug(store.dbURI)

	if err := store.connect(); err != nil {
		return nil, fmt.Errorf("store.connect(): %w", err)
	}

	go func() {
		defer func() {
			err := store.conn.Close()
			if err != nil {
				zap.L().With(zap.Error(err)).Warn("store.conn.Close()")
			}
		}()

		<-ctx.Done()
	}()

	return &store, nil
}

func (c *Clickhouse) connect() error {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{c.cfg.BindAddr},
		Auth: clickhouse.Auth{
			Database: c.cfg.Database,
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		},
		// Debug: true,
	})

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("conn.Ping(): %w", err)
	}

	c.conn = conn

	zap.L().Debug("successful clickhouse connection")

	return nil
}

func getActualMigrationVersion(migrator *migrate.Migrate) (uint, error) {
	version, dirty, err := migrator.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, fmt.Errorf("migrator.Version(): %w", err)
	}

	if dirty {
		return version, ErrDirtyMigrations
	}

	return version, nil
}

//go:embed migrations
var fs embed.FS

func (c *Clickhouse) Migrate() error {
	sourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("iofs.New(fs, \"migrations\"): %w", err)
	}

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{c.cfg.BindAddr},
		Auth: clickhouse.Auth{
			Database: c.cfg.Database,
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		},
		// Debug: true,
	})

	dbDriver, err := mclickhouse.WithInstance(conn, &mclickhouse.Config{
		DatabaseName: "default",
	})
	if err != nil {
		return fmt.Errorf("mclickhouse.WithInstance(conn, &mclickhouse.Config{...}): %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "default", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance(\"iofs\", sourceDriver, \"default\", dbDriver): %w", err)
	}

	defer func() {
		errSource, errDB := migrator.Close()
		if errSource != nil {
			zap.L().With(zap.Error(errSource)).Warn("migrator.Close()")
		}

		if errDB != nil {
			zap.L().With(zap.Error(errDB)).Warn("migrator.Close()")
		}
	}()

	version, err := getActualMigrationVersion(migrator)
	if err != nil {
		return fmt.Errorf("getActualMigrationVersion(...): %w", err)
	}

	if err = migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			zap.L().Info(fmt.Sprintf("no migrations are required. Current version %d is up to date", version))

			return nil
		}

		return fmt.Errorf("failed to applying up migrations to the database: %w", err)
	}

	newVersion, err := getActualMigrationVersion(migrator)
	if err != nil {
		return fmt.Errorf("getActualMigrationVersion(...): %w", err)
	}

	zap.L().Info(fmt.Sprintf("migration from %d sourceDriver to %d sourceDriver succeeded", version, newVersion))

	return nil
}
