
package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DBs struct {
	ChannelDB   *sql.DB
	AliasDB     *sql.DB
	GroupDB     *sql.DB
	FileDB      *sql.DB
	InvoiceDB   *sql.DB
	ReceptionDB *sql.DB
	SessionDB   *sql.DB
	UserDB      *sql.DB
}

var DB *DBs

func InitPostgres(cfg *Config) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	connect := func(name, dsn string) *sql.DB {
		if dsn == "" {
			log.Fatalf("PostgreSQL %s connection string is empty", name)
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("PostgreSQL %s open error: %v", name, err)
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(30 * time.Minute)
		db.SetConnMaxIdleTime(5 * time.Minute)

		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			log.Fatalf("PostgreSQL %s connection error: %v", name, err)
		}

		log.Printf("PostgreSQL %s connected successfully", name)
		return db
	}

	DB = &DBs{
		ChannelDB:   connect("channel", cfg.PgChannel),
		AliasDB:     connect("alias", cfg.PgAlias),
		GroupDB:     connect("group", cfg.PgGroup),
		FileDB:      connect("file", cfg.PgFile),
		InvoiceDB:   connect("invoice", cfg.PgInvoice),
		ReceptionDB: connect("reception", cfg.PgReception),
		SessionDB:   connect("session", cfg.PgSession),
		UserDB:      connect("user", cfg.PgUser),
	}
}

// ClosePostgres closes all PostgreSQL connection pools.
func ClosePostgres() error {
	if DB == nil {
		return nil
	}

	var firstErr error

	connections := []struct {
		name string
		db   *sql.DB
	}{
		{"channel", DB.ChannelDB},
		{"alias", DB.AliasDB},
		{"group", DB.GroupDB},
		{"file", DB.FileDB},
		{"invoice", DB.InvoiceDB},
		{"reception", DB.ReceptionDB},
		{"session", DB.SessionDB},
		{"user", DB.UserDB},
	}

	for _, conn := range connections {
		if conn.db == nil {
			continue
		}

		if err := conn.db.Close(); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf(
					"close PostgreSQL %s connection: %w",
					conn.name,
					err,
				)
			}
		}
	}

	return firstErr
}
