package data

import (
	"github.com/banbox/banbot/exg"
	"github.com/banbox/banbot/orm"
	"github.com/banbox/banbot/utils"
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	utils2 "github.com/banbox/banexg/utils"
)

/*
Check whether there are any missing K lines, and automatically query and update if there are any.
检查是否有缺失的K线，有则自动查询更新（一般在刚启动时，收到的爬虫推送1mK线不含前面的，需要下载前面的并保存到WaitBar中）
*/
func (j *PairTFCache) fillLacks(pair string, subTfSecs int, startMS, endMS int64) ([]*banexg.Kline, *errs.Error) {
	if j.NextMS == 0 || j.NextMS >= startMS {
		j.NextMS = endMS
		return nil, nil
	}
	// 这里NextMS < startMS，出现了bar缺失，查询更新。
	exs, err := orm.GetExSymbolCur(pair)
	if err != nil {
		return nil, err
	}
	exchange := exg.Default
	if !exchange.HasApi(banexg.ApiFetchOHLCV, exs.Market) {
		// Downloading K lines is currently not allowed, skip
		// 当前不允许下载K线，跳过
		j.NextMS = endMS
		return nil, nil
	}
	fetchTF := utils2.SecsToTF(subTfSecs)
	tfMSecs := int64(j.TFSecs * 1000)
	bigStartMS := utils2.AlignTfMSecs(j.NextMS, tfMSecs)
	_, preBars, err := orm.AutoFetchOHLCV(exchange, exs, fetchTF, bigStartMS, startMS, 0, false, nil)
	if err != nil {
		return nil, err
	}
	var doneBars []*banexg.Kline
	j.WaitBar = nil
	if len(preBars) > 0 {
		fromTFMS := int64(subTfSecs * 1000)
		oldBars, _ := utils.BuildOHLCV(preBars, tfMSecs, 0, nil, fromTFMS, j.AlignOffMS)
		if len(oldBars) > 0 {
			j.WaitBar = oldBars[len(oldBars)-1]
			doneBars = oldBars[:len(oldBars)-1]
		}
	}
	j.NextMS = endMS
	return doneBars, nil
}
