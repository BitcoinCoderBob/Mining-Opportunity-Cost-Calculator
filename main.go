package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	var slushToken, messariApiKey, startDate string
	var kwhPrice, watts, uptimePercent, fixedCosts, bitcoinMined float64
	flag.StringVar(&slushToken, "slushToken", "default-token", "Specify Slush Pool token.")
	flag.Float64Var(&kwhPrice, "kwhPrice", 0.15, "Specify price paid per kilowatt hour.")
	flag.Float64Var(&watts, "watts", 3200, "Specify watts used in total.")
	flag.Float64Var(&uptimePercent, "uptimePercent", 100.0, "Specify percent uptime of your miners.")
	flag.Float64Var(&fixedCosts, "fixedCosts", 6295.55, "Specify mining setup fix costs.")
	flag.Float64Var(&bitcoinMined, "bitcoinMined", 0, "Specify total bitcoin mined (use whole bitcoin units not sats).")
	flag.StringVar(&startDate, "startDate", "01/01/2022", "Specify start date of mining operation.")
	flag.StringVar(&messariApiKey, "messariApiKey", "default", "Specify Messari API Key")
	flag.Parse()
	if slushToken == "default-token" && bitcoinMined == 0 {
		fmt.Printf("Must enter either slush api token or bitcoinMined")
	}
	price, err := GetBitcoinPrice()
	if err != nil {
		fmt.Printf("Error getting bitcoin price: %s\n", err.Error())
		return
	}
	fmt.Printf("Bicoin current price: $%s\n", fmt.Sprintf("%.2f", price))
	daysSinceStart, err := DaysSinceStart(startDate)
	if err != nil {
		fmt.Printf("Error calculating days since start: %s\n", err.Error())
		return
	}
	fmt.Printf("Days since start: %s\n", fmt.Sprintf("%.2f", daysSinceStart))

	if slushToken != "default-token" {
		bitcoinMined, err = GetUserMinedCoinsTotal(slushToken)
		if err != nil {
			fmt.Printf("Error GetUseRMinedCoinsTotal: %s\n", err.Error())
		}
	}
	fmt.Printf("Average coins per day: %s\n", fmt.Sprintf("%.8f", AverageCoinsPerDay(daysSinceStart, bitcoinMined)))
	dollarinosEarned := DollarinosEarned(bitcoinMined, price)
	fmt.Printf("Dollarinos earned: $%s\n", fmt.Sprintf("%.2f", dollarinosEarned))
	electricCosts := ElectricCosts(kwhPrice, uptimePercent, daysSinceStart, watts)
	fmt.Printf("Total electric costs: $%s\n", fmt.Sprintf("%.2f", electricCosts))
	percentPaidOff := PercentPaidOff(dollarinosEarned, fixedCosts, electricCosts)
	fmt.Printf("Percent paid off: %s%%\n", fmt.Sprintf("%.2f", percentPaidOff))
	fmt.Printf("Bitcoin percentage increase needed to be breakeven: %s%%\n", fmt.Sprintf("%.2f", ((100/percentPaidOff)-1)*100))
	breakevenPrice := BreakEvenPrice(percentPaidOff, price)
	fmt.Printf("Breakeven price: $%s\n", fmt.Sprintf("%.2f", breakevenPrice))
	daysUntilBreakeven := DaysUntilBreakeven(daysSinceStart, percentPaidOff)
	fmt.Printf("Expected more days until breakeven: %s\n", fmt.Sprintf("%.2f", daysUntilBreakeven))
	fmt.Printf("Total mining days (past + future) to breakeven: %s\n", fmt.Sprintf("%.2f", daysUntilBreakeven+daysSinceStart))
	futureDate, err := DateFromDaysNow(daysUntilBreakeven)
	if err != nil {
		fmt.Printf("Error with DateFromDaysNow. Error: %s\n", err.Error())
		return
	}
	fmt.Printf("Expected breakeven date: %s\n", futureDate)
	fmt.Printf("\n\n------------------------------------------------\n\n")
	dailyElectricCost := electricCosts / daysSinceStart
	fmt.Printf("Electric costs per day: $%s\n", fmt.Sprintf("%.2f", dailyElectricCost))
	unixTimeStampStart, err := DateToUnixTimestamp(startDate)
	if err != nil {
		fmt.Printf("Error with DateToUnixTimestamp: %s\n", err.Error())
	}
	priceData := GetPriceDataFromDateRange(unixTimeStampStart)
	totalDollarsSpent := electricCosts + fixedCosts
	unixDaysSinceStart, err := RegularDateToUnix(startDate)
	if err != nil {
		fmt.Printf("error with RegularDateToUnix: %s\n", err)
	}
	// fmt.Printf("a: %v\n", a)
	// daysSinceStartUnix, err := DaysSinceStartUnixTimestamp("")
	// if err != nil {
	// 	fmt.Printf("error with unix timestmap: %s\n", err.Error())
	// }
	fmt.Printf("bitcoin mined: %v\n", bitcoinMined)
	swanData, swanSats := SwanDailyDCABuy(totalDollarsSpent, unixDaysSinceStart, priceData)
	ahData, ahSats := AmericanHodlSlamBuy(totalDollarsSpent, priceData[0], len(priceData))
	fmt.Printf("americanHodlSats: %v\n", ahSats)
	fmt.Printf("swanSats: %v\n", swanSats)
	// MessariData(messariApiKey)
	antiHomeMinerData, antiHomeMinerSats := AntiHomeMiner(fixedCosts, electricCosts, unixDaysSinceStart, priceData)
	fmt.Printf("antiHomeMinerSats: %v\n", antiHomeMinerSats)
	MakePlot(ahData, swanData, antiHomeMinerData, bitcoinMined)
}

