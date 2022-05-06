package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func main() {
	var messariApiKey string
	flag.StringVar(&messariApiKey, "messariApiKey", "default", "Specify Messari API Key")
	flag.Parse()
	for {
		fmt.Printf("apikey: %s\n", messariApiKey)
		MessariData(messariApiKey)
		time.Sleep(time.Hour * 24)
	}

}

func MessariData(apiKey string) {
	client := &http.Client{
		Timeout: time.Second * 600,
	}

	timestamp := time.Now().Format("2006-01-02")
	fmt.Printf("using timestamp: %s\n", timestamp)
	//"https://data.messari.io/api/v1/markets/kraken-btc-usd/metrics/price/time-series?start=2022-05-06&end=2022-05-06&interval=1d"
	url := "https://data.messari.io/api/v1/markets/kraken-btc-usd/metrics/price/time-series?start=" + timestamp + "&end=" + timestamp + "&interval=1d"
	fmt.Printf("url: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Got error in request %s", err.Error())
		return
	}
	req.Header.Add("x-messari-api-key", apiKey)
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error in do %s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error with read: %s\n", err.Error())
		log.Fatalln(err)
	}
	defer response.Body.Close()
	vals := gjson.GetBytes(body, "data.values").Array()
	for _, v := range vals {
		timestamp := v.Array()[0].String()[0:10]
		openPrice := v.Array()[1].String()
		fmt.Printf("timestamp: %s openPrice: %s\n", timestamp, openPrice)
		UpdatePriceData(timestamp, openPrice)
	}
}

func UpdatePriceData(timestamp, openPrice string) error {
	content, err := os.ReadFile("../PriceDataKraken.json")
	if err != nil {
		fmt.Printf("Error reading PriceDataKraken.json: %s\n", err.Error())
		return err
	}
	ts, err := strconv.ParseFloat(string(timestamp), 64)
	if err != nil {
		fmt.Printf("Error parsing timestamp. timestamp: %s Error: %s\n", timestamp, err.Error())
		return err
	}
	op, err := strconv.ParseFloat(string(openPrice), 64)
	if err != nil {
		fmt.Printf("Error parsing openPrice. openPrice: %s Error: %s\n", openPrice, err.Error())
		return err
	}
	value, err := sjson.Set(string(content), "data.-1", map[string]float64{"timestamp": ts, "openPrice": op})
	if err != nil {
		fmt.Printf("er with sjson set: %s\n", err.Error())
		return err
	}
	// fmt.Printf("value: %s\n", value)

	// open output file
	fo, err := os.Create("../PriceDataKraken.json")
	if err != nil {
		fmt.Printf("error creating ../PriceDataKraken.json with error: %s\n", err.Error())
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		err := fo.Close()
		if err != nil {
			fmt.Printf("error closing ../PriceDataKraken.json with error: %s\n", err.Error())
			return
		}
	}()

	if _, err := fo.Write([]byte(value)); err != nil {
		fmt.Printf("error writing to ../PriceDataKraken.json with error: %s\n", err.Error())
		return err
	}
	return nil
}
