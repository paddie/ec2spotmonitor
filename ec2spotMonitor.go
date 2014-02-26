package ec2spotmonitor

import (
// "bytes"
// "fmt"
// "github.com/titanous/goamz/aws"
// "github.com/titanous/goamz/ec2"
// // "strings"
// "log"
// "sync"
// "time"
)

// func (s *Monitor) compareUpdates(items []ec2.SpotPriceItem) []ec2.SpotPriceItem {

// 	updatedItems := []ec2.SpotPriceItem{}

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

// func (self *Monitor) Update() []ec2.SpotPriceItem {
// 	return self.update(time.Now())
// }

// func (self *Monitor) Reset(startTime time.Time) error {

// 	self.Lock()
// 	defer self.Unlock()

// 	self.startTime = startTime
// 	self.lastUpdated = time.Time{}

// 	self.m = make(map[string]*RegionTrace)

// 	return nil
// }

// func (self *Monitor) update(from, to time.Time) []ec2.SpotPriceItem {

// 	// if endtime is invalid/zero
// 	// replace by current time
// 	if endTime.IsZero() {
// 		endTime = time.Now()
// 	}

// 	var startTime time.Time
// 	// if lastupdated == zero => first time we run update
// 	if self.lastUpdated.IsZero() {
// 		// when called from test, startTime is not set
// 		if self.startTime.IsZero() {
// 			self.startTime = endTime.AddDate(0, -2, 0)
// 		}
// 		startTime = self.startTime
// 	} else {
// 		// not the first time we run update
// 		// set starttime to be the time the last update was run
// 		startTime = self.lastUpdated
// 	}

// 	r := &ec2.SpotPriceRequest{
// 		StartTime:          startTime,
// 		EndTime:            endTime,
// 		AvailabilityZone:   self.r.AvailabilityZone,
// 		InstanceType:       self.r.InstanceType,
// 		ProductDescription: self.r.ProductDescription,
// 	}

// 	// get upated list of spot price changes for that time
// 	// - including the basic filter
// 	items, err := self.s.SpotPriceHistory(r, self.r.Filter)
// 	if err != nil {
// 		log.Println(err)
// 		return []ec2.SpotPriceItem{}
// 	}

// 	// update the lastupdated time upon completion
// 	self.lastUpdated = endTime

// 	return self.compareUpdates(items)
// }
