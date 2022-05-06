package externaldata

import (
	"Mining-Profitability/pkg/config"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

type Client struct {
	PriceDataKrakenPath string
	MessariUrl          string
	BlockchainInfoUrl   string
	SlushPoolUrl        string
	httpClient          *http.Client
}

type Interface interface {
	MessariData(apiKey string)
	GetBitcoinPrice() (price float64, err error)
	GetUserMinedCoinsTotal(token string) (coins float64, err error)
	GetPriceDataFromDateRange(start string) (priceData []float64)
}

func New(cfg *config.Config) *Client {
	return &Client{
		PriceDataKrakenPath: cfg.PriceDataKrakenPath, // "PriceDataKraken.json",
		MessariUrl:          cfg.MessariUrl,          //  "https://data.messari.io/api/v1/markets/kraken-btc-usd/metrics/price/time-series?start=2021-08-17&end=2021-08-19&interval=1d",
		BlockchainInfoUrl:   cfg.BlockchainInfoUrl,   // "https://blockchain.info/tobtc?currency=USD&value=500",
		SlushPoolUrl:        cfg.SlushPoolUrl,        //  "https://slushpool.com/accounts/profile/json/btc/",
		httpClient: &http.Client{
			Timeout: time.Second * 600,
		},
	}
}

func (c *Client) MessariData(apiKey string) {

	req, err := http.NewRequest("GET", c.MessariUrl, nil)
	if err != nil {
		fmt.Printf("Got error in request %s", err.Error())
		return
	}
	req.Header.Add("x-messari-api-key", apiKey)
	response, err := c.httpClient.Do(req)
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
	// data := map[int]float64{}
	vals := gjson.GetBytes(body, "data.values").Array()
	fmt.Printf("num vals: %d\n", len(vals))
	for _, v := range vals {
		timestamp := v.Array()[0].String()[0:10]
		openPrice := v.Array()[1].String()
		// i, err := strconv.ParseInt(timestamp, 10, 64)
		// if err != nil {
		// 	return
		// }
		// tm := time.Unix(i, 0)
		// // fmt.Printf("%v\n", tm)
		// fmt.Printf("timestamp: %s\n", fmt.Sprintf("%v\n", tm))
		fmt.Printf("timestamp: %s openPrice: %s\n", timestamp, openPrice)

	}
}

func (c *Client) GetBitcoinPrice() (price float64, err error) {
	req, err := http.NewRequest("GET", c.BlockchainInfoUrl, nil)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
		return
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer response.Body.Close()
	s, err := strconv.ParseFloat(string(body), 64)
	if err != nil {
		return
	}
	price = 500 / s
	return
}

func (c *Client) GetUserMinedCoinsTotal(token string) (coins float64, err error) {

	req, err := http.NewRequest("GET", c.SlushPoolUrl, nil)
	if err != nil {
		return -1, fmt.Errorf("error got error making request to %s error: %w", c.SlushPoolUrl, err)
	}
	req.Header.Set("SlushPool-Auth-Token", token)
	response, err := c.httpClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("error Got error doing request to slush endpoint %s error: %w", c.SlushPoolUrl, err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("error rading body from http call to %s error: %w", c.SlushPoolUrl, err)
	}
	defer response.Body.Close()
	value := gjson.GetBytes(body, "btc")
	allTimeReward, err := strconv.ParseFloat(value.Get("all_time_reward").String(), 64)
	if err != nil {
		return -1, fmt.Errorf("error converting all_time_reward to float: %w", err)
	}

	unconfirmedCoins, err := strconv.ParseFloat(value.Get("unconfirmed_reward").String(), 64)
	if err != nil {
		return -1, fmt.Errorf("error converting unconfirmed_reward to float: %w", err)
	}

	coins = allTimeReward + unconfirmedCoins
	return coins, err
}

func (c *Client) GetPriceDataFromDateRange(start string) (priceData []float64) {
	content, err := os.ReadFile(c.PriceDataKrakenPath)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", c.PriceDataKrakenPath, err.Error())
	}
	vals := gjson.GetBytes(content, "data").Array()
	foundStart := false
	for _, v := range vals {
		timestamp := v.Get("timestamp").String()
		price := v.Get("openPrice").Float()
		if timestamp == start {
			foundStart = true
		}
		if foundStart {
			// i, err := strconv.ParseInt(v.Get("timestamp").String(), 10, 64)
			// if err != nil {
			// 	return
			// }
			// tm := time.Unix(i, 0)
			// fmt.Printf("%v\n", tm)
			priceData = append(priceData, price)
		}
		// fmt.Printf("timestamp: %s price: %v\n", timestamp, price)
	}
	return
}
