package ec2spotmonitor

import (
	"fmt"
	// "github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	"time"
)

type EC2InstanceDesc struct {
	request *ec2.SpotPriceRequest
	ec2     *ec2.EC2
}

func NewEC2InstanceDesc(s *ec2.EC2, instanceType, productDescription, availabilityZone string) (*EC2InstanceDesc, error) {

	if instanceType == "" ||
		productDescription == "" ||
		availabilityZone == "" {
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
	}

	return desc, nil
}

func (s *EC2InstanceDesc) GetPriceHistory(from, to time.Time) ([]*ec2.SpotPriceItem, error) {

	// date must be non-zero
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("from-date '%v' or to-date '%v' is zero", from, to)
	}
	// from-date must be before to-date
	if !to.After(from) {
		return nil, fmt.Errorf("From-date '%v' is before to-date '%v'", from, to)
	}
	// if the difference is more than one month
	if to.Sub(from) > time.Since(time.Now().AddDate(0, -6, 0)) {
		return nil, fmt.Errorf("to and from difference exceeding 6 month limit")
	}

	r := *s.request

	r.StartTime = from
	r.EndTime = to

	return getSpotPriceItems(s.ec2, &r)
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
	if from.Before(time.Now().AddDate(0, -6, 0)) {
		return nil, fmt.Errorf("horizon more than 6 months ago")
	}

	r := *s.request

	r.StartTime = from
	r.EndTime = to

	return getSpotPriceItems(s.ec2, &r)
}

func getSpotPriceItems(ec2 *ec2.EC2, r *ec2.SpotPriceRequest) ([]*ec2.SpotPriceItem, error) {
	items, err := ec2.SpotPriceHistory(r, nil)
	if err != nil {
		return nil, err
	}
	return items, nil
}
