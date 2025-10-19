package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

const (
	dbDriver = "postgres"
	dbSource = "postgres://user:password@localhost:5430/mydatabase?sslmode=disable"
)

func main() {

	db, err := sqlx.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")

	newUser := User{Name: "Diana", Email: "diana@example.com", Balance: 300}
	if err := InsertUser(db, newUser); err != nil {
		log.Println("Insert error:", err)
	}

	users, _ := GetAllUsers(db)
	fmt.Println("\n📋 All users:")
	for _, u := range users {
		fmt.Printf("%d | %s | %s | %.2f\n", u.ID, u.Name, u.Email, u.Balance)
	}

	fmt.Println("\n💸 Transfer 200 from Adilet to Ali...")
	if err := TransferBalance(db, 1, 2, 200); err != nil {
		log.Println("Transfer error:", err)
	} else {
		fmt.Println("✅ Transfer successful!")
	}

	users, _ = GetAllUsers(db)
	fmt.Println("\n📋 Updated users:")
	for _, u := range users {
		fmt.Printf("%d | %s | %.2f\n", u.ID, u.Name, u.Balance)
	}
}

func InsertUser(db *sqlx.DB, user User) error {
	query := `INSERT INTO users (name, email, balance) VALUES (:name, :email, :balance)`
	_, err := db.NamedExec(query, user)
	return err
}

func GetAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	err := db.Select(&users, "SELECT * FROM users ORDER BY id")
	return users, err
}

func GetUserByID(db *sqlx.DB, id int) (User, error) {
	var user User
	err := db.Get(&user, "SELECT * FROM users WHERE id=$1", id)
	return user, err
}

func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	var sender User
	if err := tx.Get(&sender, "SELECT * FROM users WHERE id=$1", fromID); err != nil {
		tx.Rollback()
		return fmt.Errorf("sender not found: %w", err)
	}

	if sender.Balance < amount {
		tx.Rollback()
		return fmt.Errorf("insufficient funds")
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, fromID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error decreasing sender balance: %w", err)
	}

	var receiver User
	if err := tx.Get(&receiver, "SELECT * FROM users WHERE id=$1", toID); err != nil {
		tx.Rollback()
		return fmt.Errorf("receiver not found: %w", err)
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, toID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error increasing receiver balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}
