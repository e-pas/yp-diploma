package model

import (
	"fmt"
	"time"
)

type Status int

const (
	Created Status = iota
	Processing
	Invalid
	Processed
)

type StatusName string

var Statuses = []StatusName{"NEW", "PROCESSING", "INVALID", "PROCESSED"}

type docTime time.Time

func (dt docTime) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("\"%s\"", time.Time(dt).Format("2006-01-02T15:04:05-07:00"))
	return []byte(res), nil
}

type Order struct {
	ID      string
	UserID  string
	GenTime time.Time
	Status  Status
	Accrual int
}

type Balance = struct {
	UserID   string
	Accrual  int
	Withdraw int
	Balance  int
}

type Withdraw = struct {
	OrderID  string
	GenTime  time.Time
	Withdraw int
}
