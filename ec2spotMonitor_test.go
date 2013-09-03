package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	"testing"
	"time"
)

func TestUpdate(t *testing.T) {

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	m := NewMonitor(auth, aws.EUWest)

	startTime := time.Now().AddDate(0, -3, 0)

	// filter on instances in eu-west-1* = eu-west-1a, eu-west-1b, eu-west-1c..
	filter := map[string][]string{
		"availability-zone": []string{"eu-west-1*"},
	}

	itemChan, err := m.InitiateFilter(startTime, "m1.medium", "Linux/UNIX", "", filter)
	if err != nil {
		fmt.Println(err)
	}

	go ReadItems(itemChan)

	for {
		now := startTime.AddDate(0, 0, 1)
		if now.After(time.Now()) {
			break
		}
		startTime = now
		time.Sleep(50 * time.Millisecond)
		m.update(startTime)

	}
}

func ReadItems(itemChan <-chan ec2.SpotPriceItem) {
	i := 0
	for item := range itemChan {
		i++
		fmt.Printf("[%3d] New price on channel: %v\n", i, item)
	}
}
