package goods

import (
	"fmt"
	"github.com/banbox/banbot/btime"
	"github.com/banbox/banbot/config"
	"github.com/banbox/banbot/core"
	"github.com/banbox/banbot/exg"
	"github.com/banbox/banbot/orm"
	"github.com/banbox/banbot/utils"
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	utils2 "github.com/banbox/banexg/utils"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/floats"
	"math"
	"math/rand"
	"slices"
	"sort"
	"strings"
)

func (f *BaseFilter) IsNeedTickers() bool {
	return f.NeedTickers
}

func (f *BaseFilter) IsDisable() bool {
	return f.Disable
}

func (f *BaseFilter) GetName() string {
	return f.Name
}

func (f *AgeFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	if f.Min == 0 && f.Max == 0 {
		return symbols, nil
	}
	backNum := max(f.Max, f.Min) + 1
	return filterByOHLCV(symbols, "1d", backNum, 0, func(s string, klines []*banexg.Kline) bool {
		knum := len(klines)
		if knum == 0 {
			return f.AllowEmpty
		}
		return knum >= f.Max && (f.Max == 0 || knum <= f.Max)
	})
}

func (f *VolumePairFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	var symbolVols = make([]SymbolVol, 0)
	if !f.NeedTickers {
		backTf, backNum := f.BackTimeframe, f.BackPeriod
		tfSecs := utils2.TFToSecs(backTf)
		if tfSecs > 3600 {
			backTf = "1h"
			backNum *= tfSecs / 3600
		}
		var err *errs.Error
		symbolVols, err = getSymbolVols(symbols, backTf, backNum)
		if err != nil {
			return nil, err
		}
	} else {
		for _, symbol := range symbols {
			tik, ok := tickers[symbol]
			if !ok {
				continue
			}
			symbolVols = append(symbolVols, SymbolVol{symbol, tik.QuoteVolume, tik.Close})
		}
	}
	slices.SortFunc(symbolVols, func(a, b SymbolVol) int {
		return int((b.Vol - a.Vol) / 1000)
	})
	if !f.AllowEmpty && f.MinValue == 0 {
		f.MinValue = core.AmtDust
	}
	if f.MinValue > 0 {
		for i, v := range symbolVols {
			if v.Vol >= f.MinValue {
				continue
			}
			symbolVols = symbolVols[:i]
			break
		}
	}
	resPairs, _ := filterByMinCost(symbolVols)
	if f.LimitRate > 0 && f.LimitRate < 1 {
		num := int(math.Round(f.LimitRate * float64(len(resPairs))))
		resPairs = resPairs[:num]
	}
	if f.Limit > 0 && f.Limit < len(resPairs) {
		resPairs = resPairs[:f.Limit]
	}
	return resPairs, nil
}

type SymbolVol struct {
	Symbol string
	Vol    float64
	Price  float64
}

func getSymbolVols(symbols []string, tf string, num int) ([]SymbolVol, *errs.Error) {
	var symbolVols = make([]SymbolVol, 0)
	endMS := int64(0)
	if !core.LiveMode {
		endMS = btime.TimeMS()
	}
	callBack := func(symbol string, _ string, klines []*banexg.Kline, adjs []*orm.AdjInfo) {
		if len(klines) == 0 {
			symbolVols = append(symbolVols, SymbolVol{symbol, 0, 0})
		} else {
			total := float64(0)
			slices.Reverse(klines)
			if len(klines) > num {
				klines = klines[:num]
			}
			for _, k := range klines {
				total += k.Close * k.Volume
			}
			vol := total / float64(len(klines))
			price := klines[len(klines)-1].Close
			symbolVols = append(symbolVols, SymbolVol{symbol, vol, price})
		}
	}
	exchange := exg.Default
	err := orm.FastBulkOHLCV(exchange, symbols, tf, 0, endMS, num, callBack)
	if err != nil {
		return nil, err
	}
	if len(symbolVols) == 0 {
		msg := fmt.Sprintf("No data found for %d pairs at %v", len(symbols), endMS)
		return nil, errs.NewMsg(core.ErrRunTime, msg)
	}
	return symbolVols, nil
}

func filterByMinCost(symbols []SymbolVol) ([]string, map[string]float64) {
	res := make([]string, 0, len(symbols))
	skip := make(map[string]float64)
	exchange := exg.Default
	accCost := float64(0)
	for name := range config.Accounts {
		curCost := config.GetStakeAmount(name)
		if curCost > accCost {
			accCost = curCost
		}
	}
	for _, item := range symbols {
		mar, err := exchange.GetMarket(item.Symbol)
		if err != nil {
			if ShowLog {
				log.Warn("no market found", zap.String("symbol", item.Symbol))
			}
			skip[item.Symbol] = 0
			continue
		}
		if mar.Limits == nil || mar.Limits.Amount == nil {
			skip[item.Symbol] = 0
			continue
		}
		minAmt := mar.Limits.Amount.Min
		minCost := minAmt * item.Price
		if accCost < minCost {
			skip[item.Symbol] = minCost
		} else {
			res = append(res, item.Symbol)
		}
	}
	if len(skip) > 0 {
		var b strings.Builder
		for key, amt := range skip {
			b.WriteString(fmt.Sprintf("%s: %v  ", key, amt))
		}
		if ShowLog {
			log.Warn("skip symbols as cost too big", zap.Int("num", len(skip)), zap.String("more", b.String()))
		}
	}
	return res, skip
}

