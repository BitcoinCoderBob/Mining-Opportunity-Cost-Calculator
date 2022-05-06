package utils

import (
	"math"
	"strconv"
	"time"
)

type Date struct {
}

type Interface interface {
	DateToUnixTimestamp(start string) (timestamp string, err error)
	RegularDateToUnix(start string) (days float64, err error)
}

func New() *Date {
	return &Date{}
}

func (d *Date) DateToUnixTimestamp(start string) (timestamp string, err error) {
	t, err := time.Parse("01/02/2006", start)
	if err != nil {
		return
	}
	b := t.Unix()
	return strconv.FormatInt(b, 10), err
}

func (d *Date) RegularDateToUnix(start string) (days float64, err error) {
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
