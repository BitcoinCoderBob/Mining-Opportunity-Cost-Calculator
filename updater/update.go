package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
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
		// t := GetLastTimeStamp()
		// fmt.Printf("t: %s", t)
		time.Sleep(time.Hour * 24)
	}

}

func MessariData(apiKey string) {
	client := &http.Client{
		Timeout: time.Second * 600,
	}

	currentTimestamp := time.Now().Format("2006-01-02")
	fmt.Printf("using currentTimestamp: %s\n", currentTimestamp)
	lastTimestamp, lastUnixTimestemp := GetLastTimeStamp()
	fmt.Printf("using last timestamp: %s\n", lastTimestamp)
	//"https://data.messari.io/api/v1/markets/kraken-btc-usd/metrics/price/time-series?start=2022-05-06&end=2022-05-06&interval=1d"
	url := "https://data.messari.io/api/v1/markets/kraken-btc-usd/metrics/price/time-series?start=" + lastTimestamp + "&end=" + currentTimestamp + "&interval=1d"
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
	toAdd := map[string]string{}
	for _, v := range vals {
		timestamp := v.Array()[0].String()[0:10]
		openPrice := v.Array()[1].String()
		// fmt.Printf("timestamp: %s openPrice: %s\n", timestamp, openPrice)
		toAdd[timestamp] = openPrice
	}

	sortedToAdd, err := OrderData(toAdd)
	if err != nil {
		fmt.Printf("error sorting price data: %s\n", err.Error())
	}

	UpdatePriceData(sortedToAdd, lastUnixTimestemp)

}

// order data by timestamps
func OrderData(data map[string]string) ([]map[string]string, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make([]map[string]string, 0, len(data))
	for _, k := range keys {
		sorted = append(sorted, map[string]string{k: data[k]})
	}
	return sorted, nil
}

func GetLastTimeStamp() (string, string) {
	content, err := os.ReadFile("../PriceDataKraken.json")
	if err != nil {
		fmt.Printf("Error reading ../PriceDataKraken.json: %s\n", err.Error())
	}
	vals := gjson.GetBytes(content, "data").Array()
	fmt.Printf("len vals: %d\n", len(vals))
	lastTimestampUnixString := vals[len(vals)-1].Get("timestamp").String()
	fmt.Printf("found: %s\n", lastTimestampUnixString)
	// unixTime, err := time.Parse("1136239445", lastTimestampUnixString)
	// if err != nil {
	// 	fmt.Printf("error parsing unix time: %s\n", err.Error())
	// }
	i, err := strconv.ParseInt(lastTimestampUnixString, 10, 64)
	if err != nil {
		panic(err)
	}
	unixTime := time.Unix(i, 0)

	f := unixTime.Format("2006-01-02")
	// dateString, err := time.Parse("2006-01-02", unixTime.String()[0:10])
	// if err != nil {
	// 	fmt.Printf("error doing second time parse on unix string: %s\n", err.Error())
	// }
	// return dateString.String()
	return f, lastTimestampUnixString
}

func UpdatePriceData(toAdd []map[string]string, lastUnixTimestamp string) error {
	content, err := os.ReadFile("../PriceDataKraken.json")
	if err != nil {
		fmt.Printf("Error reading PriceDataKraken.json: %s\n", err.Error())
		return err
	}
	oldPriceData := string(content)
	added := 0
	for _, m := range toAdd {
		for timestamp, openPrice := range m {
			timestampToAddInt, err := strconv.Atoi(timestamp)
			if err != nil {
				fmt.Printf("error converting timestamp to int: %s\n", err.Error())
				return err
			}
			lastTimestampInt, err := strconv.Atoi(lastUnixTimestamp)
			if err != nil {
				fmt.Printf("error converting timestamp to int: %s\n", err.Error())
				return err
			}
			if lastTimestampInt >= timestampToAddInt {
				continue
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
			oldPriceData, err = sjson.Set(oldPriceData, "data.-1", map[string]float64{"timestamp": ts, "openPrice": op})
			if err != nil {
				fmt.Printf("er with sjson set: %s\n", err.Error())
				return err
			}
			added++
		}
	}

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

	if _, err := fo.Write([]byte(oldPriceData)); err != nil {
		fmt.Printf("error writing to ../PriceDataKraken.json with error: %s\n", err.Error())
		return err
	}
	fmt.Printf("added %d new price points\n", added)
	return nil
}
