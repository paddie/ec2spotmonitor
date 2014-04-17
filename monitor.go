package ec2spotmonitor

import (
	"fmt"
	// "github.com/titanous/goamz/aws"
	"github.com/paddie/goamz/ec2"
	"time"
)

type Monitor struct {
	s           *ec2.EC2              // ec2 server credentials
	request     *ec2.SpotPriceRequest // base request
	filter      *ec2.Filter
	lastUpdated time.Time          // time of last update
	TraceChan   chan *Trace        // channel for pricepoints
	quitChan    chan bool          // channel to signal exit
	ticker      *time.Ticker       // ticks every 'duration'
	current     *ec2.SpotPriceItem // last value that was changed the
}

func (s *EC2InstanceDesc) newMonitor(interval time.Duration) (*Monitor, error) {

	if interval < time.Second {
		return nil, fmt.Errorf("Monitor: Too small update interval (1 > %d)", interval)
	}

	m := &Monitor{
		s:         s.ec2,
		filter:    s.filter,
		request:   s.request,
		quitChan:  make(chan bool),
		TraceChan: make(chan *Trace),
		ticker:    time.NewTicker(interval),
	}

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

func (s *EC2InstanceDesc) StartUpdateMonitor(interval time.Duration) (*Monitor, error) {

	m, err := s.newMonitor(interval)
	if err != nil {
		return nil, err
	}

	go m.UpdateMonitor()

	return m, nil
}

func (s *EC2InstanceDesc) StartChangeMonitor(interval time.Duration) (*Monitor, error) {

	m, err := s.newMonitor(interval)
	if err != nil {
		return nil, err
	}

	go m.ChangeMonitor()

	return m, nil
}

// Blocking Shutdown of the Monitor
func (m *Monitor) Quit() {
	if m == nil {
		return
	}
	m.quitChan <- true

	<-m.quitChan
}

// Only sends prices when they
func (m *Monitor) ChangeMonitor() {
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
			items, err := getSpotPriceHistory(m.s, &r, m.filter)
			// add error, even if nil
			trace.err = err

			// Update lastUpdated on a successfull
			// describehistory
			m.lastUpdated = to

			// only send item if it is newer AND different price
			// - spot prices arrive from new -> old
			//   so we reverse iterate through them
			for i := len(items) - 1; i > 0; i-- {
				item := items[i]
				if m.current == nil ||
					(item.SpotPrice != m.current.SpotPrice &&
						!item.Timestamp.After(m.current.Timestamp)) {

					m.current = item
					trace.Items = append(trace.Items, item)
				}
			}
			// note the processing time - because.. statistics
			trace.ProcessingTime = time.Now().Sub(now)
			// return result asynchronously
			m.TraceChan <- trace
		}
	}
}

// Sends price point updates every time there is an update
// - even if that price is the same as the previous price
func (m *Monitor) UpdateMonitor() {
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
			items, err := getSpotPriceHistory(m.s, &r, m.filter)
			// add error, even if nil
			trace.err = err

			// Update lastUpdated on a successfull
			// describehistory
			m.lastUpdated = to

			// only send item if it is newer AND different price
			// - spot prices arrive from new -> old
			//   so we reverse iterate through them
			for i := len(items) - 1; i > 0; i-- {
				item := items[i]
				if m.current == nil ||
					!item.Timestamp.After(m.current.Timestamp) {

					m.current = item
					trace.Items = append(trace.Items, item)
				}
			}
			// note the processing time - because.. statistics
			trace.ProcessingTime = time.Now().Sub(now)
			// return result asynchronously
			m.TraceChan <- trace
		}
	}
}
