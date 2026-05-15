package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "your_password" // замените на свой пароль
	dbname   = "archive_app"
)

func InitDB() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}

	log.Println("Connected to database")
	if err := createTables(); err != nil {
		return err
	}
	return nil
}

func createTables() error {
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
	if _, err := DB.Exec(createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	var columnExists bool
	checkColumn := `
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='users' AND column_name='username'
    )`
	if err := DB.QueryRow(checkColumn).Scan(&columnExists); err != nil {
		return fmt.Errorf("failed to check column: %v", err)
	}
	if !columnExists {
		log.Println("Adding username column...")
		if _, err := DB.Exec(`ALTER TABLE users ADD COLUMN username VARCHAR(50) UNIQUE`); err != nil {
			return err
		}
		DB.Exec(`UPDATE users SET username = 'temp_' || id WHERE username IS NULL`)
		DB.Exec(`ALTER TABLE users ALTER COLUMN username SET NOT NULL`)
	}

	checkPasswordColumn := `
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='users' AND column_name='password_hash'
    )`
	if err := DB.QueryRow(checkPasswordColumn).Scan(&columnExists); err != nil {
		return err
	}
	if !columnExists {
		if _, err := DB.Exec(`ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("failed to add password_hash: %v", err)
		}
	}

	createHistoryTable := `
    CREATE TABLE IF NOT EXISTS archive_history (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        archive_name VARCHAR(255) NOT NULL,
        action VARCHAR(20) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
	if _, err := DB.Exec(createHistoryTable); err != nil {
		return fmt.Errorf("failed to create archive_history: %v", err)
	}

	log.Println("Database schema ready")
	return nil
}

func RegisterUser(username, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, hashed)
	return err
}

func LoginUser(username, password string) (int, error) {
	var id int
	var hash string
	err := DB.QueryRow("SELECT id, password_hash FROM users WHERE username=$1", username).Scan(&id, &hash)
	if err != nil {
		return 0, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, err
	}
	return id, nil
}

func SaveArchiveHistory(userID int, archiveName, action string) error {
	_, err := DB.Exec("INSERT INTO archive_history (user_id, archive_name, action) VALUES ($1, $2, $3)",
		userID, archiveName, action)
	return err
}

// ArchiveRecord представляет запись истории
type ArchiveRecord struct {
	ArchiveName string
	Action      string
	CreatedAt   time.Time
}

// GetUserHistory возвращает историю операций пользователя
func GetUserHistory(userID int) ([]ArchiveRecord, error) {
	rows, err := DB.Query(`SELECT archive_name, action, created_at FROM archive_history WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []ArchiveRecord
	for rows.Next() {
		var rec ArchiveRecord
		if err := rows.Scan(&rec.ArchiveName, &rec.Action, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
