package main

func main() {

}

/*
func main() {
	var slushToken, messariApiKey, startDate string
	var kwhPrice, watts, uptimePercent, fixedCosts, bitcoinMined, electricCosts float64
	var hideBitcoinOnGraph bool
	flag.StringVar(&slushToken, "slushToken", "default-token", "Specify Slush Pool token.")
	flag.Float64Var(&kwhPrice, "kwhPrice", 0.15, "Specify price paid per kilowatt hour.")
	flag.Float64Var(&watts, "watts", 3200, "Specify watts used in total.")
	flag.Float64Var(&uptimePercent, "uptimePercent", 100.0, "Specify percent uptime of your miners.")
	flag.Float64Var(&fixedCosts, "fixedCosts", 6295.55, "Specify mining setup fix costs.")
	flag.Float64Var(&bitcoinMined, "bitcoinMined", 0, "Specify total bitcoin mined (use whole bitcoin units not bitcoin).")
	flag.Float64Var(&electricCosts, "electricCosts", 0, "Specify total amount spent on electricity")
	flag.StringVar(&startDate, "startDate", "01/01/2022", "Specify start date of mining operation.")
	flag.StringVar(&messariApiKey, "messariApiKey", "default", "Specify Messari API Key")
	flag.BoolVar(&hideBitcoinOnGraph, "hideBitcoinOnGraph", false, "Will hide bitcoin on y-axis of graph, good for opsec when sharing the image. true to hide, false to keep the figure displayed")

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
	if electricCosts == 0 {
		electricCosts = ElectricCosts(kwhPrice, uptimePercent, daysSinceStart, watts)
	}
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
	dcaData, dcaBitcoin := DailyDCABuy(totalDollarsSpent, unixDaysSinceStart, priceData)
	ahData, ahBitcoin := AmericanHodlSlamBuy(totalDollarsSpent, priceData[0], len(priceData))
	fmt.Printf("AmericanHodl: %v\n", ahBitcoin)
	fmt.Printf("Daily-DCA: %v\n", dcaBitcoin)
	// MessariData(messariApiKey)
	antiHomeMinerData, antiHomeMinerBitcoin := AntiHomeMiner(fixedCosts, electricCosts, unixDaysSinceStart, priceData)
	fmt.Printf("Anti-Miner: %v\n", antiHomeMinerBitcoin)
	MakePlot(ahData, dcaData, antiHomeMinerData, bitcoinMined, hideBitcoinOnGraph)
	fmt.Printf("\n\n------------------------------------------------\n\n")
	fmt.Printf("Percentage comparison of strategies versus mining. \n\n")
	rankings := map[float64]string{ahBitcoin: "AmericanHodl",
		dcaBitcoin:           "Daily-DCA",
		antiHomeMinerBitcoin: "Anti-Miner",
	}
	CompareStrategies(bitcoinMined, rankings)

}
*/
