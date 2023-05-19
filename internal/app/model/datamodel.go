package model

import (
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

type Accrual struct {
	OrderID string
	Status  string
	Accrual int
}
