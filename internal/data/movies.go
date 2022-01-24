package data

import "time"

type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"-"` // Use `-` tag to unshow this field to users.
	Title    string    `json:"title,omitempty"`
	Year     int32     `json:"year,omitempty"`
	Runtime  int32     `json:"runtime,omitempty"` // Movie runtime in minutes
	Genres   []string  `json:"genres"`
	Version  int32     `json:"version"`
}