func (f *VolumePairFilter) GenSymbols(tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	symbols := make([]string, 0)
	if f.NeedTickers && len(tickers) > 0 {
		for symbol, _ := range tickers {
			symbols = append(symbols, symbol)
		}
	} else {
		markets := exg.Default.GetCurMarkets()
		symbols = utils.KeysOfMap(markets)
	}
	if len(symbols) == 0 {
		return nil, errs.NewMsg(errs.CodeRunTime, "no symbols generate from VolumePairFilter")
	}
	return f.Filter(symbols, tickers)
}

func (f *PriceFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	if core.LiveMode {
		res := make([]string, 0, len(symbols))
		if len(tickers) == 0 {
			log.Warn("no tickers, PriceFilter skipped")
			return symbols, nil
		}
		for _, s := range symbols {
			tik, ok := tickers[s]
			if !ok {
				continue
			}
			if f.validatePrice(s, tik.Last) {
				res = append(res, s)
			}
		}
		return res, nil
	} else {
		return filterByOHLCV(symbols, "1h", 1, core.AdjFront, func(s string, klines []*banexg.Kline) bool {
			if len(klines) == 0 {
				return f.AllowEmpty
			}
			return f.validatePrice(s, klines[len(klines)-1].Close)
		})
	}
}

func (f *PriceFilter) validatePrice(symbol string, price float64) bool {
	exchange := exg.Default
	if f.Precision > 0 {
		pip, err := exchange.PriceOnePip(symbol)
		if err != nil {
			log.Error("get one pip of price fail", zap.String("symbol", symbol))
			return false
		}
		chgPrec := pip / price
		if chgPrec > f.Precision {
			log.Info("PriceFilter drop, 1 unit fail", zap.String("pair", symbol), zap.Float64("p", chgPrec))
			return false
		}
	}

	if f.MaxUnitValue > 0 {
		market, err := exchange.GetMarket(symbol)
		if err != nil {
			log.Error("PriceFilter drop, market not exist", zap.String("pair", symbol))
			return false
		}
		minPrec := market.Precision.Amount
		if minPrec > 0 {
			if market.Precision.ModeAmount != banexg.PrecModeTickSize {
				minPrec = math.Pow(0.1, minPrec)
			}
			unitVal := minPrec * price
			if unitVal > f.MaxUnitValue {
				log.Info("PriceFilter drop, unit value too small", zap.String("pair", symbol),
					zap.Float64("uv", unitVal))
				return false
			}
		}
	}

	if f.Min > 0 && price < f.Min {
		log.Info("PriceFilter drop, price too small", zap.String("pair", symbol), zap.Float64("price", price))
		return false
	}

	if f.Max > 0 && f.Max < price {
		log.Info("PriceFilter drop, price too big", zap.String("pair", symbol), zap.Float64("price", price))
		return false
	}
	return true
}

func (f *RateOfChangeFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	return filterByOHLCV(symbols, "1d", f.BackDays, core.AdjFront, f.validate)
}

func (f *RateOfChangeFilter) validate(pair string, arr []*banexg.Kline) bool {
	if len(arr) == 0 {
		return f.AllowEmpty
	}
	hhigh := arr[0].High
	llow := arr[0].Low
	for _, k := range arr[1:] {
		hhigh = max(hhigh, k.High)
		llow = min(llow, k.Low)
	}
	roc := float64(0)
	if llow > 0 {
		roc = (hhigh - llow) / llow
	}
	if f.Min > roc {
		log.Info("RateOfChangeFilter drop by min", zap.String("pair", pair), zap.Float64("roc", roc))
		return false
	}
	if f.Max > 0 && f.Max < roc {
		log.Info("RateOfChangeFilter drop by max", zap.String("pair", pair), zap.Float64("roc", roc))
		return false
	}
	return true
}

func filterByOHLCV(symbols []string, timeFrame string, limit int, adj int, cb func(string, []*banexg.Kline) bool) ([]string, *errs.Error) {
	var has = make(map[string]struct{})
	handle := func(pair string, _ string, arr []*banexg.Kline, adjs []*orm.AdjInfo) {
		arr = orm.ApplyAdj(adjs, arr, adj, 0, 0)
		if cb(pair, arr) {
			has[pair] = struct{}{}
		}
	}
	err := orm.FastBulkOHLCV(exg.Default, symbols, timeFrame, 0, 0, limit, handle)
	if err != nil {
		return nil, err
	}
	var res = make([]string, 0, len(has))
	for _, pair := range symbols {
		if _, ok := has[pair]; ok {
			res = append(res, pair)
		}
	}
	return res, nil
}

