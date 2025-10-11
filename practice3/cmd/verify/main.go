package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./expense.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Database connected successfully!")
}
