package domain

type User struct {
	ID          int
	Name        string
	APIToken    string
	MaskedToken string
}

type RateLimit struct {
	ID            int
	UserID        int
	UserName      string
	Endpoint      string
	LimitUnit     int
	LimitInterval string
}
