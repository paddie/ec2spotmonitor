package ec2spotmonitor

import (
	"fmt"
	// "github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	"time"
)

type Monitor struct {
	s           *ec2.EC2 // ec2 server credentials
	request     *ec2.SpotPriceRequest
	lastUpdated time.Time   // time of last update
	TraceChan   chan *Trace // New pricepoints are sent on this channel
	// quitChan    chan chan bool // for shutting down
	quitChan chan bool
	ticker   *time.Ticker
	current  *ec2.SpotPriceItem // last value that was changed the
}

func (s *EC2InstanceDesc) NewMonitor(interval time.Duration) (*Monitor, error) {

	if interval < time.Second {
		return nil, fmt.Errorf("Monitor: Too small update interval (1 > %d)", interval)
	}

	m := &Monitor{
		s:        s.ec2,
		request:  s.request,
		quitChan: make(chan bool),
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

func (t *Trace) Error() string {
	if t.err == nil {
		return ""
	}
	return t.err.Error()
}

// Blocking Shutdown of the Monitor
func (m *Monitor) Quit() {
	if m == nil {
		return
	}
	m.quitChan <- true

	<-m.quitChan
}

func (m *Monitor) MonitorSelect() {
	for {
		select {
		case <-m.quitChan:
			// stop ticker
			m.ticker.Stop()
			// close trace channel
			close(m.TraceChan)
			// send signal that cleanup is complete
			m.quitChan <- true
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

			// copy the request object
			r := *m.request
			r.StartTime = m.lastUpdated
			r.EndTime = to

			// retrieve interformation
			items, err := getSpotPriceItems(m.s, &r)
			// add error, even if nil
			trace.err = err

			// Update lastUpdated on a successfull
			// describehistory
			m.lastUpdated = to

			// only send item if it is newer AND different
			for _, item := range items {
				if m.current == nil ||
					(item.SpotPrice != m.current.SpotPrice &&
						!item.Timestamp.After(m.current.Timestamp)) {

					m.current = item
					trace.Items = append(trace.Items, item)
				}
			}

			// note the processing time
			trace.ProcessingTime = time.Now().Sub(now)
			// return result asynchronously
			m.TraceChan <- trace
		}
	}
}