func (f *CorrelationFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	if f.Timeframe == "" || f.BackNum == 0 || f.Max == 0 && f.TopN == 0 {
		return symbols, nil
	}
	if f.BackNum < 10 {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "`CorrelationFilter.back_num` should >= 10, cur: %v", f.BackNum)
	}
	var skips []string
	var names = make([]string, 0, len(symbols))
	var dataArr = make([][]float64, 0, len(symbols))
	for _, pair := range symbols {
		exs, err := orm.GetExSymbolCur(pair)
		if err != nil {
			skips = append(skips, pair)
			continue
		}
		_, klines, err := orm.GetOHLCV(exs, f.Timeframe, 0, 0, f.BackNum, false)
		if err != nil || len(klines)*2 < f.BackNum {
			skips = append(skips, pair)
			continue
		}
		names = append(names, pair)
		prices := make([]float64, 0, len(klines))
		for _, b := range klines {
			prices = append(prices, b.Close)
		}
		if len(prices) > f.BackNum {
			prices = prices[:f.BackNum]
		}
		dataArr = append(dataArr, prices)
	}
	nameNum := len(names)
	if nameNum <= 3 {
		log.Warn("too less symbols, skip CorrelationFilter", zap.Int("num", nameNum))
		return symbols, nil
	}
	if len(skips) > 0 {
		log.Warn("skip for klines too less", zap.Strings("codes", skips))
	}
	mat, avgs, err_ := utils.CalcCorrMat(f.BackNum, dataArr, true)
	if err_ != nil {
		return nil, errs.New(errs.CodeRunTime, err_)
	}
	if f.Sort != "asc" && f.Sort != "desc" {
		// Use default sorting 使用默认排序
		result := make([]string, 0, nameNum)
		for i, avg := range avgs {
			if f.Min != 0 && avg < f.Min {
				continue
			}
			if f.Max != 0 && avg > f.Max {
				continue
			}
			result = append(result, names[i])
			if f.TopN > 0 && len(result) >= f.TopN {
				break
			}
		}
		return result, nil
	}
	// 按要求基于平均相似度排序
	arr := make([]*IdVal, 0, len(avgs))
	lefts := make(map[int]bool)
	for i, avg := range avgs {
		arr = append(arr, &IdVal{Id: i, Val: avg})
		lefts[i] = true
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Val < arr[j].Val
	})
	isAsc := f.Sort == "asc"
	var it *IdVal
	if isAsc {
		it = arr[0]
	} else {
		it = arr[len(arr)-1]
	}
	sels := make([]*IdVal, 0, len(avgs))
	sels = append(sels, it)
	delete(lefts, it.Id)
	for len(lefts) > 0 {
		// 针对每个剩余标的，计算与所有sels的平均相似度
		it = nil
		for id := range lefts {
			vals := make([]float64, 0, len(sels))
			for _, v := range sels {
				vals = append(vals, mat.At(id, v.Id))
			}
			avg := floats.Sum(vals) / float64(len(vals))
			if it == nil || isAsc && avg < it.Val || !isAsc && avg > it.Val {
				it = &IdVal{Id: id, Val: avg}
			}
		}
		sels = append(sels, &IdVal{Id: it.Id, Val: avgs[it.Id]})
		delete(lefts, it.Id)
	}
	// 按规则过滤
	result := make([]string, 0, nameNum)
	for _, item := range sels {
		if f.Min != 0 && item.Val < f.Min {
			continue
		}
		if f.Max != 0 && item.Val > f.Max {
			continue
		}
		result = append(result, names[item.Id])
		if f.TopN > 0 && len(result) >= f.TopN {
			break
		}
	}
	return result, nil
}

type IdVal struct {
	Id  int
	Val float64
}

func (f *VolatilityFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	return filterByOHLCV(symbols, "1d", f.BackDays, core.AdjFront, func(s string, klines []*banexg.Kline) bool {
		if len(klines) == 0 {
			return f.AllowEmpty
		}
		var data = make([]float64, 0, len(klines))
		for i, v := range klines[1:] {
			data = append(data, v.Close/klines[i].Close)
		}
		res := utils.StdDevVolatility(data, 1)
		if res < f.Min || f.Max > 0 && res > f.Max {
			log.Info("VolatilityFilter drop", zap.String("pair", s), zap.Float64("v", res))
			return false
		}
		return true
	})
}

func (f *SpreadFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	return symbols, nil
}

func (f *OffsetFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	var res = symbols
	if f.Reverse {
		slices.Reverse(res)
	}
	if f.Offset < len(res) {
		res = res[f.Offset:]
	}
	if f.Rate > 0 && f.Rate < 1 {
		num := int(math.Round(float64(len(res)) * f.Rate))
		res = res[:num]
	}
	if f.Limit > 0 && f.Limit < len(res) {
		res = res[:f.Limit]
	}
	return res, nil
}

func (f *ShuffleFilter) Filter(symbols []string, tickers map[string]*banexg.Ticker) ([]string, *errs.Error) {
	rand.Shuffle(len(symbols), func(i, j int) {
		symbols[i], symbols[j] = symbols[j], symbols[i]
	})
	return symbols, nil
}
