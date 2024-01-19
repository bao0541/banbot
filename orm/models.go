// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package orm

import ()

type BotTask struct {
	ID       int64
	Mode     string
	Name     string
	CreateAt int64
	StartAt  int64
	StopAt   int64
	Info     string
}

type ExOrder struct {
	ID        int64
	TaskID    int32
	InoutID   int32
	Symbol    string
	Enter     bool
	OrderType string
	OrderID   string
	Side      string
	CreateAt  int64
	Price     float64
	Average   float64
	Amount    float64
	Filled    float64
	Status    int16
	Fee       float64
	FeeType   string
	UpdateAt  int64
}

type ExSymbol struct {
	ID       int32
	Exchange string
	Market   string
	Symbol   string
	ListMs   int64
	DelistMs int64
}

type IOrder struct {
	ID         int64
	TaskID     int32
	Symbol     string
	Sid        int32
	Timeframe  string
	Short      bool
	Status     int16
	EnterTag   string
	InitPrice  float64
	QuoteCost  float64
	ExitTag    string
	Leverage   int32
	EnterAt    int64
	ExitAt     int64
	Strategy   string
	StgVer     int32
	ProfitRate float64
	Profit     float64
	Info       string
}

type KHole struct {
	ID        int64
	Sid       int32
	Timeframe string
	Start     int64
	Stop      int64
}

type KInfo struct {
	Sid       int32
	Timeframe string
	Start     int64
	Stop      int64
}

type KlineUn struct {
	Sid       int32
	StartMs   int64
	StopMs    int64
	Timeframe string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type Overlay struct {
	ID       int64
	User     int32
	Sid      int32
	StartMs  int64
	StopMs   int64
	TfMsecs  int32
	UpdateAt int64
	Data     string
}

type User struct {
	ID             int32
	UserName       string
	Avatar         string
	Mobile         string
	MobileVerified bool
	Email          string
	EmailVerified  bool
	PwdSalt        string
	LastIp         string
	CreateAt       int64
	LastLogin      int64
	VipType        int32
	VipExpireAt    int64
	InviterID      int32
}
