package entry

import (
	"flag"
	"fmt"
	"github.com/banbox/banbot/biz"
	"github.com/banbox/banbot/config"
	"github.com/banbox/banbot/data"
	"github.com/banbox/banbot/optmize"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"go.uber.org/zap"
	"os"
	"strings"
)

var subHelp = map[string]string{
	"trade":      "live trade",
	"backtest":   "backtest with strategies and data",
	"down_data":  "download kline data",
	"down_ws":    "download websocket data",
	"dbcmd":      "run db command",
	"spider":     "start the spider",
	"cmp_orders": "compare backTest orders with exchange orders",
}

const VERSION = "0.1.1"

type FuncEntry = func(args *config.CmdArgs) *errs.Error

func RunCmd() {
	if len(os.Args) < 2 {
		printAndExit()
	}
	cmdName := os.Args[1]
	if cmdName == "cmp_orders" {
		optmize.CompareExgBTOrders(os.Args[2:])
	} else {
		runMainEntrys(cmdName)
	}
}

func runMainEntrys(cmdName string) {
	var args config.CmdArgs

	var sub = flag.NewFlagSet(cmdName, flag.ExitOnError)
	var options []string
	var entry FuncEntry

	switch cmdName {
	case "trade":
		options = []string{"stake_amount", "pairs", "stg_dir", "with_spider", "task_hash", "task_id"}
		entry = RunTrade
	case "backtest":
		options = []string{"timerange", "stake_amount", "pairs", "stg_dir", "cpu_profile", "mem_profile"}
		entry = RunBackTest
	case "down_data":
		options = []string{"timerange", "pairs", "timeframes", "medium"}
		entry = RunDownData
	case "down_ws":
		break
	case "dbcmd":
		options = []string{"action", "tables", "force"}
		entry = RunDbCmd
	case "spider":
		entry = RunSpider
		break
	case "cvt_tick":
		options = []string{"in", "out", "cpu_profile", "mem_profile"}
		entry = data.RunFormatTick
	case "tick2kline":
		options = []string{"in", "out", "cpu_profile", "mem_profile"}
		entry = data.Build1mWithTicks
	case "load_kline":
		options = []string{"in", "cpu_profile", "mem_profile"}
		entry = LoadKLinesToDB
	case "load_cal":
		options = []string{"in"}
		entry = biz.LoadCalendars
	case "export_kline":
		options = []string{"out", "pairs", "timeframes", "adj", "tz"}
		entry = biz.ExportKlines
	default:
		printAndExit()
	}
	bindSubFlags(&args, sub, options...)

	err := sub.Parse(os.Args[2:])
	if err != nil {
		log.Error("fail", zap.Error(err))
		printAndExit()
	}
	args.Init()
	err2 := entry(&args)
	if err2 != nil {
		panic(err2)
	}
}

func bindSubFlags(args *config.CmdArgs, cmd *flag.FlagSet, opts ...string) {
	cmd.Var(&args.Configs, "config", "config path to use, Multiple -config options may be used")
	cmd.StringVar(&args.Logfile, "logfile", "", "Log to the file specified")
	cmd.StringVar(&args.DataDir, "datadir", "", "Path to data dir.")
	cmd.BoolVar(&args.NoDb, "nodb", false, "dont save orders to database")
	cmd.BoolVar(&args.Debug, "debug", false, "set logging level to debug")
	cmd.BoolVar(&args.NoCompress, "no-compress", false, "disable compress for hyper table")
	cmd.BoolVar(&args.NoDefault, "no-default", false, "ignore default: config.yml, config.local.yml")
	cmd.IntVar(&args.MaxPoolSize, "max-pool-size", 0, "max pool size for db")

	for _, key := range opts {
		switch key {
		case "stake_amount":
			cmd.Float64Var(&args.StakeAmount, "stake-amount", 0.0, "Override `stake_amount` in config")
		case "stake_pct":
			cmd.Float64Var(&args.StakePct, "stake-pct", 0.0, "Override `stake_pct` in config")
		case "pairs":
			cmd.StringVar(&args.RawPairs, "pairs", "", "comma-separated pairs")
		case "stg_dir":
			cmd.Var(&args.StrategyDirs, "stg-dir", "dir path for strategies")
		case "with_spider":
			cmd.BoolVar(&args.WithSpider, "spider", false, "start spider if not running")
		case "timerange":
			cmd.StringVar(&args.TimeRange, "timerange", "", "Specify what timerange of data to use")
		case "timeframes":
			cmd.StringVar(&args.RawTimeFrames, "timeframes", "", "comma-seperated timeframes to use")
		case "medium":
			cmd.StringVar(&args.Medium, "medium", "", "data medium:db,file")
		case "action":
			cmd.StringVar(&args.Action, "action", "", "db action name")
		case "tables":
			cmd.StringVar(&args.RawTables, "tables", "", "db tables, comma-separated")
		case "force":
			cmd.BoolVar(&args.Force, "force", false, "skip confirm")
		case "task_hash":
			cmd.StringVar(&args.TaskHash, "task-hash", "", "hash code to use")
		case "cpu_profile":
			cmd.BoolVar(&args.CPUProfile, "cpu-profile", false, "enable cpu profile")
		case "mem_profile":
			cmd.BoolVar(&args.MemProfile, "mem-profile", false, "enable memory profile")
		case "task_id":
			cmd.IntVar(&args.TaskId, "task-id", 0, "task")
		case "in":
			cmd.StringVar(&args.InPath, "in", "", "input file or directory")
		case "out":
			cmd.StringVar(&args.OutPath, "out", "", "output file or directory")
		case "adj":
			cmd.StringVar(&args.AdjType, "adj", "", "qfq/hfq for kline")
		case "tz":
			cmd.StringVar(&args.TimeZone, "tz", "", "timeZone, default: utc")
		default:
			log.Warn(fmt.Sprintf("undefined argument: %s", key))
			os.Exit(1)
		}
	}
}

func printAndExit() {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("banbot %v\nplease run with a subcommand:\n", VERSION))
	for k, v := range subHelp {
		b.WriteString(fmt.Sprintf("  %s\n\t%v\n", k, v))
	}
	log.Warn(b.String())
	os.Exit(1)
}
