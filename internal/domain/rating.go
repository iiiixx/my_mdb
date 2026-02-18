package domain

import "time"

type Rating struct {
	UserID  int       `json:"user_id"`
	MovieID int       `json:"movie_id"`
	Value   float32   `json:"rating"`
	TS      time.Time `json:"ts"`
}
