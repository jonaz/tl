package main

import (
	"fmt"
	"strconv"
	"time"

	cli "github.com/urfave/cli/v2"
)

func statusHandler(c *cli.Context) error {
	var err error
	err = timelog.Load(c.String("file"))
	if err != nil {
		return err
	}

	days := 1
	if c.Args().Get(0) != "" {
		days, err = strconv.Atoi(c.Args().Get(0))
		if err != nil {
			return err
		}
	}

	for i := 0; i < days; i++ {
		date, err := time.Parse("2006-01-02", c.String("date"))
		if err != nil {
			return err
		}
		date = date.AddDate(0, 0, i*-1)
		err = status(c, date)
		if err != nil {
			return err
		}
	}
	return nil
}

func status(c *cli.Context, date time.Time) error {
	var duration time.Duration
	var current *TimeEntry
	for k, v := range timelog {
		if !v.Time.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			continue
		}

		if v.Direction == Out {
			prev := timelog[k-1]
			diff := v.Time.Sub(prev.Time)
			duration += diff
		}

		if !c.Bool("compact") {
			fmt.Println(v.Time, v.Direction)
		}
		current = v
	}

	// if last it in and not out we use current time to calculate how long we have worked if its on the same day as today
	//if time.Now().YearDay() == date.YearDay() && current != nil {
	if time.Now().Truncate(24*time.Hour).Equal(date.Truncate(24*time.Hour)) && current != nil {
		if current.Direction == In {
			duration += time.Now().Round(time.Minute).Sub(current.Time)
			timeLeft := time.Minute*491 - duration
			fmt.Printf("Left to work: %s (%s)\n", timeLeft, time.Now().Add(timeLeft).Format("15:04")) //8h11m
		}
	}

	if c.Bool("compact") {
		fmt.Println(date.Format("2006-01-02"), duration)
		return nil
	}

	fmt.Println("Total:", duration)
	return nil
}
