package ec2spotmonitor

import (
	// "bytes"
	"fmt"
	"github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	// "strings"
	"log"
	"sync"
	"time"
)

type InstanceFilter struct {
	// t1.micro | m1.small | m1.medium |
	// m1.large | m1.xlarge | m3.xlarge |
	// m3.2xlarge | c1.medium | c1.xlarge |
	// m2.xlarge | m2.2xlarge | m2.4xlarge |
	// cr1.8xlarge | cc1.4xlarge |
	// cc2.8xlarge | cg1.4xlarge
	// http://goo.gl/Nk2JJ0
	// Required: NO
	StartTime    time.Time
	InstanceType string
	// Linux/UNIX | SUSE Linux | Windows |
	// Linux/UNIX (Amazon VPC) |
	// SUSE Linux (Amazon VPC) |
	// Windows (Amazon VPC)
	// Required: NO
	ProductDescription string
	// us-east-1a, etc.
	// Required: NO
	AvailabilityZone string
	Filter           *ec2.Filter
}

type Monitor struct {
	running     bool
	m           *RegionMap             // region map
	s           *ec2.EC2               // ec2 server credentials
	r           *InstanceFilter        // query arguments
	itemChan    chan ec2.SpotPriceItem // change channel
	quitChan    chan bool
	lastUpdated time.Time // time of last update
	sync.Mutex
}

const debug = true

func NewMonitor(auth aws.Auth, region aws.Region, request *InstanceFilter) *Monitor {

	monitor := &Monitor{
		m:           &RegionMap{region: make(map[string]*RegionTrace)},
		s:           ec2.New(auth, region),
		r:           request,
		lastUpdated: request.StartTime,
		quitChan:    make(chan bool),
	}

	return monitor
}

func NewInstanceFilter(from time.Time, instancetype, productdescription, availabilityzone string, filter map[string][]string) *InstanceFilter {

	fil := ec2.NewFilter()
	for k, v := range filter {
		fil.Add(k, v...)
	}

	request := &InstanceFilter{
		AvailabilityZone:   availabilityzone,
		InstanceType:       instancetype,
		ProductDescription: productdescription,
		Filter:             fil,
		StartTime:          from,
	}

	log.Printf("Registered InstanceFilter: %v", request)

	return request
}

func (self *Monitor) StopMonitor() {
	log.Println("Stopping monitor...")
	// shut down the old monitor ticker
	self.quitChan <- true
	// signal listening processes that we're shutting down
	close(self.itemChan)
	// delete reference to channel
	self.itemChan = nil
	log.Println("Monitor stopped.")
}

func (self *Monitor) StartPriceMonitor(duration time.Duration) <-chan ec2.SpotPriceItem {
	// stop monitor if one is already running
	if self.itemChan != nil {
		self.StopMonitor()
	}

	// allocate new item channel
	self.itemChan = make(chan ec2.SpotPriceItem)

	// launch goroutine that calls update every 'duration'
	// but also listens for when to shut down
	go func() {
		log.Printf("Launching SpotPriceMonitor with tick-time: %v", duration)
		tick := time.Tick(duration)
		select {
		case t := <-tick:
			err := self.update(t)
			if err != nil {
				fmt.Println(err)
			}
		case _ = <-self.quitChan:
			log.Println("Recieved Quit Signal, exiting and cleaning up..")
			// quit sending ticks
			break
		}
	}()

	self.running = true

	return self.itemChan
}

// us-east-1 -> Region based trace
type RegionMap struct {
	region map[string]*RegionTrace
	sync.Mutex
}

func (self *RegionMap) Add(items []ec2.SpotPriceItem, itemChan chan<- ec2.SpotPriceItem) {
	// for _, item := range items {
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]

		// zone: region + group: us-east-1 + a = "us-east-1a"
		zone := item.AvailabilityZone
		region := zone[0 : len(zone)-1]
		group := zone[len(zone)-1 : len(zone)]

		if _, ok := self.region[region]; !ok {
			// allocate for new instance
			self.region[region] = &RegionTrace{
				region: region,
				group:  make(map[string]*InstanceTrace),
			}
		}

		if self.region[region].Add(group, item) {
			itemChan <- item
		}
	}
}

