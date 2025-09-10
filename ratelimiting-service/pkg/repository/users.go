package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/internal/domain"
)

type Users struct {
	DB *sql.DB
}

func (u *Users) GetAllUsers() ([]domain.User, error) {
	rows, err := u.DB.Query("SELECT id, name, api_token FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Name, &user.APIToken); err != nil {
			fmt.Printf("Error scanning user row: %v", err)
			continue
		}
		user.MaskedToken = strings.Repeat("*", 3) + user.APIToken[len(user.APIToken)-5:]
		users = append(users, user)
	}
	return users, nil
}

func (u *Users) GetAPITokenByID(id int) (string, error) {
	var token string
	err := u.DB.QueryRow("SELECT api_token FROM users WHERE id = ?", id).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("could not find the api token")
		}
		return "", err
	}
	return token, nil
}
