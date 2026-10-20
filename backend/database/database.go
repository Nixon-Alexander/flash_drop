package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migrateMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func InitDatabase() (*sql.DB, error) {
	connString := getConnectionString()

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func RunMigration(log *zap.Logger) {
	db, err := InitDatabase()
	if err != nil {
		log.Error("failed to initialize database",
			zap.Error(err),
		)
		return
	}
	defer db.Close()

	driver, err := migrateMysql.WithInstance(
		db,
		&migrateMysql.Config{},
	)
	if err != nil {
		log.Error("failed to create migration driver",
			zap.Error(err),
		)
		return
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./database/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Error("failed to initialize migration",
			zap.Error(err),
		)
		return
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("database migration is up to date")
			return
		}

		log.Error("failed to run database migration",
			zap.Error(err),
		)
		return
	}

	log.Info("database migration completed successfully")
}

func getConnectionString() string {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)
}
