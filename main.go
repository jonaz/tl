package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/se"
	cli "github.com/urfave/cli/v2"
)

var timelog TimeLog

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "date",
			Value: time.Now().Format(Date),
			Usage: "Date",
		},
		&cli.StringFlag{
			Name:  "file",
			Value: "/var/log/time.log",
			Usage: "Logfile to save JSON data",
		},
		&cli.StringFlag{
			Name:  "lock-log-file",
			Value: "/var/log/i3lock",
			Usage: "Logfile of computer lock history",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "calculate",
			Aliases: []string{"c"},
			Usage:   "Calculate worked time from start lunch-duration end",
			Action:  calculate,
		},
		{
			Name:  "in",
			Usage: "stämpla in",
			Action: func(c *cli.Context) error {
				t, err := stamp(c, In)
				if err != nil {
					return err
				}

				return status(c, t)
			},
		},
		{
			Name:  "out",
			Usage: "stämpla ut",
			Action: func(c *cli.Context) error {
				t, err := stamp(c, Out)
				if err != nil {
					return err
				}

				return status(c, t)
			},
		},
		{
			Name:    "status",
			Usage:   "Print status of today or selected time range",
			Aliases: []string{"st"},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "compact",
					Aliases: []string{"c"},
					Usage:   "Print compact layout",
				},
			},
			Action: statusHandler,
		},
		{
			Name:    "synclocklog",
			Usage:   "initiate interactive sync from locklog to tl json database",
			Aliases: []string{"s"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "from",
					Aliases: []string{"f"},
					// default to last 30 days
					Value: time.Now().Add(-time.Hour * 24 * 30).Format(Date),
					Usage: "from date",
				},
				&cli.StringFlag{
					Name:    "to",
					Aliases: []string{"t"},
					Value:   time.Now().Format(Date),
					Usage:   "to date",
				},
			},
			Action: syncFromLockLog,
		},
		{
			Name:  "undo",
			Usage: "Undo last stamp",
			Action: func(c *cli.Context) error {
				err := timelog.Load(c.String("file"))
				if err != nil {
					return err
				}
				timelog.RemoveLast()
				return timelog.Save(c.String("file"))
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func calculate(c *cli.Context) error {
	start, err := time.Parse("15:04", c.Args().Get(0))
	if err != nil {
		return err
	}

	lunch, err := time.ParseDuration(c.Args().Get(1))
	if err != nil {
		return err
	}
	end, err := time.Parse("15:04", c.Args().Get(2))
	if err != nil {
		return err
	}
	difference := end.Sub(start)
	difference -= lunch
	fmt.Println("You have worked:", difference)

	return nil
}

const RFC3339Local = "2006-01-02T15:04:05"
const Date = "2006-01-02"

func stamp(c *cli.Context, dir Direction) (time.Time, error) {
	filename := c.String("file")

	err := timelog.Load(filename)
	if err != nil {
		return time.Time{}, err
	}

	t := c.Args().Get(0)
	if t == "" {
		t = time.Now().Format("15:04")
	}

	tl, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", c.String("date"), t), location)
	if err != nil {
		// we might have added a time in format 2021-11-09T17:13:01 lets try to parse that.
		tl, err = time.ParseInLocation(RFC3339Local, t, location)
		if err != nil {
			return time.Time{}, fmt.Errorf("error parsing time in local format: %w", err)
		}
	}

	te := &TimeEntry{
		Time:      tl,
		Direction: dir,
	}

	timelog.Add(te)

	// if we pass second argument (duration) also do that stamp. Usable for lunch etc. tl out 12:00 30m
	if c.Args().Get(1) != "" {
		duration, err := time.ParseDuration(c.Args().Get(1))
		if err != nil {
			return tl, err
		}

		timelog.Add(&TimeEntry{
			Time:      tl.Add(duration),
			Direction: dir.Invert(),
		})
	}
	return tl, timelog.Save(filename)
}

var location *time.Location
var bCal *cal.BusinessCalendar

func init() {
	var err error
	location, err = time.LoadLocation("Europe/Stockholm")
	if err != nil {
		log.Fatal(err)
	}
	bCal = cal.NewBusinessCalendar()
	for _, hd := range se.Holidays {
		bCal.AddHoliday(hd)
	}
}
