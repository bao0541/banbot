package orm

import (
	"context"
	"github.com/anyongjin/banbot/config"
	"github.com/anyongjin/banbot/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"os"
)

var (
	pool   *pgxpool.Pool
	TaskID int64
	Task   *BotTask
)

func Setup() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
	dbCfg := config.Database
	if dbCfg == nil {
		log.Panic("database config is missing!")
	}
	dbPool, err := pgxpool.New(context.Background(), dbCfg.Url)
	if err != nil {
		log.Error("Unable to create connection pool", zap.Error(err))
		os.Exit(1)
	}
	pool = dbPool
}
