package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	"time"
)

type Monitor struct {
	s           *ec2.EC2  // ec2 server credentials
	r           *Filter   // query arguments
	lastUpdated time.Time // time of last update

	TraceChan chan *Trace    // New pricepoints are sent on this channel
	quitChan  chan chan bool // for shutting down
	ticker    *time.Ticker
	current   *ec2.SpotPriceItem // last value that was changed the
}

func NewMonitor(auth aws.Auth, region aws.Region, filter *Filter, interval time.Duration) (*Monitor, error) {
	if interval < time.Second {
		return nil, fmt.Errorf("Monitor: Too small update interval (1 > %d)", interval)
	}

	if !filter.IsValid() {
		return nil, fmt.Errorf("Monitor: Non-specific filter %v", *filter)
	}

	m := &Monitor{
		s:        ec2.New(auth, region),
		r:        filter,
		quitChan: make(chan chan bool),
		// lock:      make(chan chan bool),
		TraceChan: make(chan *Trace),
		// horizon:   make(chan *HorizonRequest),
		ticker: time.NewTicker(interval),
		// active: true,
	}

	go m.MonitorSelect()

	return m, nil
}

type Trace struct {
	Items          []*ec2.SpotPriceItem
	From, To       time.Time
	ProcessingTime time.Duration
	err            error
}

// type HorizonRequest struct {
// 	From time.Time
// 	Resp chan *Trace
// }

func (m *Monitor) Quit() {
	fmt.Println("Sending exit signal to monitor")
	resp := make(chan bool)
	m.quitChan <- resp

	<-resp
}

func (m *Monitor) MonitorSelect() {
	for {
		select {
		case ch := <-m.quitChan:
			fmt.Println("Monitor: Exiting")
			m.ticker.Stop()
			// close(m.TraceChan)
			// close(m.quitChan)
			ch <- true
			return
		// used to lock the monitor object
		// case resp := <-m.lock:
		// 	resp <- true
		case to := <-m.ticker.C:
			// A tick was received from the ticker
			// record current time to measure processing time
			now := time.Now()
			// use first tick to initialize the
			if m.lastUpdated.IsZero() {
				m.lastUpdated = to
				continue
			}
			// init the response trace
			trace := &Trace{
				From: m.lastUpdated,
				To:   to,
			}
			// retrieve interformation
			items, err := m.retrieveInterval(m.lastUpdated, to)
			if err != nil {
				trace.err = err
			} else {
				for _, item := range items {
					if m.current == nil || (item.SpotPrice != m.current.SpotPrice &&
						!item.Timestamp.After(m.current.Timestamp)) {
						m.current = item
						trace.Items = append(trace.Items, item)
					}
				}
				// no error-changes the lastUpdated time
				m.lastUpdated = to
			}
			// note the processing time
			trace.ProcessingTime = time.Now().Sub(now)
			// return result asynchronously
			// go func() {
			m.TraceChan <- trace
			// }()
		}
	}
}

func (m *Monitor) retrieveInterval(from, to time.Time) ([]*ec2.SpotPriceItem, error) {

	if from.Equal(to) {
		return nil, fmt.Errorf("retrieveInterval: from time %v == %v to time", from, to)
	}

	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("retrieveInterval: from '%v' or to %v is zero", from, to)
	}

	if from.After(to) {
		fmt.Errorf("from-date '%v' is before to-date '%v'", from, to)
	}

	return m.getItems(from, to)
}

func (m *Monitor) getItems(from, to time.Time) ([]*ec2.SpotPriceItem, error) {
	r := &ec2.SpotPriceRequest{
		StartTime:          from,
		EndTime:            to,
		AvailabilityZone:   m.r.AvailabilityZone,
		InstanceType:       m.r.InstanceType,
		ProductDescription: m.r.ProductDescription,
	}

	// get upated list of spot price changes for that time
	// - including the basic filter
	items, err := m.s.SpotPriceHistory(r, m.r.Filter)
	if err != nil {
		return nil, err
	}

	// update the lastupdated time upon completion
	// self.lastUpdated = endTime

	return items, nil
}

// func (m *Monitor) block() chan bool {
// 	resp := make(chan bool)
// 	m.lock <- resp

// 	return resp
// }

// // Blocking
// func (m *Monitor) IsActive() bool {
// 	// lock the monitor
// 	resp := m.block()
// 	// retrieve state
// 	state := m.active
// 	// unlock monitor
// 	_ = <-resp

// 	return state
// }
