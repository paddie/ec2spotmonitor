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
		t.Fatal("succeeded in checking with invalid parameters")
	}
	fmt.Println(err)
}

func TestEmptyDesc(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}
	s := ec2.New(auth, aws.EUWest)

	_, err = NewEC2InstanceDesc(s, "", "", "")
	if err == nil {
		t.Fatal("Accepting blank arrguments")
	}
	fmt.Println(err)

}

func TestInvalidToFrom(t *testing.T) {

	now := time.Now()

	_, err := desc.GetPriceHistory(now.AddDate(0, -7, 0), now)
	if err == nil {
		t.Error("Accepting too big an horizon: -7 months")
	}
	fmt.Println(err.Error())

	_, err = desc.GetHorizon(time.Now().AddDate(0, -7, 0))
	if err == nil {
		t.Error("Accepting too big an horizon: -7 months")
	}
	fmt.Println(err.Error())
}

func TestHorizonRepeatability(t *testing.T) {

	items1, err := desc.GetHorizon(time.Now().AddDate(0, -5, 0))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)

	items2, err := desc.GetHorizon(time.Now().AddDate(0, -5, 0))
	if err != nil {
		t.Fatal(err)
	}

	if len(items1) != len(items2) {
		t.Fatal("Results are not of equal length")
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
