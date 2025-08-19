package pgmigrations

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Migration struct {
	Name string
	path string
}

var migrationsList = []Migration{}
var log *slog.Logger

func Init(location string, loggger *slog.Logger) {
	log = loggger
	files, err := os.ReadDir(location)
	if err != nil {
		log.Error("Unable to read migrations directory", "error", err)
		panic(err)
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			log.Info("Found migration", "name", strings.TrimSuffix(file.Name(), ".sql"))
			migrationsList = append(migrationsList, Migration{Name: strings.TrimSuffix(file.Name(), ".sql"), path: filepath.Join(location, file.Name())})
		}
	}

}
func RunMigrations(conn *pgx.Conn) {
	migrationTableExists := false
	conn.QueryRow(context.Background(), `SELECT EXISTS (
   SELECT FROM information_schema.tables 
   WHERE  table_schema = 'public'
   AND    table_name   = '_migrations'
   );`).Scan(&migrationTableExists)
	if !migrationTableExists {
		log.Info("Creating _migrations table")
		_, err := conn.Exec(context.Background(), `CREATE TABLE _migrations (
			name TEXT PRIMARY KEY,	
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP		
		);`)
		if err != nil {
			log.Error("Error creating _migrations table", "error", err)
			panic(err)
		}
	}

	for _, v := range migrationsList {
		var exists bool
		log.Info("Checking migration", "name", v.Name)
		conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT 1 FROM _migrations WHERE name = $1);`, v.Name).Scan(&exists)
		if !exists {
			log.Info("Running migration", "name", v.Name)
			tx, err := conn.Begin(context.Background())
			if err != nil {
				log.Error("Error starting transaction", "error", err)
				panic(err)
			}
			defer tx.Rollback(context.Background())
			_, err = tx.Exec(context.Background(), `INSERT INTO _migrations (name) VALUES ($1);`, v.Name)
			if err != nil {
				log.Error("Error inserting migration into _migrations", "error", err)
				panic(err)
			}
			sql, err := os.ReadFile(v.path)
			if err != nil {
				log.Error("Error reading migration file", "error", err)
				panic(err)
			}
			_, err = tx.Exec(context.Background(), string(sql))
			if err != nil {
				log.Error("Error running migration", "error", err)
				panic(err)
			}

			err = tx.Commit(context.Background())
			if err != nil {
				log.Error("Error committing transaction", "error", err)
				panic(err)
			}
		}
	}
	log.Info("Migrations complete")
	migrationsList = []Migration{}
}
