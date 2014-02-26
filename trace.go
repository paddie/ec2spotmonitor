package ec2spotmonitor

import (
	// "errors"
	"fmt"
	"github.com/titanous/goamz/ec2"
	"time"
)

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
		panic("Monitor.addPoint: <nil> item")
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
		point.Date.Before(t.Current.Date) ||
		point.Date.Equal(t.Current.Date) {
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