func DateToUnixTimestamp(start string) (timestamp string, err error) {
	t, err := time.Parse("01/02/2006", start)
	if err != nil {
		return
	}
	b := t.Unix()
	return strconv.FormatInt(b, 10), err
}

func MessariData(apiKey string) {
	client := &http.Client{
		Timeout: time.Second * 600,
	}
	req, err := http.NewRequest("GET", "https://data.messari.io/api/v1/markets/coinbase-btc-usd/metrics/price/time-series?start=2021-08-17&end=2021-08-19&interval=1d", nil)
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
	return
}

func GetBitcoinPrice() (price float64, err error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", "https://blockchain.info/tobtc?currency=USD&value=500", nil)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
		return
	}
	response, err := client.Do(req)
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

func AverageCoinsPerDay(days, coins float64) (averageCoinsPerDay float64) {
	return coins / days
}

func DollarinosEarned(coins, price float64) (dollarinos float64) {
	return coins * price
}

func ElectricCosts(kwhPrice, uptimePercentage, uptimeDays, watts float64) (electricCosts float64) {
	kwhPerDay := watts * 24 / 1000
	electricCosts = kwhPrice * kwhPerDay * (uptimePercentage / 100) * uptimeDays
	return
}

func PercentPaidOff(dollarinosEarned, fixedCosts, variableCosts float64) (percentPaidOff float64) {
	return dollarinosEarned / (fixedCosts + variableCosts) * 100
}

func DaysSinceStart(startDate string) (days float64, err error) {
	t, err := time.Parse("01/02/2006", startDate)
	if err != nil {
		return
	}
	durationSinceStart := time.Since(t)
	days = durationSinceStart.Hours() / 24
	return
}

func RegularDateToUnix(start string) (days float64, err error) {
	t, err := time.Parse("01/02/2006", start)
	if err != nil {
		return
	}
	b := t.Unix()
	tm := time.Unix(b, 0)
	durationSinceStart := time.Since(tm)
	days = durationSinceStart.Hours() / 24
	return math.Floor(days), err
}

func DaysSinceStartUnixTimestamp(startDate string) (days float64, err error) {
	i, err := strconv.ParseInt(startDate, 10, 64)
	if err != nil {
		return
	}
	tm := time.Unix(i, 0)
	// t, err := time.Parse("1136239445", "1405544146")
	// if err != nil {
	// 	return
	// }
	durationSinceStart := time.Since(tm)
	days = durationSinceStart.Hours() / 24
	return math.Floor(days), err
}

func BreakEvenPrice(percentPaidOff, bitcoinPrice float64) (breakevenPrice float64) {
	return bitcoinPrice * (1 / (percentPaidOff / 100))
}

func GetUserMinedCoinsTotal(token string) (coins float64, err error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", "https://slushpool.com/accounts/profile/json/btc/", nil)
	if err != nil {
		fmt.Printf("Got error making request to https://slushpool.com/accounts/profile/json/btc/ Error: %s\n", err.Error())
		return
	}
	req.Header.Set("SlushPool-Auth-Token", token)
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error doing request to slush endpoint https://slushpool.com/accounts/profile/json/btc/ Error: %s\n", err.Error())
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error rading body from http call to https://slushpool.com/accounts/profile/json/btc/  Error: %s\n", err.Error())
		return
	}
	defer response.Body.Close()
	value := gjson.GetBytes(body, "btc")
	allTimeReward, err := strconv.ParseFloat(value.Get("all_time_reward").String(), 64)
	if err != nil {
		fmt.Printf("Error converting all_time_reward to float: %s\n", err.Error())
		return
	}

	unconfirmedCoins, err := strconv.ParseFloat(value.Get("unconfirmed_reward").String(), 64)
	if err != nil {
		fmt.Printf("Error converting unconfirmed_reward to float: %s\n", err.Error())
		return
	}

	coins = allTimeReward + unconfirmedCoins
	return coins, err
}

