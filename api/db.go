package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	// Buscar as variáveis de ambiente
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	dbname := os.Getenv("DATABASE_NAME") // Caso tenhas, senão define manualmente

	if dbname == "" {
		dbname = "nome_da_tua_base" // Define aqui diretamente se não estiver no .env
	}

	// Construir a DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)

	// Conectar
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Erro ao abrir conexão: %v", err)
		return nil, err
	}

	// Verificar se está vivo
	err = db.Ping()
	if err != nil {
		log.Printf("Erro ao conectar à base de dados: %v", err)
		return nil, err
	}

	fmt.Println("Conexão à base de dados estabelecida com sucesso.")
	return db, nil
}

func json(db *sql.DB, query string, args ...interface{}) ([]byte, error)
{
	rows, err := db.Query(query, args...)

	// executa a query e verifica se houve erro
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Verifica se há linhas retornadas
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Cria um slice para armazenar os resultados
	var results []map[string]interface{}


	// faz um loop por cada linha retornada
	for rows.Next(){
		// Slice de interfaces para os valores
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		// associar os ponteiros aos valores
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan da linha e popular os valores via ponteiros
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println("Erro ao escanear linha:", err)
			continue
		}

		// Mapear colunas para valores
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Lidar com []byte (strings vindas da DB)
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		results = append(results, row)
	}

	return json.Marshal(results)
}