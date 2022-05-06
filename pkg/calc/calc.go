package calc

import (
	"Mining-Profitability/pkg/config"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type Client struct {
	PriceDataKrakenPath   string
	PriceDataCoinbasePath string
	DataPlotFileName      string
}

func New(cfg *config.Config) *Client {
	return &Client{
		PriceDataKrakenPath:   cfg.PriceDataKrakenPath,   // "PriceDataKraken.json",
		PriceDataCoinbasePath: cfg.PriceDataCoinbasePath, // "PriceDataCoinbase.json",
		DataPlotFileName:      cfg.DataPlotFileName,      // "points.png",
	}
}

type Interface interface {
	AverageCoinsPerDay(days, coins float64) (averageCoinsPerDay float64)
	DollarinosEarned(coins, price float64) (dollarinos float64)
	ElectricCosts(kwhPrice, uptimePercentage, uptimeDays, watts float64) (electricCosts float64)
	PercentPaidOff(dollarinosEarned, fixedCosts, variableCosts float64) (percentPaidOff float64)
	DaysSinceStart(startDate string) (days float64, err error)
	DaysSinceStartUnixTimestamp(startDate string) (days float64, err error)
	BreakEvenPrice(percentPaidOff, bitcoinPrice float64) (breakevenPrice float64)
	DaysUntilBreakeven(daysSinceStart, percentPaidOff float64) (moreDays float64)
	DateFromDaysNow(days float64) (futureDate string, err error)
	AmericanHodlSlamBuy(dollarsAvailable, openPrice float64, numberDays int) (cumulativeTotal []float64, bitcoinAcquired float64)
	DailyDCABuy(dollarsAvialble, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64)
	AntiHomeMiner(fixedCosts, electricCosts, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64)
	CompareData()
	MakeMinedBitcoinData(ahData []float64, minedBitcoin float64) (minedData []float64)
}

func (c *Client) AverageCoinsPerDay(days, coins float64) (averageCoinsPerDay float64) {
	return coins / days
}

func (c *Client) DollarinosEarned(coins, price float64) (dollarinos float64) {
	return coins * price
}

func (c *Client) ElectricCosts(kwhPrice, uptimePercentage, uptimeDays, watts float64) (electricCosts float64) {
	kwhPerDay := watts * 24 / 1000
	electricCosts = kwhPrice * kwhPerDay * (uptimePercentage / 100) * uptimeDays
	return
}

func (c *Client) PercentPaidOff(dollarinosEarned, fixedCosts, variableCosts float64) (percentPaidOff float64) {
	return dollarinosEarned / (fixedCosts + variableCosts) * 100
}

func (c *Client) DaysSinceStart(startDate string) (days float64, err error) {
	t, err := time.Parse("01/02/2006", startDate)
	if err != nil {
		return
	}
	durationSinceStart := time.Since(t)
	days = durationSinceStart.Hours() / 24
	return
}

func (c *Client) DaysSinceStartUnixTimestamp(startDate string) (days float64, err error) {
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

func (c *Client) BreakEvenPrice(percentPaidOff, bitcoinPrice float64) (breakevenPrice float64) {
	return bitcoinPrice * (1 / (percentPaidOff / 100))
}

func (c *Client) DaysUntilBreakeven(daysSinceStart, percentPaidOff float64) (moreDays float64) {
	moreDays = (daysSinceStart * (1 / (percentPaidOff / 100))) - daysSinceStart
	return
}

func (c *Client) DateFromDaysNow(days float64) (futureDate string, err error) {
	hours := days * 24
	hourDuration, err := time.ParseDuration(fmt.Sprintf("%f", hours) + "h")
	if err != nil {
		return
	}
	futureTime := time.Now().Add(hourDuration)
	futureDate = futureTime.Format("01/02/2006")
	return
}

func (c *Client) AmericanHodlSlamBuy(dollarsAvailable, openPrice float64, numberDays int) (cumulativeTotal []float64, bitcoinAcquired float64) {
	bitcoinAcquired = dollarsAvailable / openPrice
	for i := 0; i < numberDays; i++ {
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func (c *Client) DailyDCABuy(dollarsAvialble, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64) {
	dollarsToSpendPerDay := dollarsAvialble / daysSinceStart
	// fmt.Printf("number of days to stack: %v   lenPriceData: %d\n", daysSinceStart, len(priceData))
	for _, val := range priceData {
		bitcoinAcquired += dollarsToSpendPerDay / val
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func (c *Client) AntiHomeMiner(fixedCosts, electricCosts, daysSinceStart float64, priceData []float64) (cumulativeTotal []float64, bitcoinAcquired float64) {
	bitcoinAcquired += fixedCosts / priceData[0]
	cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	dollarsToSpendPerDay := electricCosts / daysSinceStart
	for _, val := range priceData {
		bitcoinAcquired += dollarsToSpendPerDay / val
		cumulativeTotal = append(cumulativeTotal, bitcoinAcquired)
	}
	return
}

func (c *Client) CompareData() {
	krakenContent, err := os.ReadFile(c.PriceDataKrakenPath)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", c.PriceDataKrakenPath, err.Error())
	}
	krakenVals := gjson.GetBytes(krakenContent, "data").Array()
	krakenTimestamps := []string{}
	for _, val := range krakenVals {
		timestamp := val.Get("timestamp").String()
		krakenTimestamps = append(krakenTimestamps, timestamp)
	}

	coinbaseContent, err := os.ReadFile(c.PriceDataCoinbasePath)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", c.PriceDataCoinbasePath, err.Error())
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

func (c *Client) MakeMinedBitcoinData(ahData []float64, minedBitcoin float64) (minedData []float64) {
	minedData = []float64{}
	for range ahData {
		minedData = append(minedData, minedBitcoin)
	}
	return
}

func (c *Client) MakePlot(ahData, dcaData, antiMinerData []float64, minedBitcoin float64, hideAxis bool) {
	minedBitcoinData := c.MakeMinedBitcoinData(ahData, minedBitcoin)
	p := plot.New()
	// p.Y.Tick.Label
	p.Title.Text = "Bitcoin Acquired Over Time"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Bitcoin"
	if hideAxis {
		p.Y.Tick.Length = 0
		p.Y.Tick.Label = text.Style{
			Color:   color.White,
			Font:    font.From(plot.DefaultFont, 0),
			XAlign:  draw.XCenter,
			YAlign:  draw.YBottom,
			Handler: plot.DefaultTextHandler,
		}
	}
	err := plotutil.AddLinePoints(p,
		"AmericanHodl", c.plotData(ahData),
		"Daily DCA", c.plotData(dcaData),
		"Anti-Miner", c.plotData(antiMinerData),
		"Mined", c.plotData(minedBitcoinData))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

func (c *Client) plotData(bitcoinData []float64) plotter.XYs {
	pts := make(plotter.XYs, len(bitcoinData))
	for index, bitcoin := range bitcoinData {
		pts[index].X = float64(index)
		pts[index].Y = bitcoin
	}
	return pts
}

func (c *Client) plotMined(mined float64, days int) plotter.XYs {
	pts := make(plotter.XYs, 1)
	pts[0].X = float64(days)
	pts[0].Y = mined

	return pts
}
