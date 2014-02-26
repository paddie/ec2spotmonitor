package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	// "github.com/titanous/goamz/ec2"
	"runtime"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func TestAutolUpdate(t *testing.T) {

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	filter, err := NewFilter("m1.medium", "Linux/UNIX", "eu-west-1b")
	if err != nil {
		t.Fatal(err)
	}

	_, itemChan, err := NewMonitor(auth, aws.EUWest, filter, time.Second*2)
	if err != nil {
		t.Fatal(err)
	}

	// run for 10 seconds
	quit := time.NewTimer(time.Second * 10)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("tick %v\n", t)
		case trace := <-itemChan:
			if trace.err != nil {
				t.Fatal(trace.err)
			}
			if len(trace.Items) == 0 {
				fmt.Println("No results..")
				continue
			}
			for _, item := range trace.Items {
				fmt.Printf("Price change: %#v", item)
			}
		case _ = <-quit.C:
			quit.Stop()
			fmt.Println("Exit signal")
			ticker.Stop()
			// m.Quit()
			time.Sleep(time.Second * 2)
			return
		}
	}
}

// func TestUpdate(t *testing.T) {
// 	auth, err := aws.EnvAuth()
// 	if err != nil {
// 		panic(err)
// 	}

// 	filter := NewInstanceFilter("m1.medium", "Linux/UNIX", "eu-west-1b", nil)

// 	m := NewMonitor(auth, aws.EUWest, filter)

// 	i := 0
// 	items := m.update(time.Now())
// 	for _, item := range items {
// 		i++
// 		fmt.Printf("[%2d] New price on channel: %v\n", i, item)
// 	}
// }

// // Test if / how the updater handles the updates
// func TestIterativeUpdate(t *testing.T) {

// 	auth, err := aws.EnvAuth()
// 	if err != nil {
// 		panic(err)
// 	}

// 	filter := NewInstanceFilter("m1.medium", "Linux/UNIX", "eu-west-1b", nil)

// 	m := NewMonitor(auth, aws.EUWest, filter)

// 	startTime := time.Now().AddDate(0, -3, 0)

// 	m.startTime = startTime

// 	endTime := startTime.AddDate(0, 0, 15)

// 	now := time.Now()
// 	i := 0
// 	flag := false
// 	for {
// 		if flag {
// 			break
// 		}
// 		if endTime.After(now) {
// 			endTime = now
// 			flag = true
// 		}

// 		items := m.update(endTime)
// 		for _, item := range items {
// 			i++
// 			fmt.Printf("[%3d] New price on channel: %v\n", i, item)
// 		}
// 		endTime = endTime.AddDate(0, 0, 15)
// 	}
// }
