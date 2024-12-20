// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: pg_query.sql

package orm

import (
	"context"
)

type AddAdjFactorsParams struct {
	Sid     int32   `json:"sid"`
	SubID   int32   `json:"sub_id"`
	StartMs int64   `json:"start_ms"`
	Factor  float64 `json:"factor"`
}

type AddCalendarsParams struct {
	Name    string `json:"name"`
	StartMs int64  `json:"start_ms"`
	StopMs  int64  `json:"stop_ms"`
}

const addInsKline = `-- name: AddInsKline :one
insert into ins_kline ("sid", "timeframe", "start_ms", "stop_ms")
values ($1, $2, $3, $4) RETURNING id
`

type AddInsKlineParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
	StartMs   int64  `json:"start_ms"`
	StopMs    int64  `json:"stop_ms"`
}

func (q *Queries) AddInsKline(ctx context.Context, arg AddInsKlineParams) (int32, error) {
	row := q.db.QueryRow(ctx, addInsKline,
		arg.Sid,
		arg.Timeframe,
		arg.StartMs,
		arg.StopMs,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

type AddKHolesParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
	Start     int64  `json:"start"`
	Stop      int64  `json:"stop"`
}

const addKInfo = `-- name: AddKInfo :one
insert into kinfo
("sid", "timeframe", "start", "stop")
values ($1, $2, $3, $4)
    returning sid, timeframe, start, stop
`

type AddKInfoParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
	Start     int64  `json:"start"`
	Stop      int64  `json:"stop"`
}

func (q *Queries) AddKInfo(ctx context.Context, arg AddKInfoParams) (*KInfo, error) {
	row := q.db.QueryRow(ctx, addKInfo,
		arg.Sid,
		arg.Timeframe,
		arg.Start,
		arg.Stop,
	)
	var i KInfo
	err := row.Scan(
		&i.Sid,
		&i.Timeframe,
		&i.Start,
		&i.Stop,
	)
	return &i, err
}

type AddSymbolsParams struct {
	Exchange string `json:"exchange"`
	ExgReal  string `json:"exg_real"`
	Market   string `json:"market"`
	Symbol   string `json:"symbol"`
}

const delAdjFactors = `-- name: DelAdjFactors :exec
delete from adj_factors
where sid=$1
`

func (q *Queries) DelAdjFactors(ctx context.Context, sid int32) error {
	_, err := q.db.Exec(ctx, delAdjFactors, sid)
	return err
}

const delInsKline = `-- name: DelInsKline :exec
delete from ins_kline
where id=$1
`

func (q *Queries) DelInsKline(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, delInsKline, id)
	return err
}

const delKHoleRange = `-- name: DelKHoleRange :exec
delete from khole
where sid = $1 and timeframe=$2 and start >= $3 and stop <= $4
`

type DelKHoleRangeParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
	Start     int64  `json:"start"`
	Stop      int64  `json:"stop"`
}

func (q *Queries) DelKHoleRange(ctx context.Context, arg DelKHoleRangeParams) error {
	_, err := q.db.Exec(ctx, delKHoleRange,
		arg.Sid,
		arg.Timeframe,
		arg.Start,
		arg.Stop,
	)
	return err
}

const getAdjFactors = `-- name: GetAdjFactors :many
select id, sid, sub_id, start_ms, factor from adj_factors
where sid=$1
order by start_ms
`

