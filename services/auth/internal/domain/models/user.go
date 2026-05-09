package models

import "time"

type User struct {
	ID        int
	Email     string
	PassHash  []byte
	Nickname  string
	CreatedAt time.Time
}
