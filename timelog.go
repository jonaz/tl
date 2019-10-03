package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

type Direction uint8

const (
	In = Direction(iota)
	Out
)

func (s Direction) Invert() Direction {
	if s == In {
		return Out
	}
	return In
}
func (s Direction) String() string {
	name := []string{"in", "out"}
	i := uint8(s)
	return name[i]
}

type TimeEntry struct {
	Time      time.Time
	Direction Direction
}

type TimeLog []*TimeEntry

func (tl *TimeLog) RemoveLast() {
	if len(*tl) > 0 {
		*tl = (*tl)[:len(*tl)-1]
	}
}
func (tl *TimeLog) Add(t *TimeEntry) {
	*tl = append(*tl, t)
}

func (tl *TimeLog) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &tl)
	sort.Slice(timelog, func(i, j int) bool { return timelog[i].Time.Before(timelog[j].Time) })
	return err
}
func (tl *TimeLog) Save(filePath string) error {
	//f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(f)
	err = enc.Encode(tl)
	if err != nil {
		return err
	}

	return f.Close()
}
