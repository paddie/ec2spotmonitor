package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/ec2"
	"time"
)

type Monitor struct {
	trace           *InstanceTrace
	s               *ec2.EC2        // ec2 server credentials
	r               *InstanceFilter // query arguments
	itemChan        chan *Package   // change channel
	lastUpdated     time.Time       // time of last update
	startTime       time.Time       // oldest price point
	quitM_chan      chan time.Time  // channel for quiting an active price monitor
	tick_chan       chan time.Time
	horizon         time.Time
	checkpoint_chan chan time.Time
	sync.Mutex      // global lock
	active          bool
	lock            chan chan bool
	current         *ec2.SpotPriceItem
	key             string
}

func NewMonitor(auth aws.Auth, region aws.Region, request *InstanceFilter) *Monitor {
	monitor := &Monitor{
		s: ec2.New(auth, region),
		r: request,
	}

	return monitor
}

func (m *Monitor) StartMonitoring(horizon time.Time, intervals time.Duration) (<-chan []ec2.SpotPriceItem, error) {

	// get data as far back as the horizon defines
	history, err := m.retrieveHorizon(horizon)
	if err != nil {
		return err
	}

	// initialize the instance trace from the first object
	if len(history) > 0 {
		item := history[0]
		m.trace = &InstanceTrace{
			Key:                item.Key(),
			AvailabilityZone:   item.AvailabilityZone,   // complete zone
			InstanceType:       item.InstanceType,       //"m3.xlarge"
			ProductDescription: item.ProductDescription, //"Linux/UNIX"
		}
	}

	for i, item := range history {
		m.trace.addPoint(item)
	}

	// we are now in an active state
	m.active = true
	// interval defined when to check for price-point updates
	m.tick_chan = time.NewTicker(interval)

	go m.MonitorSelect()

	return self.itemChan
}

type Package struct {
	Items      *ec2.SpotPriceItem
	Tick       time.Time
	Processing time.Duration
	err        error
}

func (m *Monitor) MonitorSelect() {
	for {
		select {
		case _ <- m.quitM_chan:
			fmt.Println("Monitor: Received exit signal")
			m.active = false
			break
		// used to lock the monitor object
		case resp := <-m.lock:
			resp <- true
		case t := <-m.tick_chan:
			// check for updated values
			items, err := m.retrieveInterval(m.lastUpdated, t)
			if err != nil {
				pkg.err = err
				m.itemChan <- pkg
				continue
			}
			if len(items) == 0 {
				continue
			}
			// send every SpotPriceItem with a different price than the current
			for _, item := range items {
				if m.current.SpotPrice != item.SpotPrice &&
					m.current.Timestamp.Before(item.Timestamp) {

					m.current = item
					m.itemChan <- &Package{
						Items:      items,
						Tick:       t,
						Processing: time.Now().Sub(t),
					}
				}
			}

		}
	}
	fmt.Println("Monitor: exit")
}

func (m *Monitor) retrieveHorizon(horizon time.Time) ([]*ec2.SpotPriceItem, error) {
	now := time.Now()

	if horizon.After(now) {
		return nil, fmt.Errorf("from %v is after current datetime %v", from, now)
	}

	return m.GetItems(horizon, now)
}

func (m *Monitor) retrieveInterval(from, to time.Time) ([]*ec2.SpotPriceItem, error) {

	if from.Equal(to) || from.IsZero() || to.IsZero() {
		return nil
	}

	if from.After(to) {
		fmt.Errorf("from-date '%v' is before to-date '%v'", from, to)
	}

	return m.GetItems(from, to)

}

func (m *Monitor) GetItems(from, to time.Time) ([]*ec2.SpotPriceItem, error) {
	r := &ec2.SpotPriceRequest{
		StartTime:          from,
		EndTime:            to,
		AvailabilityZone:   m.r.AvailabilityZone,
		InstanceType:       m.r.InstanceType,
		ProductDescription: m.r.ProductDescription,
	}

	// get upated list of spot price changes for that time
	// - including the basic filter
	items, err := self.s.SpotPriceHistory(r, m.r.Filter)
	if err != nil {
		return nil, err
	}

	// update the lastupdated time upon completion
	self.lastUpdated = endTime

	return items, nil
}

// func (s *Monitor) addPoints(items []ec2.SpotPriceItem) []*ec2.SpotPriceItem {

// 	updatedItems := []*ec2.SpotPriceItem{}

// 	for _, item := range items {
// 		region := item.AvailabilityZone[0 : len(item.AvailabilityZone)-1]
// 		if _, ok := s.m[region]; !ok {
// 			log.Printf("Adding region: %s", region)
// 			s.m[region] = &RegionTrace{
// 				region:        region,
// 				instanceTypes: make(map[string]*InstanceTrace),
// 			}
// 		}
// 		if s.m[region].Add(item) {
// 			updatedItems = append(updatedItems, item)
// 		}
// 	}
// 	return updatedItems
// }

// func (s *Monitor) compareUpdates(items []*ec2.SpotPriceItem) []*ec2.SpotPriceItem {

// 	updatedItems := []*ec2.SpotPriceItem{}

// 	// for _, item := range items {
// 	for i := len(items) - 1; i >= 0; i-- {
// 		item := items[i]

// 		// us-east-1a => us-east-1
// 		region := item.AvailabilityZone[0 : len(item.AvailabilityZone)-1]
// 		if _, ok := s.m[region]; !ok {
// 			log.Printf("Adding region: %s", region)
// 			s.m[region] = &RegionTrace{
// 				region:        region,
// 				instanceTypes: make(map[string]*InstanceTrace),
// 			}
// 		}
// 		if s.m[region].Add(item) {
// 			updatedItems = append(updatedItems, item)
// 		}
// 	}
// 	return updatedItems
// }

func (m *Monitor) block() chan bool {
	resp := make(chan bool)
	m.lock <- resp

	return resp
}

// Blocking
func (m *Monitor) IsActive() bool {
	// lock the monitor
	resp := m.block()
	// retrieve state
	state := m.active
	// unlock monitor
	_ <- resp

	return state
}

func (m *Monitor) SetUpdateInterval(horizon time.Time, interval time.Duration) error {

	if interval == 0 {
		// quit monitor ?
		return fmt.Errorf("SetUpdateInterval: Invalid update interval = %v", interval)
	}

	if interval < time.Second {
		return fmt.Errorf("SetUpdateInterval: Update interval too small = 1 > %d", interval)
	}

	self.tick_chan = time.NewTicker(interval)
}
