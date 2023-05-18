package model

import "encoding/json"

type orderDoc struct {
	ID      string     `json:"number"`
	Status  StatusName `json:"status"`
	Accrual int        `json:"accrual,omitempty"`
	GenTime docTime    `json:"uploaded_at"`
}

type withdrawDoc struct {
	OrderID  string  `json:"order"`
	Withdraw int     `json:"sum"`
	GenTime  docTime `json:"processed_at"`
}

type balanceDoc struct {
	Balance  int `json:"current"`
	Accrual  int `json:"-"`
	Withdraw int `json:"withdraw,omitempty"`
}

type WithdrawReq struct {
	OrderID  string `json:"order"`
	Withdraw int    `json:"sum"`
}

type AccrualResp struct {
	OrderID string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

func MarshalUserOrdersDoc(orders ...Order) []byte {
	if len(orders) == 0 {
		return []byte{}
	}
	docs := make([]orderDoc, len(orders))
	for i := range orders {
		docs[i].ID = orders[i].ID
		docs[i].GenTime = docTime(orders[i].GenTime)
		docs[i].Accrual = orders[i].Accrual
		docs[i].Status = Statuses[orders[i].Status]
	}
	buf, _ := json.MarshalIndent(docs, "", " ")
	return buf
}

func MarshalUserWithdrawsDoc(withdraws ...Withdraw) []byte {
	if len(withdraws) == 0 {
		return []byte{}
	}
	docs := make([]withdrawDoc, len(withdraws))
	for i := range withdraws {
		docs[i].OrderID = withdraws[i].OrderID
		docs[i].GenTime = docTime(withdraws[i].GenTime)
		docs[i].Withdraw = withdraws[i].Withdraw
	}
	buf, _ := json.MarshalIndent(docs, "", " ")
	return buf
}

func MarshalUserBalanceDoc(balance Balance) []byte {
	doc := balanceDoc{
		Accrual:  balance.Accrual,
		Withdraw: balance.Withdraw,
		Balance:  balance.Balance,
	}
	buf, _ := json.MarshalIndent(doc, "", " ")
	return buf
}

func UnmarshalAcrrualResponse(buf []byte) (AccrualResp, error) {
	res := AccrualResp{}
	err := json.Unmarshal(buf, &res)
	if err != nil {
		return AccrualResp{}, err
	}
	return res, nil
}
