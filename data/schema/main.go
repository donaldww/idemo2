package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// port is the default first port on the cockroach cluster
const (
	port = 26257
)

const (
	_ = "CREATE TABLE IF NOT EXISTS BALANCE (id INT PRIMARY KEY, balance INT)"
)

func main() {
	// TODO: create a database library
	login := fmt.Sprintf("postgres://root@:%d/defaultdb?sslmode=disable", port)
	db, err := sql.Open("postgres", login)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT)"); err != nil {
		log.Fatal(err)
	}

	// Get the largest row id.
	rows, err := db.Query("SELECT max(id) FROM accounts")
	var maxID int
	for rows.Next() {
		if err = rows.Scan(&maxID); err != nil {
			log.Fatal(err)
		}
	}

	insertBalance(maxID+1, db)

	// Print out the balances.
	rows, err = db.Query("SELECT id, balance FROM accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Initial balances:")
	for rows.Next() {
		var id, balance int
		if err := rows.Scan(&id, &balance); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %d\n", id, balance)
	}
}

func insertBalance(id int, db *sql.DB) {
	statement := fmt.Sprintf("INSERT INTO accounts (id, balance) VALUES (%d, 1000), (%d, 250)", id,
		id+1)
	if _, err := db.Exec(statement); err != nil {
		log.Fatal(err)
	}
}
