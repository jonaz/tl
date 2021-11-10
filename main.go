package main

import (
	"fmt"
	"log"
	"os"
	"time"

	cli "github.com/urfave/cli/v2"
)

var timelog TimeLog

func main() {
	app := cli.NewApp()
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "date",
			Value: time.Now().Format("2006-01-02"),
			Usage: "Date",
		},
		&cli.StringFlag{
			Name:  "file",
			Value: "/var/log/time.log",
			Usage: "Logfile to save JSON data",
		},
		&cli.BoolFlag{
			Name:  "compact",
			Usage: "Print compact layout",
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
			Flags: flags,
			Action: func(c *cli.Context) error {
				err := stamp(c, In)
				if err != nil {
					return err
				}
				date, err := time.Parse("2006-01-02", c.String("date"))
				if err != nil {
					return err
				}

				return status(c, date)
			},
		},
		{
			Name:  "out",
			Usage: "stämpla ut",
			Flags: flags,
			Action: func(c *cli.Context) error {
				err := stamp(c, Out)
				if err != nil {
					return err
				}

				date, err := time.Parse("2006-01-02", c.String("date"))
				if err != nil {
					return err
				}

				return status(c, date)
			},
		},
		{
			Name:    "status",
			Usage:   "Print status of today or selected time range",
			Aliases: []string{"st"},
			Flags:   flags,
			Action:  statusHandler,
		},
		{
			Name:  "undo",
			Usage: "Undo last stamp",
			Flags: flags,
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

func stamp(c *cli.Context, dir Direction) error {
	filename := c.String("file")

	err := timelog.Load(filename)
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		return err
	}

	t := c.Args().Get(0)
	if t == "" {
		t = time.Now().Format("15:04")
	}

	tl, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", c.String("date"), t), loc)
	if err != nil {
		// we might have added a time in format 2021-11-09T17:13:01 lets try to parse that.
		tl, err = time.ParseInLocation(RFC3339Local, t, loc)
		if err != nil {
			return fmt.Errorf("error parsing time in local format: %w", err)
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
			return err
		}

		timelog.Add(&TimeEntry{
			Time:      tl.Add(duration),
			Direction: dir.Invert(),
		})
	}
	return timelog.Save(filename)
}
