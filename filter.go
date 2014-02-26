package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/ec2"
)

type Filter struct {
	// t1.micro | m1.small | m1.medium |
	// m1.large | m1.xlarge | m3.xlarge |
	// m3.2xlarge | c1.medium | c1.xlarge |
	// m2.xlarge | m2.2xlarge | m2.4xlarge |
	// cr1.8xlarge | cc1.4xlarge |
	// cc2.8xlarge | cg1.4xlarge
	// http://goo.gl/Nk2JJ0
	// Required: NO
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

func NewFilter(instancetype, productdescription, availabilityzone string) (*Filter, error) {

	if instancetype == "" ||
		productdescription == "" ||
		availabilityzone == "" {
		return nil, fmt.Errorf(`InstanceFilter: None of the parameters can be empty
    InstanceType:       '%s'
    ProductDescription: '%s'
    AvailabilityZone:   '%s'`, instancetype, productdescription, availabilityzone)
	}

	request := &Filter{
		AvailabilityZone:   availabilityzone,
		InstanceType:       instancetype,
		ProductDescription: productdescription,
		Filter:             nil,
	}

	return request, nil
}

func (f *Filter) IsValid() bool {
	if f.InstanceType == "" ||
		f.ProductDescription == "" ||
		f.AvailabilityZone == "" {
		return false
	}
	return true
}
