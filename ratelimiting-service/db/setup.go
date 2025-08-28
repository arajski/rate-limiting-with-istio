package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./demo.db")
	if err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	userTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"api_token" TEXT
	);`

	rateLimitTableSQL := `CREATE TABLE IF NOT EXISTS rate_limits (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"user_id" INTEGER,
		"endpoint" TEXT,
		"limit_unit" INTEGER,
		"limit_interval" TEXT
	);`

	_, err := DB.Exec(userTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(rateLimitTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func SeedDB() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("failed to check user count: %v", err)
	}

	if count > 0 {
		log.Println("database already seeded with users, skipping.")
		return
	}

	sampleUsers := []struct {
		Name  string
		Token string
	}{
		{"Acme Corp", "a1b2c3d4e5f6g7h2i9j0"},
		{"Globex Inc", "k1l2m3n4o5p6q7r3s9t0"},
		{"Soylent Corp", "u1v2w3x4y5z6a748c9d0"},
	}

	for _, user := range sampleUsers {
		_, err := DB.Exec("INSERT INTO users(name, api_token) VALUES(?, ?)", user.Name, user.Token)
		if err != nil {
			log.Printf("Error seeding user: %v", err)
		}
	}
}
