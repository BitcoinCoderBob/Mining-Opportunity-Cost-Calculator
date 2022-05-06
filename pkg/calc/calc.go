package calc

import (
	"Mining-Profitability/pkg/config"
	"Mining-Profitability/pkg/externaldata"
	"Mining-Profitability/pkg/utils"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type RequestPayload struct {
	SlushToken         string  `json:"slushToken"`
	StartDate          string  `json:"startDate"`
	KwhPrice           float64 `json:"kwhPrice"`
	Watts              float64 `json:"watts"`
	ElectricCosts      float64 `json:"electicCosts"`
	UptimePercent      float64 `json:"updtimePercent"`
	FixedCosts         float64 `json:"fixedCosts"`
	BitcoinMined       float64 `json:"bitcoinMined"`
	MessariApiKey      string  `json:"messariApiKey"`
	HideBitcoinOnGraph bool    `json:"hideBitcoinOnGraph"`
}

type Client struct {
	PriceDataKrakenPath   string
	PriceDataCoinbasePath string
	DataPlotFileName      string
	Logger                *logrus.Logger
}

func New(cfg *config.Config, logger *logrus.Logger) *Client {
	return &Client{
		PriceDataKrakenPath:   cfg.PriceDataKrakenPath,   // "PriceDataKraken.json",
		PriceDataCoinbasePath: cfg.PriceDataCoinbasePath, // "PriceDataCoinbase.json",
		DataPlotFileName:      cfg.DataPlotFileName,      // "points.png",
		Logger:                logger,
	}
}

type Interface interface {
	Drive(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (map[float64]string, error)
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
	CompareData() error
	MakeMinedBitcoinData(ahData []float64, minedBitcoin float64) (minedData []float64)
}

func (c *Client) Drive(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (map[float64]string, error) {

	price, err := externalData.GetBitcoinPrice()
	if err != nil {
		c.Logger.Error("error getting bitcoin price: %w", err)
		return nil, fmt.Errorf("error getting bitcoin price: %w", err)
	}

	c.Logger.Info("Bicoin current price: $%s\n", fmt.Sprintf("%.2f", price))
	daysSinceStart, err := c.DaysSinceStart(requestPayload.StartDate)
	if err != nil {
		c.Logger.Error("error calculating days since start: %w", err)
		return nil, fmt.Errorf("error calculating days since start: %w", err)
	}
	c.Logger.Info("Days since start: %s\n", fmt.Sprintf("%.2f", daysSinceStart))

	if requestPayload.SlushToken != "default-token" {
		requestPayload.BitcoinMined, err = externalData.GetUserMinedCoinsTotal(requestPayload.SlushToken)
		if err != nil {
			c.Logger.Error("Error GetUseRMinedCoinsTotal: %w\n", err.Error())
			return nil, fmt.Errorf("error GetUseRMinedCoinsTotal: %w", err)
		}
	}
	c.Logger.Info("Average coins per day: %s\n", fmt.Sprintf("%.8f", c.AverageCoinsPerDay(daysSinceStart, requestPayload.BitcoinMined)))
	dollarinosEarned := c.DollarinosEarned(requestPayload.BitcoinMined, price)
	c.Logger.Info("Dollarinos earned: $%s\n", fmt.Sprintf("%.2f", dollarinosEarned))
	if requestPayload.ElectricCosts == 0 {
		requestPayload.ElectricCosts = c.ElectricCosts(requestPayload.KwhPrice, requestPayload.UptimePercent, daysSinceStart, requestPayload.Watts)
	}
	c.Logger.Info("Total electric costs: $%s\n", fmt.Sprintf("%.2f", requestPayload.ElectricCosts))
	percentPaidOff := c.PercentPaidOff(dollarinosEarned, requestPayload.FixedCosts, requestPayload.ElectricCosts)
	c.Logger.Info("Percent paid off: %s%%\n", fmt.Sprintf("%.2f", percentPaidOff))
	c.Logger.Info("Bitcoin percentage increase needed to be breakeven: %s%%\n", fmt.Sprintf("%.2f", ((100/percentPaidOff)-1)*100))
	breakevenPrice := c.BreakEvenPrice(percentPaidOff, price)
	c.Logger.Info("Breakeven price: $%s\n", fmt.Sprintf("%.2f", breakevenPrice))
	daysUntilBreakeven := c.DaysUntilBreakeven(daysSinceStart, percentPaidOff)
	c.Logger.Info("Expected more days until breakeven: %s\n", fmt.Sprintf("%.2f", daysUntilBreakeven))
	c.Logger.Info("Total mining days (past + future) to breakeven: %s\n", fmt.Sprintf("%.2f", daysUntilBreakeven+daysSinceStart))
	futureDate, err := c.DateFromDaysNow(daysUntilBreakeven)
	if err != nil {
		c.Logger.Error("error with DateFromDaysNow: %w", err)
		return nil, fmt.Errorf("error with DateFromDaysNow: %w", err)
	}
	c.Logger.Info("Expected breakeven date: %s\n", futureDate)
	c.Logger.Info("\n\n------------------------------------------------\n\n")
	dailyElectricCost := requestPayload.ElectricCosts / daysSinceStart
	c.Logger.Info("Electric costs per day: $%s\n", fmt.Sprintf("%.2f", dailyElectricCost))
	unixTimeStampStart, err := utils.DateToUnixTimestamp(requestPayload.StartDate)
	if err != nil {
		c.Logger.Error("error with DateToUnixTimestamp: %w", err)
		return nil, fmt.Errorf("error with DateToUnixTimestamp: %w", err)
	}
	priceData := externalData.GetPriceDataFromDateRange(unixTimeStampStart)
	totalDollarsSpent := requestPayload.ElectricCosts + requestPayload.FixedCosts
	unixDaysSinceStart, err := utils.RegularDateToUnix(requestPayload.StartDate)
	if err != nil {
		fmt.Printf("error with RegularDateToUnix: %s\n", err)
	}
	// fmt.Printf("a: %v\n", a)
	// daysSinceStartUnix, err := DaysSinceStartUnixTimestamp("")
	// if err != nil {
	// 	fmt.Printf("error with unix timestmap: %s\n", err.Error())
	// }
	c.Logger.Info("bitcoin mined: %v\n", requestPayload.BitcoinMined)
	dcaData, dcaBitcoin := c.DailyDCABuy(totalDollarsSpent, unixDaysSinceStart, priceData)
	ahData, ahBitcoin := c.AmericanHodlSlamBuy(totalDollarsSpent, priceData[0], len(priceData))
	c.Logger.Info("AmericanHodl: %v\n", ahBitcoin)
	c.Logger.Info("Daily-DCA: %v\n", dcaBitcoin)
	// MessariData(messariApiKey)
	antiHomeMinerData, antiHomeMinerBitcoin := c.AntiHomeMiner(requestPayload.FixedCosts, requestPayload.ElectricCosts, unixDaysSinceStart, priceData)
	c.Logger.Info("Anti-Miner: %v\n", antiHomeMinerBitcoin)
	c.MakePlot(ahData, dcaData, antiHomeMinerData, requestPayload.BitcoinMined, requestPayload.HideBitcoinOnGraph)
	c.Logger.Info("\n\n------------------------------------------------\n\n")
	c.Logger.Info("Percentage comparison of strategies versus mining. \n\n")
	rankings := map[float64]string{ahBitcoin: "AmericanHodl",
		dcaBitcoin:           "Daily-DCA",
		antiHomeMinerBitcoin: "Anti-Miner",
	}

	return rankings, nil
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
	c.Logger.Info("number of days to stack: %v   lenPriceData: %d\n", daysSinceStart, len(priceData))
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

func (c *Client) CompareData() error {
	krakenContent, err := os.ReadFile(c.PriceDataKrakenPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", c.PriceDataKrakenPath, err)
	}
	krakenVals := gjson.GetBytes(krakenContent, "data").Array()
	krakenTimestamps := []string{}
	for _, val := range krakenVals {
		timestamp := val.Get("timestamp").String()
		krakenTimestamps = append(krakenTimestamps, timestamp)
	}

	coinbaseContent, err := os.ReadFile(c.PriceDataCoinbasePath)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", c.PriceDataCoinbasePath, err)
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
			c.Logger.Info("timestamp: %s  openPrice: %v\n", timestamp, price)
		}

	}
	return nil
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
