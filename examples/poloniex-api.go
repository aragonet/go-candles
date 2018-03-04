package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"time"

	gc "github.com/carlesar/go-candles"
)

func main() {
	now := time.Now()
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)
	stmt := fmt.Sprintf("https://poloniex.com/public?command=returnChartData&currencyPair=BTC_BTS&start=%d&end=%d&period=14400", oneWeekAgo.Unix(), now.Unix())

	resp, err := http.Get(stmt)
	if err != nil {
		fmt.Println("Error making api request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Could not parse response information")
			return
		}

		var candlesData []gc.Candle
		err = json.Unmarshal(bodyBytes, &candlesData)
		if err != nil {
			fmt.Printf("Response type not expected. Error: %v\n", err)
			return
		}
		ioutil.WriteFile("data", bodyBytes, 0644)

		gc.CreateChart(candlesData, gc.Options{
			LinesChartColor:      color.RGBA{255, 255, 0, 255},
			BackgroundChartColor: color.RGBA{0, 0, 0, 255},
			YLabelText:           "BitShares coin (BTS)",
			YLabelColor:          color.RGBA{255, 255, 255, 255},
			PositiveCandleColor:  color.RGBA{0, 255, 0, 255},
			NegativeCandleColor:  color.RGBA{255, 0, 0, 255},
			PikeCandleColor:      color.RGBA{211, 211, 211, 255},
			Width:                800,
			Height:               600,
			CandleWidth:          4,
			Rows:                 5,
			Columns:              7,
			OutputFileName:       "out.png",
		})
	}
}
