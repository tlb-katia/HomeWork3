# Pipeline Processing and Candle Formation in Go


## Table of Contents
1. [Introduction](#introduction)
2. [Pipeline Stages](#pipeline_stages)
   1. [1-Minute Candles](1_minute_candles)
   2. [2-Minute Candles](2_minute_candles)
   3. [10-Minute Candles](10_minute_candles)
6. [Graceful Shutdown](#graceful_shutdown)
7. [Data Structures ](#data_structures)
   1. [Price](#price) 
   2. [Candle](#candle)
   3. [CandlePeriod](#Candleperiod)
11. [File Output Example](#file_output_example)
12. [Running the Application](#running_the_application)



## Introduction
This project implements a pipeline for processing price data and forming candles at different intervals (1-minute, 2-minute, and 10-minute). The pipeline stages include:

1. Forming 1-minute candles from the price stream.
2. Forming 2-minute candles from the stream of 1-minute candles.
3. Forming 10-minute candles from the stream of 2-minute candles.


Each type of candle is saved in two files:

- A CSV log file.

The program should gracefully shut down when receiving a SIGINT (Ctrl+C), ensuring that the current price being processed completes the entire pipeline.

## Pipeline Stages
### 1-Minute Candles
- **Input:** Stream of prices.
- **Output:**
  - Log file: candles_1m_log.csv
### 2-Minute Candles
- **Input:** Stream of 1-minute candles.
- **Output:**
  - Log file: candles_2m_log.csv
  
### 10-Minute Candles
**Input:** Stream of 2-minute candles.
- **Output:**
    - Log file: candles_10m_log.csv
  
## Graceful Shutdown
The program should handle SIGINT (Ctrl+C) to ensure that:

- The current price in the pipeline completes processing.
- All candles are saved before shutdown.

## Data Structures
### Price
The `Price` struct represents a price point with a ticker, value, and timestamp.

``` go
type Price struct {
    Ticker string
    Value  float64
    TS     time.Time
}
```

### Candle
The `Candle` struct represents a candle formed from price data.

``` go
type Candle struct {
    Ticker string
    Period CandlePeriod // Interval
    Open   float64      // Opening price
    High   float64      // Highest price
    Low    float64      // Lowest price
    Close  float64      // Closing price
    TS     time.Time    // Interval start time
}
```

### CandlePeriod
The `CandlePeriod` type defines the intervals for candles and a function to truncate timestamps to the start of each period.

``` go
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
``` 
## File Output Examples
### candles_1m_log.csv
``` csv
AAPL,2021-09-06T16:07:00+03:00,112.093206,112.093206,101.939390,104.285277
SBER,2021-09-06T16:07:00+03:00,218.810182,218.810182,206.018237,207.613144
NVDA,2021-09-06T16:07:00+03:00,313.291201,313.291201,301.312740,306.361163
TSLA,2021-09-06T16:07:00+03:00,408.754284,416.272799,403.130385,409.377797
AAPL,2021-09-06T16:08:00+03:00,105.660683,117.887235,103.166566,117.887235
SBER,2021-09-06T16:08:00+03:00,205.862037,213.934383,201.182413,201.949092
NVDA,2021-09-06T16:08:00+03:00,313.581694,319.538337,305.070810,319.538337
TSLA,2021-09-06T16:08:00+03:00,404.371061,417.249829,400.566062,401.485820
AAPL,2021-09-06T16:09:00+03:00,104.445788,119.578587,101.710410,108.698249
SBER,2021-09-06T16:09:00+03:00,213.621566,219.098909,202.535059,212.501901
```

## Running the Application
1. Ensure Go is installed on your machine.
2. Clone the repository and navigate to the project directory.
3. Run the application:
``` sh
Копировать код
go run main.go
```
4. To stop the application, press Ctrl+C. The program will ensure all in-progress data is processed and saved before shutting down.