func DaysUntilBreakeven(daysSinceStart, percentPaidOff float64) (moreDays float64) {
	moreDays = (daysSinceStart * (1 / (percentPaidOff / 100))) - daysSinceStart
	return
}

func DateFromDaysNow(days float64) (futureDate string, err error) {
	hours := days * 24
	hourDuration, err := time.ParseDuration(fmt.Sprintf("%f", hours) + "h")
	if err != nil {
		return
	}
	futureTime := time.Now().Add(hourDuration)
	futureDate = futureTime.Format("01/02/2006")
	return
}

func GetPriceDataFromDateRange(start string) (priceData []float64) {
	content, err := os.ReadFile("PriceDataKraken.json")
	if err != nil {
		fmt.Printf("Error reading PriceDataKraken.json: %s\n", err.Error())
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

func AmericanHodlSlamBuy(dollarsAvailable, openPrice float64, numberDays int) (cumulativeTotal []float64, bitcoinAcquired float64) {
	bitcoinAcquired = dollarsAvailable / openPrice
	for i := 0; i < numberDays; i++ {
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func SwanDailyDCABuy(dollarsAvialble, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64) {
	dollarsToSpendPerDay := dollarsAvialble / daysSinceStart
	// fmt.Printf("number of days to stack: %v   lenPriceData: %d\n", daysSinceStart, len(priceData))
	for _, val := range priceData {
		bitcoinAcquired += dollarsToSpendPerDay / val
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func AntiHomeMiner(fixedCosts, electricCosts, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64) {
	bitcoinAcquired += fixedCosts / priceData[0]
	cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	dollarsToSpendPerDay := electricCosts / daysSinceStart
	for _, val := range priceData {
		bitcoinAcquired += dollarsToSpendPerDay / val
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func CompareData() {
	krakenContent, err := os.ReadFile("PriceDataKraken.json")
	if err != nil {
		fmt.Printf("Error reading PriceDataKraken.json: %s\n", err.Error())
	}
	krakenVals := gjson.GetBytes(krakenContent, "data").Array()
	krakenTimestamps := []string{}
	for _, val := range krakenVals {
		timestamp := val.Get("timestamp").String()
		krakenTimestamps = append(krakenTimestamps, timestamp)
	}

	coinbaseContent, err := os.ReadFile("PriceDataCoinbase.json")
	if err != nil {
		fmt.Printf("Error reading PriceDatacoinbase.json: %s\n", err.Error())
	}
	coinbaseVals := gjson.GetBytes(coinbaseContent, "data").Array()

	for _, v := range coinbaseVals {
		timestamp := v.Get("timestamp").String()
		price := v.Get("openPrice").Float()
		foundTimestamp := false
		for _, val := range krakenTimestamps {
			if timestamp == val {
				foundTimestamp = true
			}
		}
		if !foundTimestamp {
			fmt.Printf("timestamp: %s  openPrice: %v\n", timestamp, price)
		}

	}
}

func MakeMinedSatsData(ahData []float64, minedSats float64) (minedData []float64) {
	minedData = []float64{}
	for range ahData {
		minedData = append(minedData, minedSats)
	}
	return
}

func MakePlot(ahData, swanData, antiMinerData []float64, minedSats float64) {

	minedSatsData := MakeMinedSatsData(ahData, minedSats)
	p := plot.New()
	p.Title.Text = "Sats Acquired Over Time"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Sats"
	err := plotutil.AddLinePoints(p,
		"AmericanHodl", plotData(ahData),
		"DCA", plotData(swanData),
		"AntiMiner", plotData(antiMinerData),
		"Mined", plotData(minedSatsData))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

func plotData(satsData []float64) plotter.XYs {
	pts := make(plotter.XYs, len(satsData))
	for index, sats := range satsData {
		pts[index].X = float64(index)
		pts[index].Y = sats
	}
	return pts
}

func plotMined(mined float64, days int) plotter.XYs {
	pts := make(plotter.XYs, 1)
	pts[0].X = float64(days)
	pts[0].Y = mined

	return pts
}
