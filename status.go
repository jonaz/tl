package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
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
		date, err := time.Parse(Date, c.String("date"))
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

	if !bCal.IsWorkday(date) {
		return nil
	}

	complete, _ := isDayComplete(date)

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
	// if time.Now().YearDay() == date.YearDay() && current != nil {
	if time.Now().Truncate(24*time.Hour).Equal(date.Truncate(24*time.Hour)) && current != nil {
		if current.Direction == In {
			duration += time.Now().Round(time.Minute).Sub(current.Time)
			timeLeft := time.Minute*491 - duration
			fmt.Printf("Left to work: %s (%s)\n", timeLeft, time.Now().Add(timeLeft).Format("15:04")) // 8h11m
		}
	}

	if c.Bool("compact") {
		if complete {
			fmt.Println(date.Format(Date), duration)
		} else {
			fmt.Println(date.Format(Date), duration, incompleteString(complete))
		}
		return nil
	}

	fmt.Println("Total:", duration, incompleteString(complete))
	return nil
}

func isDayComplete(date time.Time) (bool, []Direction) {
	inCnt := 0
	outCnt := 0
	for _, v := range timelog {
		if !v.Time.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			continue
		}
		if v.Direction == Out {
			outCnt++
		}
		if v.Direction == In {
			inCnt++
		}
	}

	if inCnt == 0 && outCnt == 0 && bCal.IsWorkday(date) {
		return false, []Direction{In, Out} // missing both directions
	}

	if inCnt == outCnt {
		return true, nil
	}

	resDir := []Direction{}
	if inCnt == 0 {
		resDir = append(resDir, In)
	}
	if outCnt == 0 {
		resDir = append(resDir, Out)
	}

	return false, resDir
}
func incompleteString(complete bool) string {
	if complete {
		return ""
	}
	return color.RedString("incomplete")
}
