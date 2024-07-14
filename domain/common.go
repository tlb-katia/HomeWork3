package domain

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"time"
)

type Price struct {
	Ticker string
	Value  float64
	TS     time.Time
}

var ErrUnknownPeriod = errors.New("unknown period")

type CandlePeriod string

const (
	CandlePeriod1m  CandlePeriod = "1m"
	CandlePeriod2m  CandlePeriod = "2m"
	CandlePeriod10m CandlePeriod = "10m"
)

func PeriodTS(period CandlePeriod, ts time.Time) (time.Time, error) {
	switch period {
	case CandlePeriod1m:
		return ts.Truncate(time.Minute), nil
	case CandlePeriod2m:
		return ts.Truncate(2 * time.Minute), nil
	case CandlePeriod10m:
		return ts.Truncate(10 * time.Minute), nil
	default:
		return time.Time{}, ErrUnknownPeriod
	}
}

type Candle struct {
	Ticker string
	Period CandlePeriod // Интервал
	Open   float64      // Цена открытия
	High   float64      // Максимальная цена
	Low    float64      // Минимальная цена
	Close  float64      // Цена закрытие
	TS     time.Time    // Время начала интервала
}

func Create1mCandles(inside <-chan Price, ctx context.Context) <-chan Candle {
	out := make(chan Candle)

	go func() {
		defer close(out)
		candleMap := make(map[string]map[time.Time][]Price)

		for {
			select {
			case <-ctx.Done():
				return
			case price, ok := <-inside:
				if !ok {
					return
				}
				startTime, err := PeriodTS(CandlePeriod1m, price.TS)
				if err != nil {
					return
				}
				if candleMap[price.Ticker] == nil {
					candleMap[price.Ticker] = make(map[time.Time][]Price)
				}
				candleMap[price.Ticker][startTime] = append(candleMap[price.Ticker][startTime], price)

				for ticker, timeGroups := range candleMap {
					var prevTS time.Time
					for startTime, _ := range timeGroups {
						prevTS = startTime
						break
					}
					for startTime, prices := range timeGroups {
						if prevTS.Sub(startTime) == time.Minute {
							out <- calculateCandle(prices, startTime, CandlePeriod1m)
							//function call to write down candles
							delete(candleMap[ticker], startTime)
						}
					}
				}
			}

		}
	}()

	return out
}

func Create2mCandles(inside <-chan Candle, ctx context.Context) <-chan Candle {
	out := make(chan Candle)

	go func() {
		defer close(out)
		candleMap := make(map[string]map[time.Time][]Candle)

		for candle := range inside {
			ticker := candle.Ticker
			timeStamp, err := PeriodTS(CandlePeriod2m, candle.TS)
			if err != nil {
				return
			}
			if candleMap[ticker] == nil {
				candleMap[ticker] = make(map[time.Time][]Candle)
			}
			candleMap[ticker][timeStamp] = append(candleMap[ticker][timeStamp], candle)

			if len(candleMap[ticker][timeStamp]) == 2 {
				prices := []Price{
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][0].Open, TS: candleMap[ticker][timeStamp][0].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][0].Close, TS: candleMap[ticker][timeStamp][0].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][0].High, TS: candleMap[ticker][timeStamp][0].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][0].Low, TS: candleMap[ticker][timeStamp][0].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][1].Open, TS: candleMap[ticker][timeStamp][1].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][1].Close, TS: candleMap[ticker][timeStamp][1].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][1].High, TS: candleMap[ticker][timeStamp][1].TS,
					},
					{
						Ticker: ticker, Value: candleMap[ticker][timeStamp][1].Low, TS: candleMap[ticker][timeStamp][1].TS,
					},
				}
				out <- calculateCandle(prices, timeStamp, CandlePeriod2m)
				delete(candleMap[ticker], timeStamp)
			}
		}
	}()
	return out
}

func Create10mCandles(inside <-chan Candle, ctx context.Context) <-chan Candle {
	out := make(chan Candle)

	go func() {
		defer close(out)
		candleMap := make(map[string]map[time.Time][]Candle)
		for candle := range inside {
			startTime, err := PeriodTS(CandlePeriod10m, candle.TS)
			if err != nil {
				return
			}
			ticker := candle.Ticker
			if candleMap[ticker] == nil {
				candleMap[ticker] = make(map[time.Time][]Candle)
			}
			candleMap[ticker][startTime] = append(candleMap[ticker][startTime], candle)

			if len(candleMap[ticker][startTime]) == 5 {
				prices := make([]Price, 0)
				for i := 0; i < 5; i++ {
					priceStructs := []Price{
						{Ticker: ticker, Value: candleMap[ticker][startTime][i].Open, TS: candleMap[ticker][startTime][i].TS},
						{Ticker: ticker, Value: candleMap[ticker][startTime][i].Close, TS: candleMap[ticker][startTime][i].TS},
						{Ticker: ticker, Value: candleMap[ticker][startTime][i].High, TS: candleMap[ticker][startTime][i].TS},
						{Ticker: ticker, Value: candleMap[ticker][startTime][i].Low, TS: candleMap[ticker][startTime][i].TS},
					}
					for _, p := range priceStructs {
						prices = append(prices, p)
					}
				}
				out <- calculateCandle(prices, startTime, CandlePeriod10m)
				delete(candleMap[ticker], startTime)
			}
		}
	}()

	return out
}

func calculateCandle(prices []Price, startTime time.Time, period CandlePeriod) Candle {
	openPrice := prices[0].Value
	closePrice := prices[0].Value
	timeBegin := prices[0].TS
	timeEnd := prices[0].TS
	highPrice := openPrice
	lowPrice := openPrice

	for _, price := range prices {
		if price.TS.Sub(timeBegin) < 0 {
			openPrice = price.Value
		}
		if price.TS.Sub(timeEnd) > 0 {
			closePrice = price.Value
		}
		if price.Value > highPrice {
			highPrice = price.Value
		}
		if price.Value < lowPrice {
			lowPrice = price.Value
		}

	}

	candle := Candle{
		Ticker: prices[0].Ticker,
		TS:     startTime,
		Open:   openPrice,
		Close:  closePrice,
		High:   highPrice,
		Low:    lowPrice,
	}

	writeToCSV(candle, period)

	return candle
}

func writeToCSV(candle Candle, period CandlePeriod) error {
	fileName := fmt.Sprintf("candles_%s_log.csv", period)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(candle.Slice()); err != nil {
		return fmt.Errorf("cannot write the data: %v", err)
	}
	return writer.Error()
}

func (c Candle) Slice() []string {
	return []string{
		c.Ticker,
		c.TS.String(),
		fmt.Sprintf("%.2f", c.Open),
		fmt.Sprintf("%.2f", c.Close),
		fmt.Sprintf("%.2f", c.High),
		fmt.Sprintf("%.2f", c.Low),
	}
}
