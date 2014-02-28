package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	// "runtime"
	"testing"
	"time"
)

func TestMonitorInvalidDesc(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	s := ec2.New(auth, aws.EUWest)

	d, err := NewEC2InstanceDesc(s, "adfg", "asdf", "asd")
	if err != nil {
		t.Fatal(err)
	}

	m, err := d.StartChangeMonitor(time.Second * 1)
	if err != nil {
		t.Fatal(err)
	}

	quit := make(chan bool)
	go func() {
		for trace := range m.TraceChan {
			if trace.err == nil {
				t.Error("Monitor didn't fail on invalid arguments")
			}
			// only run once
			fmt.Println(trace.err.Error())
			quit <- true
		}
	}()
	// wait for the error to return after a tick
	<-quit
	// exit the monitor
	m.Quit()
}

func TestMonitor(t *testing.T) {

	t.SkipNow()

	m, err := desc.StartUpdateMonitor(time.Second * 1)

	// m, err := NewMonitor(auth, aws.EUWest, filter, time.Second*2)
	if err != nil {
		t.Fatal(err)
	}

	// run for 10 seconds
	quit := time.After(time.Second * 10)
	// tick := time.NewTicker(time.Second)

	go func() {
		for trace := range m.TraceChan {
			if trace.err != nil {
				t.Fatal(trace.err)
			}
			if len(trace.Items) == 0 {
				fmt.Println("No results..")
				continue
			}
			for _, item := range trace.Items {
				fmt.Printf("Price change: %#v\n", *item)
			}
		}
		fmt.Println("exiting test")
	}()
	<-quit
	m.Quit()
}
