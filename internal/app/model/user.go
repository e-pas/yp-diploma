package model

import "time"

type User struct {
	ID           string
	Name         string
	HashedPasswd string
}

type SessKey struct {
	ID      string
	User_ID string
	Expires time.Time
}