// group = us-east-1a => a --> actual instance traces
type RegionTrace struct {
	region string                    // region name
	group  map[string]*InstanceTrace // reference for all the instances
}

func (self *RegionTrace) Add(group string, item ec2.SpotPriceItem) bool {
	if _, ok := self.group[group]; !ok {
		// allocate for new instance
		self.group[group] = &InstanceTrace{
			Group:              group,
			AvailabilityZone:   item.AvailabilityZone,
			ProductDescription: item.ProductDescription,
			InstanceType:       item.ProductDescription,
		}
	}

	return self.group[group].addPoint(ec2.PricePoint{
		DateTime: item.Timestamp,
		Price:    item.SpotPrice,
	})
}

type InstanceTrace struct {
	Group              string          //
	AvailabilityZone   string          // complete zone
	InstanceType       string          //"m3.xlarge"
	ProductDescription string          //"Linux/UNIX"
	Current            *ec2.PricePoint // the last pricepoint to change the price
	Latest             *ec2.PricePoint // the latest pricepoint (mainly for the time)
	Points             []*ec2.PricePoint
}

func (self *Monitor) Update() error {

	if self.running == false {
		return fmt.Errorf("Monitor is not running, call StartMonitor(...)")
	}

	go self.update(time.Now())

	return nil
}

func (self *Monitor) update(endTime time.Time) {

	self.Lock()
	defer self.Unlock()

	var startTime time.Time
	if self.lastUpdated.IsZero() {
		// make starttime to be 20 months ago
		// - maybe, maybe not return the complete trace of all the instances
		//   for those particular months
		if debug {
			fmt.Println("resetting startTime")
		}
		startTime = time.Now().AddDate(0, -2, 0)
	} else {
		// set starttime to be the time the last update was run
		startTime = self.lastUpdated
	}

	r := &ec2.SpotPriceRequest{
		StartTime:          startTime,
		EndTime:            endTime,
		AvailabilityZone:   self.r.AvailabilityZone,
		InstanceType:       self.r.InstanceType,
		ProductDescription: self.r.ProductDescription,
	}

	// get upated list of spot price changes for that time
	// - including the basic filter
	items, err := self.s.SpotPriceHistory(r, self.r.Filter)
	if err != nil {
		log.Panic(err)
	}

	if debug && len(items) > 0 {
		// fmt.Printf("Update: %d new pricepoints (%v -> %v)\n", len(items), startTime, endTime)
	}
	// update the lastupdated time upon completion
	self.lastUpdated = endTime

	self.m.Add(items, self.itemChan)
	// print out the result
	// - this should be replaced by a signalling
	//   to all listeneres of each instance type
	// fmt.Println(instances)
}

func (self *InstanceTrace) addPoint(point ec2.PricePoint) bool {

	// self.Latest contains the most recent pricepoint
	// associated with this configuration
	// - meaning: we update everytime a price point
	//   with a more recent date is encountered.
	if self.Latest == nil {
		self.Latest = &point
	} else {
		// if the date received is before or equal
		// to the latest date collected, ignore
		if self.Latest.DateTime.After(point.DateTime) {
			return false
		}
		self.Latest = &point
	}

	// self.current contains the current price and the date
	// when it was updated - only update when/if the price
	// changes
	if self.Current == nil {
		self.Current = &point
	} else {
		if point.Price == self.Current.Price {
			return false
		}
	}

	// The price has changed, add to list of pricepoints,
	// and update current price
	self.Current = &point

	self.Points = append(self.Points, &point)

	return true
}

// func (self *InstanceTrace) String() string {
// 	return fmt.Sprintf("\n\tLatestUpdate: %v\n\tCurrent: %v\n\tPoints: %v\n",
// 		self.Latest.DateTime, self.Current, self.Points)
// }
