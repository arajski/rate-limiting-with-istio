package repository

import (
	"database/sql"
	"fmt"

	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/internal/domain"
)

type RateLimits struct {
	DB *sql.DB
}

func (r *RateLimits) GetAllRateLimits() ([]*domain.RateLimit, error) {
	limitRows, err := r.DB.Query("SELECT T1.id, T1.user_id, T2.name, T1.endpoint, T1.limit_unit, T1.limit_interval FROM rate_limits AS T1 INNER JOIN users AS T2 ON T1.user_id = T2.id")
	if err != nil {
		return nil, err
	}
	defer limitRows.Close()

	var limits []*domain.RateLimit
	for limitRows.Next() {
		var limit domain.RateLimit
		if err := limitRows.Scan(&limit.ID, &limit.UserID, &limit.UserName, &limit.Endpoint, &limit.LimitUnit, &limit.LimitInterval); err != nil {
			fmt.Printf("Error scanning rate limit row: %v", err)
			continue
		}
		limits = append(limits, &limit)
	}
	return limits, nil
}

func (r *RateLimits) GetRateLimitByID(id string) (*domain.RateLimit, error) {
	var limit domain.RateLimit
	err := r.DB.QueryRow("SELECT T1.id, T1.user_id, T2.name, T1.endpoint, T1.limit_unit, T1.limit_interval FROM rate_limits AS T1 INNER JOIN users AS T2 ON T1.user_id = T2.id WHERE T1.id = ?", id).Scan(&limit.ID, &limit.UserID, &limit.UserName, &limit.Endpoint, &limit.LimitUnit, &limit.LimitInterval)
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

func (r *RateLimits) CreateRateLimit(userID string, endpoint string, limitUnit string, limitInterval string) error {
	_, err := r.DB.Exec("INSERT INTO rate_limits(user_id, endpoint, limit_unit, limit_interval) VALUES(?, ?, ?, ?)", userID, endpoint, limitUnit, limitInterval)
	if err != nil {
		return err
	}
	return nil
}

func (r *RateLimits) UpdateRateLimit(id string, userID string, endpoint string, limitUnit string, limitInterval string) error {
	_, err := r.DB.Exec("UPDATE rate_limits SET user_id=?, endpoint=?, limit_unit=?, limit_interval=? WHERE id=?", userID, endpoint, limitUnit, limitInterval, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *RateLimits) DeleteRateLimit(id string) error {
	_, err := r.DB.Exec("DELETE FROM rate_limits WHERE id=?", id)
	if err != nil {
		return err
	}
	return nil
}
