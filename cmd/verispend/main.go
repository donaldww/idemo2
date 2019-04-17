package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const port int = 26257 // Default port on first cockroach node

func main() {
	for i := 0; i < 3; i++ {
		currentPort := port + i
		loginString := fmt.Sprintf("postgres://root@:%d/defaultdb?sslmode=disable", currentPort)
		// Connect to the "bank" database.
		db, err := sql.Open("postgres", loginString)
		if err != nil {
			log.Fatal("error connecting to the database: ", err)
		}
		defer db.Close()

		// Print out the balances.
		rows, err := db.Query("SELECT id, balance FROM bank.accounts")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		fmt.Printf("Balance: node=%d port=%d", i+1, currentPort)
		fmt.Println()
		for rows.Next() {
			var id int
			var balance float32
			if err := rows.Scan(&id, &balance); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%d %.2f\n", id, balance)
		}
		fmt.Println()
	}
}
