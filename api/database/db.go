package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDB opens and verifies a MySQL connection using environment variables.
func ConnectDB() (*sql.DB, error) {
	user     := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	host     := os.Getenv("DATABASE_HOST")
	port     := os.Getenv("DATABASE_PORT")
	dbname   := os.Getenv("DATABASE_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Erro ao abrir conexão: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err = db.Ping(); err != nil {
		log.Printf("Erro ao conectar à base de dados: %v", err)
		return nil, err
	}

	log.Println("Conexão à base de dados estabelecida com sucesso.")
	return db, nil
}

// InitSchema creates all required tables if they do not exist.
func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id           INT AUTO_INCREMENT PRIMARY KEY,
		name         VARCHAR(255) NOT NULL,
		hostname     VARCHAR(255) NOT NULL,
		os           VARCHAR(100) NOT NULL DEFAULT '',
		version      VARCHAR(50)  NOT NULL DEFAULT '0.0.0',
		api_key      VARCHAR(64)  NOT NULL UNIQUE,
		status       ENUM('online','offline','unknown') NOT NULL DEFAULT 'unknown',
		last_heartbeat DATETIME    NULL,
		created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS updates (
		id           INT AUTO_INCREMENT PRIMARY KEY,
		version      VARCHAR(50)  NOT NULL,
		changelog    TEXT         NOT NULL,
		status       ENUM('pending','applied','failed') NOT NULL DEFAULT 'pending',
		scheduled_at DATETIME     NULL,
		created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS update_confirmations (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		update_id  INT NOT NULL,
		client_id  INT NOT NULL,
		status     ENUM('pending','success','failed') NOT NULL DEFAULT 'pending',
		message    TEXT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (update_id) REFERENCES updates(id) ON DELETE CASCADE,
		FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	for _, stmt := range splitStatements(schema) {
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("erro ao executar schema: %v — statement: %s", err, stmt)
		}
	}
	return nil
}

// splitStatements splits a multi-statement SQL string by semicolon.
func splitStatements(sql string) []string {
	var stmts []string
	current := ""
	for _, r := range sql {
		current += string(r)
		if r == ';' {
			stmts = append(stmts, current)
			current = ""
		}
	}
	return stmts
}

// QueryRows executes a SELECT query and returns results as JSON bytes.
func QueryRows(db *sql.DB, query string, args ...interface{}) ([]byte, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values    := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println("Erro ao ler linha:", err)
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return json.Marshal(results)
}

// QueryRow executes a SELECT and returns the first row as a map.
func QueryRow(db *sql.DB, query string, args ...interface{}) (map[string]interface{}, error) {
	data, err := QueryRows(db, query, args...)
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Exec executes a non-SELECT statement and returns the last insert ID.
func Exec(db *sql.DB, query string, args ...interface{}) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ExecAffected executes a non-SELECT statement and returns rows affected.
func ExecAffected(db *sql.DB, query string, args ...interface{}) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
