package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	cli "github.com/urfave/cli/v2"
)

// sync from /var/log/i3lock
// FORMAT:
// echo "`date  +"%Y-%m-%dT%H:%M:%S"` LOCKED" >> /var/log/i3lock
// echo "`date  +"%Y-%m-%dT%H:%M:%S"` UNLOCKED" >> /var/log/i3lock
// 2022-11-10T08:08:05 LOCKED
// 2022-11-10T08:11:51 UNLOCKED

func syncFromLockLog(c *cli.Context) error {
	from, err := time.ParseInLocation(Date, c.String("from"), location)
	if err != nil {
		return err
	}

	to, err := time.ParseInLocation(Date, c.String("to"), location)
	if err != nil {
		return err
	}

	err = timelog.Load(c.String("file"))
	if err != nil {
		return err
	}

	lockLog, err := getFromLockLog(c.String("lock-log-file"), from, to)
	if err != nil {
		return err
	}

	changes := 0

	day := from.Truncate(time.Hour * 24)
	for {
		day = day.Add(time.Hour * 24)
		if day.After(to) {
			break
		}
		complete, missingDirections := isDayComplete(day)
		if complete {
			fmt.Println(day.Format(Date), color.GreenString("complete"))
			continue
		} else {
			fmt.Println(day.Format(Date), color.RedString("incomplete"))
		}

		locksDay := getLockLogForDay(lockLog, day)
		if len(locksDay) == 0 {
			log.Println("found no locks for day ", day)
			continue
		}
		for _, dir := range missingDirections {
			log := locksDay[0]
			if dir == Out {
				log = locksDay[len(locksDay)-1]
			}
			r, err := Ask(fmt.Sprintf("missing %s. Do your want so stamp %s (y/n)", dir, log))
			if err != nil {
				return err
			}
			fmt.Println(r)
			if r == "y" {
				timelog.Add(log)
				changes++
			}
		}
	}

	if changes > 0 {
		return timelog.Save(c.String("file"))
	}

	return nil
}

func getLockLogForDay(lockLog TimeLog, day time.Time) TimeLog {
	ret := TimeLog{}
	for _, log := range lockLog {
		if !log.Time.Truncate(24 * time.Hour).Equal(day.Truncate(24 * time.Hour)) {
			continue
		}
		ret = append(ret, log)
	}

	return ret
}
func getFromLockLog(path string, from, to time.Time) (TimeLog, error) {
	log := TimeLog{}
	file, err := os.Open(path)
	if err != nil {
		return log, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tmp := strings.Split(line, " ")
		if len(tmp) != 2 {
			continue
		}
		parsedTime, err := time.ParseInLocation(RFC3339Local, tmp[0], location)
		if err != nil {
			return log, err
		}

		if parsedTime.After(to) || parsedTime.Before(from) {
			continue
		}

		direction := In
		if tmp[1] == "LOCKED" {
			direction = Out
		}
		entry := &TimeEntry{
			Time:      parsedTime,
			Direction: direction,
		}
		log.Add(entry)
	}

	if err := scanner.Err(); err != nil {
		return log, err
	}

	return log, nil
}

// Ask asks a question and waits for use to input it on Stdin.
func Ask(q string) (string, error) {
	if q != "" {
		fmt.Printf("%s: ", q)
	}
	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(data), nil
}
