package migrator

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator interface {
	MigrateUp() error
	MigrateDown() error
}

type PgMigrator struct {
	//config        PgConfig
	DB            *sql.DB // Use connection pool
	MigrationsDir string
	//	driverURL     string // Store the constructed driver URL here

}

func NewMigrator(db *sql.DB, migrationsDir string) (*PgMigrator, error) {
	migrator := &PgMigrator{
		DB:            db,
		MigrationsDir: migrationsDir,
	}
	// Check if the database exists, if not, create it
	if err := migrator.createDatabaseIfNotExists(); err != nil {
		return nil, err
	}

	return migrator, nil

}

// ! The current implementation creates a new driverURL for each migration execution
// !(both MigrateUp and MigrateDown). While it still functions, it's slightly less efficient.
// driverURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
// 	m.config.User, m.config.Password, m.config.Host, m.config.Port, m.config.DbName)

func (m *PgMigrator) createDatabaseIfNotExists() error {

	rows, err := m.DB.Query("SELECT 1 FROM pg_database WHERE datname = 'authDB'")
	if err != nil {
		return err
	}
	defer rows.Close() // Close the rows after use

	if !rows.Next() { // Check if there are any rows
		// Database doesn't exist, create it
		_, err := m.DB.Exec("CREATE DATABASE authDB")
		if err != nil {
			return err
		}
		log.Println("Database created successfully!")
	}
	return nil
}

func (m *PgMigrator) MigrateUp() error {

	migrator, err := migrate.New(
		"file://"+m.MigrationsDir, "")
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
		log.Println("No migrations found to apply (Up)")
	} else {
		log.Printf("Up migrations applied successfully!")
	}
	return nil
}

func (m *PgMigrator) MigrateDown() error {

	migrator, err := migrate.New(
		"file://"+m.MigrationsDir, "")
	if err != nil {
		return err
	}

	if err := migrator.Down(); err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
		log.Println("No migrations found to apply (Down)")
	} else {
		log.Printf("Down migrations applied successfully!")
	}
	return nil
}
