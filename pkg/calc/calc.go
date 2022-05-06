package calc

import (
	"Mining-Profitability/pkg/config"
	"Mining-Profitability/pkg/externaldata"
	"Mining-Profitability/pkg/utils"
	"fmt"
	"image/color"
	"math"
	"os"
	"sort"
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
	SlushToken         *string `json:"slushToken"`
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

type ReturnPayload struct {
	BitcoinMined               float64            `json:"bitcoinMined"`
	ElectricCosts              float64            `json:"electicCosts"`
	FixedCosts                 float64            `json:"fixedCosts"`
	BitcoinPrice               float64            `json:"bitcoinPrice"`
	DaysSinceStarted           float64            `json:"daysSinceStart"`
	AverageCoinsPerDay         float64            `json:"averageCoinsPerDay"`
	DollarinosEarned           float64            `json:"dollarinosEarned"`
	PercentPaidOff             float64            `json:"percentPaidOff"`
	BreakevenPriceIncrease     float64            `json:"breakevenPriceIncrease"`
	BreakevenPrice             float64            `json:"breakevenPrice"`
	DaysUntilBreakeven         float64            `json:"daysUntilBreakeven"`
	TotalMiningDaysToBreakEven float64            `json:"totalMiningDaysToBreakEven"`
	ExpectedBreakevenDate      string             `json:"expectedBreakevenDate"`
	DailyElectricCost          float64            `json:"dailyElectricCost"`
	TotalDollarsSpent          float64            `json:"totalDollarsSpent"`
	DcaBitcoin                 float64            `json:"dcaBitcoin"`
	DcaData                    []float64          `json:"dcaData"`
	AhBitcoin                  float64            `json:"ahBitcoin"`
	AhData                     []float64          `json:"ahData"`
	AntiHomeMinerBitcoin       float64            `json:"antiHomeMinerBitcoin"`
	AntiHomeMinerData          []float64          `json:"antiHomeMinerData"`
	Rankings                   map[string]float64 `json:"rankings"`
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
	GenerateImage(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (*string, error)
	GenerateStats(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (*ReturnPayload, error)
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
	CompareStrategies(bitcoinMined float64, m map[float64]string) map[string]float64
	MakeMinedBitcoinData(ahData []float64, minedBitcoin float64) (minedData []float64)
}

func (c *Client) GenerateStats(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (*ReturnPayload, error) {
	returnPayload := &ReturnPayload{}
	price, err := externalData.GetBitcoinPrice()
	if err != nil {
		c.Logger.Error("error getting bitcoin price: %w", err)
		return nil, fmt.Errorf("error getting bitcoin price: %w", err)
	}
	(*returnPayload).BitcoinPrice = price
	(*returnPayload).DaysSinceStarted, err = c.DaysSinceStart(requestPayload.StartDate)
	if err != nil {
		c.Logger.Error("error calculating days since start: %w", err)
		return nil, fmt.Errorf("error calculating days since start: %w", err)
	}
	if requestPayload.SlushToken != nil {
		requestPayload.BitcoinMined, err = externalData.GetUserMinedCoinsTotal(*requestPayload.SlushToken)
		if err != nil {
			c.Logger.Error("Error GetUseRMinedCoinsTotal: %w\n", err)
			return nil, fmt.Errorf("error GetUseRMinedCoinsTotal: %w", err)
		}
	}
	(*returnPayload).BitcoinMined = requestPayload.BitcoinMined
	(*returnPayload).AverageCoinsPerDay = c.AverageCoinsPerDay((*returnPayload).DaysSinceStarted, returnPayload.BitcoinMined)
	(*returnPayload).DollarinosEarned = c.DollarinosEarned(returnPayload.BitcoinMined, price)

	if requestPayload.ElectricCosts == 0 {
		requestPayload.ElectricCosts = c.ElectricCosts(requestPayload.KwhPrice, requestPayload.UptimePercent, (*returnPayload).DaysSinceStarted, requestPayload.Watts)
	}
	(*returnPayload).FixedCosts = requestPayload.FixedCosts
	(*returnPayload).ElectricCosts = requestPayload.ElectricCosts
	(*returnPayload).PercentPaidOff = c.PercentPaidOff((*returnPayload).DollarinosEarned, (*returnPayload).FixedCosts, (*returnPayload).ElectricCosts)
	(*returnPayload).BreakevenPriceIncrease = ((100 / (*returnPayload).PercentPaidOff) - 1) * 100
	(*returnPayload).BreakevenPrice = c.BreakEvenPrice((*returnPayload).PercentPaidOff, price)
	(*returnPayload).DaysUntilBreakeven = c.DaysUntilBreakeven((*returnPayload).DaysSinceStarted, (*returnPayload).PercentPaidOff)
	(*returnPayload).TotalMiningDaysToBreakEven = (*returnPayload).DaysUntilBreakeven + (*returnPayload).DaysSinceStarted
	(*returnPayload).ExpectedBreakevenDate, err = c.DateFromDaysNow((*returnPayload).DaysUntilBreakeven)
	if err != nil {
		c.Logger.Error("error with DateFromDaysNow: %w", err)
		return nil, fmt.Errorf("error with DateFromDaysNow: %w", err)
	}

	(*returnPayload).DailyElectricCost = (*returnPayload).ElectricCosts / (*returnPayload).DaysSinceStarted
	unixTimeStampStart, err := utils.DateToUnixTimestamp(requestPayload.StartDate)
	if err != nil {
		c.Logger.Error("error with DateToUnixTimestamp: %w", err)
		return nil, fmt.Errorf("error with DateToUnixTimestamp: %w", err)
	}
	priceData := externalData.GetPriceDataFromDateRange(unixTimeStampStart)
	(*returnPayload).TotalDollarsSpent = requestPayload.ElectricCosts + requestPayload.FixedCosts
	unixDaysSinceStart, err := utils.RegularDateToUnix(requestPayload.StartDate)
	if err != nil {
		c.Logger.Error("error with RegularDateToUnix: %w", err)
		return nil, fmt.Errorf("error with RegularDateToUnix: %w", err)
	}

	(*returnPayload).DcaData, (*returnPayload).DcaBitcoin = c.DailyDCABuy((*returnPayload).TotalDollarsSpent, unixDaysSinceStart, priceData)
	(*returnPayload).AhData, (*returnPayload).AhBitcoin = c.AmericanHodlSlamBuy((*returnPayload).TotalDollarsSpent, priceData[0], len(priceData))
	(*returnPayload).AntiHomeMinerData, (*returnPayload).AntiHomeMinerBitcoin = c.AntiHomeMiner((*returnPayload).FixedCosts, requestPayload.ElectricCosts, unixDaysSinceStart, priceData)

	rankings := map[float64]string{
		(*returnPayload).AhBitcoin:            "AmericanHodl",
		(*returnPayload).DcaBitcoin:           "Daily-DCA",
		(*returnPayload).AntiHomeMinerBitcoin: "Anti-Miner",
	}
	(*returnPayload).Rankings = c.CompareStrategies(requestPayload.BitcoinMined, rankings)
	return returnPayload, nil
}

func (c *Client) GenerateImage(requestPayload RequestPayload, externalData externaldata.Interface, utils utils.Interface) (*string, error) {
	returnPayload, err := c.GenerateStats(requestPayload, externalData, utils)
	if err != nil {
		return nil, fmt.Errorf("error generating stats: %w", err)
	}
	return c.MakePlot(returnPayload.AhData, returnPayload.DcaData, (*returnPayload).AntiHomeMinerData, requestPayload.BitcoinMined, requestPayload.HideBitcoinOnGraph)
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

func (c *Client) CompareStrategies(bitcoinMined float64, m map[float64]string) map[string]float64 {

	results := make(map[float64]float64, len(m))
	keys := make([]float64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	// reverse the order so its in decreasing order
	for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
		keys[i], keys[j] = keys[j], keys[i]
	}

	for _, k := range keys {
		percentage := k / bitcoinMined
		switch {
		case percentage < 1:
			percentage = -(1 - percentage)
		case percentage > 1:
			percentage = percentage - 1
		}
		results[k] = percentage * 100
	}
	rankingResults := make(map[string]float64, len(m))
	for _, k := range keys {
		keyName := m[k]
		rankingResults[keyName] = results[k]
	}

	return rankingResults
}

func (c *Client) MakeMinedBitcoinData(ahData []float64, minedBitcoin float64) (minedData []float64) {
	minedData = []float64{}
	for range ahData {
		minedData = append(minedData, minedBitcoin)
	}
	return
}

func (c *Client) MakePlot(ahData, dcaData, antiMinerData []float64, minedBitcoin float64, hideAxis bool) (*string, error) {
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
		return nil, fmt.Errorf("error making plot: %w", err)
	}

	// Save the plot to a PNG file.
	fileName := fmt.Sprintf("%d-points.png", time.Now().UnixNano())

	if err := p.Save(4*vg.Inch, 4*vg.Inch, fileName); err != nil {
		return nil, fmt.Errorf("error saving plot: %w", err)
	}
	return &fileName, nil
}

func (c *Client) plotData(bitcoinData []float64) plotter.XYs {
	pts := make(plotter.XYs, len(bitcoinData))
	for index, bitcoin := range bitcoinData {
		pts[index].X = float64(index)
		pts[index].Y = bitcoin
	}
	return pts
}
