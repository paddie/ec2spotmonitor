package ec2spotmonitor

import (
	"errors"
	"fmt"
	"github.com/titanous/goamz/ec2"
	"time"
)

type RegionTrace struct {
	region        string                    // region name
	instanceTypes map[string]*InstanceTrace // reference for all the instances
}

func (self *RegionTrace) Add(item *ec2.SpotPriceItem) bool {
	key := item.Key()
	if _, ok := self.instanceTypes[key]; !ok {
		// allocate for new instance
		self.instanceTypes[key] = &InstanceTrace{
			ID:                 key,
			AvailabilityZone:   item.AvailabilityZone,
			ProductDescription: item.ProductDescription,
			InstanceType:       item.InstanceType,
		}
	}

	return self.instanceTypes[key].addPoint(ec2.PricePoint{
		DateTime: item.Timestamp,
		Price:    item.SpotPrice,
	})
}

type InstanceTrace struct {
	Key                string
	AvailabilityZone   string      // complete zone
	InstanceType       string      //"m3.xlarge"
	ProductDescription string      //"Linux/UNIX"
	Current            *PricePoint // the last pricepoint to change the price
	Changes            []*PricePoint
	Max, Min           *PricePoint
}

type PricePoint struct {
	Date  time.Time
	Price float64
}

func (t *InstanceTrace) addItem(item *ec2.SpotPriceItem) (bool, error) {

	if item == nil {
		return panic("Monitor.addPoint: <nil> item")
	}

	if t.Key != item.Key() {
		return false, fmt.Errorf("Monitor.addPoint: Non-specific filter. trace.key %s != %s", t.Key, item.Key())
	}

	point := &PricePoint{
		Date:  item.Timestamp,
		Price: item.SpotPrice,
	}

	// first addition
	if t.Current == nil {
		t.Current = point
		t.Changes = append(t.Changes, point)
		return true, nil
	}

	// add point only if it is newer than the latest value
	if t.Current.Price == point.Price ||
		point.Date.Before(t.Latest.Date) ||
		point.Date.Equal(t.Latest.Date) {
		return false, nil
	}

	// update stupid stats
	if t.Max.Price < point.Price {
		t.Max = point
	} else if t.Min.Price > point.Price {
		t.Min = point
	}

	t.Current = point
	t.Changes = append(t.Changes, point)

	return true, nil
}
