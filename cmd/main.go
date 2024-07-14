package main

import (
	"HomeWork3/domain"
	"HomeWork3/generator"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func main() {
	logger := log.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exit := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier is not blocked
	signal.Notify(exit, os.Interrupt)

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	logger.Info("start prices generator...")
	prices := pg.Prices(ctx)

	oneMinuteCandles := domain.Create1mCandles(prices, ctx)
	twoMinuteCandles := domain.Create2mCandles(oneMinuteCandles, ctx)
	domain.Create10mCandles(twoMinuteCandles, ctx)

	<-exit
	logger.Info("exit")
	cancel()

}
