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
	startTime := time.Now().AddDate(0, 0, -7)

	filter := NewInstanceFilter("m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	m := NewMonitor(auth, aws.EUWest, filter)

	itemChan := m.StartPriceMonitor(startTime, 1*time.Second)

	i := 0

	quit := time.Tick(5 * time.Second)

	for {
		select {
		case items := <-itemChan:
			for _, item := range items {
				i++
				fmt.Printf("[%2d] New price on channel: %v\n", i, item)
			}
		case _ = <-quit:
			fmt.Println("Received 'quit' signal. Exiting test")
			return
		}
	}
}

func TestUpdate(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	filter := NewInstanceFilter("m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	m := NewMonitor(auth, aws.EUWest, filter)

	i := 0
	items := m.update(time.Now())
	for _, item := range items {
		i++
		fmt.Printf("[%2d] New price on channel: %v\n", i, item)
	}
}

// Test if / how the updater handles the updates
func TestIterativeUpdate(t *testing.T) {

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	filter := NewInstanceFilter("m1.medium", "Linux/UNIX", "eu-west-1b", nil)

	m := NewMonitor(auth, aws.EUWest, filter)

	startTime := time.Now().AddDate(0, -3, 0)

	m.startTime = startTime

	endTime := startTime.AddDate(0, 0, 15)

	now := time.Now()
	i := 0
	flag := false
	for {
		if flag {
			break
		}
		if endTime.After(now) {
			endTime = now
			flag = true
		}

		items := m.update(endTime)
		for _, item := range items {
			i++
			fmt.Printf("[%3d] New price on channel: %v\n", i, item)
		}
		endTime = endTime.AddDate(0, 0, 15)
	}
}
