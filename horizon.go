package ec2spotmonitor

import (
	"fmt"
	"github.com/paddie/goamz/ec2"
	"time"
)

const (
	MONTH = 60 * 24 * 30
)

type EC2InstanceDesc struct {
	request *ec2.SpotPriceRequest
	filter  *ec2.Filter
	ec2     *ec2.EC2
}

func (s *EC2InstanceDesc) Key() string {
	return fmt.Sprintf("%s.%s.%s", s.request.AvailabilityZone, s.request.InstanceType, s.request.ProductDescription)
}

func NewEC2InstanceDesc(s *ec2.EC2, instanceType, productDescription, availabilityZone string, filter *ec2.Filter) (*EC2InstanceDesc, error) {

	if filter == nil && (instanceType == "" ||
		productDescription == "" ||
		availabilityZone == "") {
		return nil, fmt.Errorf(`empty desciption parameters:
    InstanceType:       '%s'
    ProductDescription: '%s'
    AvailabilityZone:   '%s'`, instanceType, productDescription, availabilityZone)
	}

	request := &ec2.SpotPriceRequest{
		InstanceType:       instanceType,
		ProductDescription: productDescription,
		AvailabilityZone:   availabilityZone,
	}

	desc := &EC2InstanceDesc{
		request: request,
		ec2:     s,
		filter:  filter,
	}

	return desc, nil
}

func (s *EC2InstanceDesc) GetPriceHistory(from, to time.Time) ([]*ec2.SpotPriceItem, error) {

	// date must be non-zero
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("from-date '%v' or to-date '%v' is zero-value", from, to)
	}
	// from-date must be before to-date
	if !to.After(from) {
		return nil, fmt.Errorf("From-date '%v' is before to-date '%v'", from, to)
	}
	// if the difference is more than 3 months
	// if to.Sub(from) > time.Since(time.Now().AddDate(0, -4, 0)) {
	// 	return nil, fmt.Errorf("to and from difference exceeds 4 month limit")
	// }

	r := *s.request

	r.StartTime = from
	r.EndTime = to

	return getSpotPriceHistory(s.ec2, &r, s.filter)
}

func (s *EC2InstanceDesc) GetHorizon(from time.Time) ([]*ec2.SpotPriceItem, error) {
	to := time.Now()

	// date must be non-zero
	if from.IsZero() {
		return nil, fmt.Errorf("from-date '%v' is zero", from, to)
	}
	// from-date must be before to-date
	if !to.After(from) {
		return nil, fmt.Errorf("From-date '%v' is before to-date '%v'", from, to)
	}
	// if the difference is more than one month
	now := time.Now()
	if from.Before(now.AddDate(0, -4, 0)) {
		return nil, fmt.Errorf("from-date exceeds the 4 month limit")
	}

	r := *s.request

	r.StartTime = from
	r.EndTime = to

	return getSpotPriceHistory(s.ec2, &r, s.filter)
}

func getSpotPriceHistory(e *ec2.EC2, r *ec2.SpotPriceRequest, f *ec2.Filter) ([]*ec2.SpotPriceItem, error) {

	if r.EndTime.Sub(r.StartTime) > time.Minute*MONTH*3 {
		fmt.Println("Interval is more than 3 months, splitting up into seperate 1 month requests..")
		to, from := r.EndTime, r.StartTime
		base := *r

		appitems := []*ec2.SpotPriceItem{}
		next := to.AddDate(0, -3, 0)
		for next.After(from) {

			tmp := base
			tmp.StartTime = next
			tmp.EndTime = to

			fmt.Printf("from: %v--> to: %v\n", next, to)

			items, err := e.SpotPriceHistory(&tmp, f)
			if err != nil {
				return nil, err
			}

			for _, v := range items {
				fmt.Printf("\t%v - %.4f\n", v.Timestamp, v.SpotPrice)
			}

			appitems = append(appitems, items...)

			to = next
			next = next.AddDate(0, -3, 0)
		}

		base.StartTime = from
		base.EndTime = to

		items, err := e.SpotPriceHistory(&base, f)
		if err != nil {
			return nil, err
		}

		appitems = append(appitems, items...)

		for _, v := range appitems {
			fmt.Println(v.Timestamp)
		}

		return appitems, nil
	}

	items, err := e.SpotPriceHistory(r, f)
	if err != nil {
		return nil, err
	}
	return items, nil
}