func (q *Queries) GetAdjFactors(ctx context.Context, sid int32) ([]*AdjFactor, error) {
	rows, err := q.db.Query(ctx, getAdjFactors, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*AdjFactor{}
	for rows.Next() {
		var i AdjFactor
		if err := rows.Scan(
			&i.ID,
			&i.Sid,
			&i.SubID,
			&i.StartMs,
			&i.Factor,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllInsKlines = `-- name: GetAllInsKlines :many
select id, sid, timeframe, start_ms, stop_ms from ins_kline
`

func (q *Queries) GetAllInsKlines(ctx context.Context) ([]*InsKline, error) {
	rows, err := q.db.Query(ctx, getAllInsKlines)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*InsKline{}
	for rows.Next() {
		var i InsKline
		if err := rows.Scan(
			&i.ID,
			&i.Sid,
			&i.Timeframe,
			&i.StartMs,
			&i.StopMs,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInsKline = `-- name: GetInsKline :one
select id, sid, timeframe, start_ms, stop_ms from ins_kline
where sid=$1
`

func (q *Queries) GetInsKline(ctx context.Context, sid int32) (*InsKline, error) {
	row := q.db.QueryRow(ctx, getInsKline, sid)
	var i InsKline
	err := row.Scan(
		&i.ID,
		&i.Sid,
		&i.Timeframe,
		&i.StartMs,
		&i.StopMs,
	)
	return &i, err
}

const getKHoles = `-- name: GetKHoles :many
select id, sid, timeframe, start, stop from khole
where sid = $1 and timeframe = $2
`

type GetKHolesParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
}

func (q *Queries) GetKHoles(ctx context.Context, arg GetKHolesParams) ([]*KHole, error) {
	rows, err := q.db.Query(ctx, getKHoles, arg.Sid, arg.Timeframe)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*KHole{}
	for rows.Next() {
		var i KHole
		if err := rows.Scan(
			&i.ID,
			&i.Sid,
			&i.Timeframe,
			&i.Start,
			&i.Stop,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listExchanges = `-- name: ListExchanges :many
select distinct exchange from exsymbol
`

func (q *Queries) ListExchanges(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, listExchanges)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var exchange string
		if err := rows.Scan(&exchange); err != nil {
			return nil, err
		}
		items = append(items, exchange)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listKHoles = `-- name: ListKHoles :many
select id, sid, timeframe, start, stop from khole
WHERE sid = ANY($1::int[])
`

func (q *Queries) ListKHoles(ctx context.Context, dollar_1 []int32) ([]*KHole, error) {
	rows, err := q.db.Query(ctx, listKHoles, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*KHole{}
	for rows.Next() {
		var i KHole
		if err := rows.Scan(
			&i.ID,
			&i.Sid,
			&i.Timeframe,
			&i.Start,
			&i.Stop,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listKInfos = `-- name: ListKInfos :many
select sid, timeframe, start, stop from kinfo
`

func (q *Queries) ListKInfos(ctx context.Context) ([]*KInfo, error) {
	rows, err := q.db.Query(ctx, listKInfos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*KInfo{}
	for rows.Next() {
		var i KInfo
		if err := rows.Scan(
			&i.Sid,
			&i.Timeframe,
			&i.Start,
			&i.Stop,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSymbols = `-- name: ListSymbols :many
select id, exchange, exg_real, market, symbol, combined, list_ms, delist_ms from exsymbol
where exchange = $1
order by id
`

func (q *Queries) ListSymbols(ctx context.Context, exchange string) ([]*ExSymbol, error) {
	rows, err := q.db.Query(ctx, listSymbols, exchange)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*ExSymbol{}
	for rows.Next() {
		var i ExSymbol
		if err := rows.Scan(
			&i.ID,
			&i.Exchange,
			&i.ExgReal,
			&i.Market,
			&i.Symbol,
			&i.Combined,
			&i.ListMs,
			&i.DelistMs,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setKHole = `-- name: SetKHole :exec
update khole set start = $2, stop = $3
where id = $1
`

type SetKHoleParams struct {
	ID    int64 `json:"id"`
	Start int64 `json:"start"`
	Stop  int64 `json:"stop"`
}

func (q *Queries) SetKHole(ctx context.Context, arg SetKHoleParams) error {
	_, err := q.db.Exec(ctx, setKHole, arg.ID, arg.Start, arg.Stop)
	return err
}

const setKInfo = `-- name: SetKInfo :exec
update kinfo set start = $3, stop = $4
where sid = $1 and timeframe = $2
`

type SetKInfoParams struct {
	Sid       int32  `json:"sid"`
	Timeframe string `json:"timeframe"`
	Start     int64  `json:"start"`
	Stop      int64  `json:"stop"`
}

func (q *Queries) SetKInfo(ctx context.Context, arg SetKInfoParams) error {
	_, err := q.db.Exec(ctx, setKInfo,
		arg.Sid,
		arg.Timeframe,
		arg.Start,
		arg.Stop,
	)
	return err
}

const setListMS = `-- name: SetListMS :exec
update exsymbol set list_ms = $2, delist_ms = $3
where id = $1
`

type SetListMSParams struct {
	ID       int32 `json:"id"`
	ListMs   int64 `json:"list_ms"`
	DelistMs int64 `json:"delist_ms"`
}

func (q *Queries) SetListMS(ctx context.Context, arg SetListMSParams) error {
	_, err := q.db.Exec(ctx, setListMS, arg.ID, arg.ListMs, arg.DelistMs)
	return err
}
