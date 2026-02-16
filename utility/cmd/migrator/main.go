package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		host            string
		port            int
		user            string
		password        string
		dbname          string
		sslmode         string
		migrationsPath  string
		migrationsTable string
	)

	flag.StringVar(&host, "host", "localhost", "postgres host")
	flag.IntVar(&port, "port", 5432, "postgres port")
	flag.StringVar(&user, "user", "", "postgres user")
	flag.StringVar(&password, "password", "", "postgres password")
	flag.StringVar(&dbname, "dbname", "", "postgres db name")
	flag.StringVar(&sslmode, "sslmode", "disable", "ssl mode")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "migrations table")
	flag.Parse()

	if user == "" || password == "" || dbname == "" {
		panic("user, password and dbname are required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
		migrationsTable,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied")
}
