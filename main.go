package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

func main() {
	var token, startDate string
	var kwhPrice, watts, uptimePercent, fixedCosts float64
	flag.StringVar(&token, "token", "default-token", "Specify Slush Pool token.")
	flag.Float64Var(&kwhPrice, "kwhPrice", 0.15, "Specify price paid per kilowatt hour.")
	flag.Float64Var(&watts, "watts", 3200, "Specify watts used in total.")
	flag.Float64Var(&uptimePercent, "uptimePercent", 100.0, "Specify percent uptime of your miners.")
	flag.Float64Var(&fixedCosts, "fixedCosts", 6295.55, "Specify mining setup fix costs.")
	flag.StringVar(&startDate, "startDate", "01/01/2022", "Specify start date of mining operation.")
	flag.Parse()

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
	coinsMined, err := GetUserMinedCoinsTotal(token)
	if err != nil {
		fmt.Printf("Error GetUseRMinedCoinsTotal: %s\n", err.Error())
	}
	fmt.Printf("Average coins per day: %s\n", fmt.Sprintf("%.8f", AverageCoinsPerDay(daysSinceStart, coinsMined)))
	dollarinosEarned := DollarinosEarned(coinsMined, price)
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
