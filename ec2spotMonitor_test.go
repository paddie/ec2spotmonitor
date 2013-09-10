package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	// "github.com/titanous/goamz/ec2"
	// "log"
	"runtime"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func TestAutolUpdate(t *testing.T) {
	// get authentication stuff from ENVironment
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	// Set starttime to 3 months ago
	startTime := time.Now().AddDate(0, -9, 0)

	// Define simple filter for just one instance
	filter := NewInstanceFilter(startTime, "m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	// instantiate object with the region you want
	m := NewMonitor(auth, aws.EUWest, filter)

	// have the monitor check every second for updates
	// - that is a bit much btw.
	itemChan := m.StartPriceMonitor(2 * time.Second)

	// Only collect some 20 instances

	go func() {
		time.Sleep(10 * time.Second)
		m.StopMonitor()
	}()

	i := 0
	for items := range itemChan {
		for _, item := range items {
			i++
			fmt.Printf("[%3d] New price on channel: %v\n", i, item)
		}
	}
}

func TestManualUpdate(t *testing.T) {
	// get authentication stuff from ENVironment
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	// Set starttime to 3 months ago
	startTime := time.Now().AddDate(0, -9, 0)

	// Define simple filter for just one instance
	filter := NewInstanceFilter(startTime, "m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	// instantiate object with the region you want
	m := NewMonitor(auth, aws.EUWest, filter)

	i := 0
	now := time.Now()
	for {
		endTime := startTime.AddDate(0, 0, 7)
		if endTime.After(now.AddDate(0, 0, 7)) {
			break
		}
		items, err := m.update(endTime)
		if err != nil {
			t.Error(err)
		}
		for _, item := range items {
			i++
			fmt.Printf("[%3d] New price on channel: %v\n", i, item)
		}
		startTime = endTime
	}

}
