package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	// "github.com/titanous/goamz/ec2"
	"testing"
	"time"
)

func TestAutolUpdate(t *testing.T) {

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	startTime := time.Now().AddDate(0, -3, 0)

	filter := NewInstanceFilter(startTime, "m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	m := NewMonitor(auth, aws.EUWest, filter)

	itemChan := m.StartPriceMonitor(1 * time.Second)

	i := 0
	for item := range itemChan {
		i++
		fmt.Printf("[%3d] New price on channel: %v\n", i, item)
		if i > 20 {
			break
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
