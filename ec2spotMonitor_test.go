package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	// "github.com/titanous/goamz/ec2"
	"log"
	"testing"
	"time"
)

func TestAutolUpdate(t *testing.T) {
	// get authentication stuff from ENVironment
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	// Set starttime to 3 months ago
	startTime := time.Now().AddDate(0, -3, 0)

	// Define simple filter for just one instance
	filter := NewInstanceFilter(startTime, "m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	// instantiate object with the region you want
	m := NewMonitor(auth, aws.EUWest, filter)

	// have the monitor check every second for updates
	// - that is a bit much btw.
	itemChan := m.StartPriceMonitor(1 * time.Second)

	// Only collect some 20 instances
	i := 0
	for item := range itemChan {
		i++
		fmt.Printf("[%3d] New price on channel: %v\n", i, item)
		if i > 20 {
			m.StopMonitor()
		}
	}
}

// func TestAutoUpdate(t *testing.T) {

// 	auth, err := aws.EnvAuth()
// 	if err != nil {
// 		panic(err)
// 	}
// 	m := NewMonitor(auth, aws.EUWest)

// 	startTime := time.Now().AddDate(0, -3, 0)

// 	itemChan, err := m.InitiateFilter(startTime, "m1.medium", "Linux/UNIX", "eu-west-1b", nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	m.LaunchPriceMonitorTicker(1 * time.Second)

// 	i := 0
// 	for item := range itemChan {
// 		i++
// 		// if item.AvailabilityZone == "eu-west-1b" {
// 		fmt.Printf("[%3d] New price on channel: %v\n", i, item)
// 		// }

// 		if i > 20 {
// 			break
// 		}
// 	}
// }

// func ReadItems(itemChan <-chan ec2.SpotPriceItem) {
// 	i := 0
// 	for item := range itemChan {
// 		i++
// 		if item.AvailabilityZone == "eu-west-1b" {
// 			fmt.Printf("[%3d] New price on channel: %v\n", i, item)
// 		}

// 	}
// }
