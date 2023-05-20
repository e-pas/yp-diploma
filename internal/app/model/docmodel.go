package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"yp-diploma/internal/app/config"
)

type docTime time.Time
type points uint

func (dt docTime) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("\"%s\"", time.Time(dt).Format("2006-01-02T15:04:05-07:00"))
	return []byte(res), nil
}

func (p points) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("%d.%d", p/100, p%100)
	if p%100 == 0 {
		res = fmt.Sprintf("%d", p/100)
	}
	return []byte(res), nil
}

func (p *points) UnmarshalJSON(data []byte) error {
	parts := strings.Split(string(data), ".")
	if len(parts) > 2 {
		return config.ErrInvalidData
	}
	if len(parts) == 1 {
		parts = append(parts, "0")
	}
	res1, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	res2, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	if res1 < 0 || res2 < 0 {
		return config.ErrInvalidData
	}
	*p = points(res1*100 + res2)
	return nil
}

type orderDoc struct {
	ID      string     `json:"number"`
	Status  StatusName `json:"status"`
	Accrual points     `json:"accrual,omitempty"`
	GenTime docTime    `json:"uploaded_at"`
}

type withdrawDoc struct {
	OrderID  string  `json:"order"`
	Withdraw points  `json:"sum"`
	GenTime  docTime `json:"processed_at"`
}

type balanceDoc struct {
	Balance  points `json:"current"`
	Accrual  points `json:"-"`
	Withdraw points `json:"withdrawn,omitempty"`
}

type withdrawReq struct {
	OrderID  string `json:"order"`
	Withdraw points `json:"sum"`
}

type accrualResp struct {
	OrderID string `json:"order"`
	Status  string `json:"status"`
	Accrual points `json:"accrual"`
}

func MarshalUserOrdersDoc(orders []Order) []byte {
	if len(orders) == 0 {
		return []byte{}
	}
	docs := make([]orderDoc, len(orders))
	for i := range orders {
		docs[i].ID = orders[i].ID
		docs[i].GenTime = docTime(orders[i].GenTime)
		docs[i].Accrual = points(orders[i].Accrual)
		docs[i].Status = Statuses[orders[i].Status]
	}
	buf, _ := json.MarshalIndent(docs, "", " ")
	return buf
}

func MarshalUserWithdrawsDoc(withdraws []Withdraw) []byte {
	if len(withdraws) == 0 {
		return []byte{}
	}
	docs := make([]withdrawDoc, len(withdraws))
	for i := range withdraws {
		docs[i].OrderID = withdraws[i].OrderID
		docs[i].GenTime = docTime(withdraws[i].GenTime)
		docs[i].Withdraw = points(withdraws[i].Withdraw)
	}
	buf, _ := json.MarshalIndent(docs, "", " ")
	return buf
}

func UnmarshalWithdrawRequest(buf []byte) (Withdraw, error) {
	req := withdrawReq{}
	err := json.Unmarshal(buf, &req)
	if err != nil {
		return Withdraw{}, err
	}
	return Withdraw{
		OrderID:  req.OrderID,
		Withdraw: int(req.Withdraw),
	}, nil
}

func MarshalUserBalanceDoc(balance Balance) []byte {
	doc := balanceDoc{
		Accrual:  points(balance.Accrual),
		Withdraw: points(balance.Withdraw),
		Balance:  points(balance.Balance),
	}
	buf, _ := json.MarshalIndent(doc, "", " ")
	return buf
}

func UnmarshalAcrrualResponse(buf []byte) (Accrual, error) {
	req := accrualResp{}
	err := json.Unmarshal(buf, &req)
	if err != nil {
		return Accrual{}, err
	}
	return Accrual{
		OrderID: req.OrderID,
		Status:  req.Status,
		Accrual: int(req.Accrual),
	}, nil
}
