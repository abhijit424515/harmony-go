package db

import (
	"database/sql"
	"fmt"
	"harmony/backend/common"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func createTableIfNotExists(tableName string, schema string) error {
	// Check if table exists
	var count int
	err := common.Db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	if err != nil {
		return fmt.Errorf("[error] checking for table %s: %w", tableName, err)
	}

	if count > 0 {
		log.Printf("Table %s already exists, skipping creation.\n", tableName)
		return nil
	}

	// Create the table if it doesn't exist
	_, err = common.Db.Exec(schema)
	if err != nil {
		return fmt.Errorf("[error] failed to create table %s: %w", tableName, err)
	}

	log.Printf("Table %s created successfully.\n", tableName)
	return nil
}

func StartLightweightCleanupJob() {
	go func() {
		for {
			_, err := common.Db.Exec("DELETE FROM buffer WHERE ttl < unixepoch()")
			if err != nil {
				log.Printf("Error cleaning up expired buffers: %v", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()
}

func setupTables() error {
	userSchema := `
	CREATE TABLE user (
		_id TEXT PRIMARY KEY,
		email TEXT UNIQUE
	);
	CREATE INDEX IF NOT EXISTS email_index ON user(email);
	`

	bufferSchema := `
	CREATE TABLE buffer (
		_id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		time INTEGER NOT NULL,
		ttl INTEGER NOT NULL,
		type TEXT NOT NULL,
		data BLOB,
		FOREIGN KEY (user_id) REFERENCES user(_id)
	);
	CREATE INDEX IF NOT EXISTS userid_index ON buffer(user_id);
	`

	if err := createTableIfNotExists("user", userSchema); err != nil {
		return fmt.Errorf("[error] creating user table: %v", err)
	}

	if err := createTableIfNotExists("buffer", bufferSchema); err != nil {
		return fmt.Errorf("[error] creating buffer table: %v", err)
	}

	StartLightweightCleanupJob()
	return nil
}

func Setup() error {
	dbDir := "./data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "harmony.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("unable to open SQLite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	common.Db = db

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign key constraints: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	if err := setupTables(); err != nil {
		return err
	}

	log.Println("SQLite database setup completed successfully")
	return nil
}
