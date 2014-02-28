package ec2spotmonitor

import (
	"fmt"
	"github.com/titanous/goamz/aws"
	"github.com/titanous/goamz/ec2"
	"runtime"
	"testing"
	"time"
)

var desc *EC2InstanceDesc

func init() {
	runtime.GOMAXPROCS(4)

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	s := ec2.New(auth, aws.EUWest)

	desc, err = NewEC2InstanceDesc(s,
		"m1.medium",
		"Linux/UNIX",
		"eu-west-1b")
	if err != nil {
		panic(err)
	}
}

// Make sure that we fail gracefully
// on invalid arguments
func TestInvalidDesc(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	s := ec2.New(auth, aws.EUWest)

	d, err := NewEC2InstanceDesc(s, "adfg", "asdf", "asd")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.GetHorizon(time.Now().AddDate(0, 0, -7))
	if err == nil {
		t.Fatal("invalid parameters acepted")
	}
}

// Make sure that no empty arguments are accepted.
// TODO: make sure that the parameters
// do not contains wild-cards '*'
func TestEmptyDesc(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	s := ec2.New(auth, aws.EUWest)

	_, err = NewEC2InstanceDesc(s, "", "", "")
	if err == nil {
		t.Fatal("Blank arguments accepted")
	}
}

// Make sure that we do not accept dates
// that are more than 6 months apart.
func TestInvalidToFrom(t *testing.T) {

	now := time.Now()

	_, err := desc.GetPriceHistory(now.AddDate(0, -7, 0), now)
	if err == nil {
		t.Error("Accepted to-from > 6 months")
	}
	fmt.Println(err.Error())

	_, err = desc.GetHorizon(time.Now().AddDate(0, -7, 0))
	if err == nil {
		t.Error("Accepted horizon > 6 months")
	}
}

// Make sure that no values are filtered out
// if we set different horizons
// - from start to finish, they should be equal
func TestUnevenInterval(t *testing.T) {

	items1, err := desc.GetHorizon(time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatal(err)
	}

	items2, err := desc.GetHorizon(time.Now().AddDate(0, -2, 0))
	if err != nil {
		t.Fatal(err)
	}

	// run backwards through the items
	for i := 0; i < len(items1); i++ {

		i1 := items1[i]
		i2 := items2[i]

		if i1.SpotPrice != i2.SpotPrice ||
			!i1.Timestamp.Equal(i2.Timestamp) {
			t.Errorf("%v: i1 %f != %f i2 %v",
				i1.Timestamp, i1.SpotPrice, i2.SpotPrice, i2.Timestamp)
		}
	}
}

// Make sure that two consequtive calls
// produce equal results
func TestHorizonRepeatability(t *testing.T) {

	items1, err := desc.GetHorizon(time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatal(err)
	}

	items2, err := desc.GetHorizon(time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(items1); i++ {

		i1, i2 := items1[i], items2[i]

		if i1.SpotPrice != i2.SpotPrice ||
			!i1.Timestamp.Equal(i2.Timestamp) {
			t.Errorf("%v: i1 %f != %f i2",
				i1.Timestamp, i1.SpotPrice, i2.SpotPrice)
		}
	}
}
