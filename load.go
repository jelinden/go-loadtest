package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/paulbellamy/ratecounter"
)

var client *http.Client
var mutex = &sync.Mutex{}
var counter = ratecounter.NewRateCounter(1 * time.Second)

const requests = 300

var listOfAddresses = []string{
	"https://www.uutispuro.fi/fi",
	"https://www.google.fi",
	"https://portfolio.jelinden.fi",
	"https://jelinden.fi",
}

func main() {
	m := make(map[string]Address)
	for _, url := range listOfAddresses {
		m[url] = Address{URL: url}
	}
	addresses := &Addresses{Addresses: m}

	client = &http.Client{
		Timeout: time.Second * 3,
	}

	c := make(chan bool, 10)
	// start up the printing of requests done per second
	go doEvery(time.Second, printReqRate)
	// run requests loop
	runRequests(addresses, &c)

	close(c)
	printResults(*addresses)
}

func runRequests(addresses *Addresses, c *chan bool) {
	for k := 0; k < requests; k++ {
		go getAddresses(addresses, c)
		time.Sleep(500 * time.Millisecond)
	}
	for k := 0; k < requests; k++ {
		<-*c
	}
}

func getAddresses(addresses *Addresses, c *chan bool) {
	for _, item := range addresses.Addresses {
		if res := get(item); res != nil {
			mutex.Lock()
			res.Requests = addresses.Addresses[item.URL].Requests + 1
			addresses.Addresses[item.URL] = *res
			mutex.Unlock()
		} else {
			mutex.Lock()
			item.Failed = addresses.Addresses[item.URL].Failed + 1
			addresses.Addresses[item.URL] = item
			mutex.Unlock()
		}
	}

	*c <- true
}

func get(address Address) *Address {
	t := time.Now()
	counter.Incr(1)
	res, err := client.Get(address.URL)
	if err != nil {
		log.Println("ERROR", err)
		return nil
	}
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("ERROR", err)
		return nil
	}
	address.Times = append(address.Times, time.Now().Sub(t).Seconds())
	return &address
}

func printReqRate() {
	log.Println("rate:", counter.Rate(), "req/s")
}

func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}
}

func printResults(addresses Addresses) {
	fmt.Println(" ")
	for _, item := range addresses.Addresses {
		log.Println(item.URL, getMinMaxAvg(item))
	}
}

func getMinMaxAvg(address Address) string {
	var min, max, avg = 0.0, 0.0, 0.0
	for _, t := range address.Times {
		if min > t || min == 0.0 {
			min = t
		}
		if max < t {
			max = t
		}
		avg += t
	}
	return fmt.Sprintf("requests: %v, failed: %v, min: %.3fs, max: %.3fs, avg: %.3fs",
		address.Requests,
		address.Failed,
		min,
		max,
		avg/float64(len(address.Times)))
}

type Addresses struct {
	Addresses map[string]Address
}
type Address struct {
	URL      string
	Times    []float64
	Requests int
	Failed   int
}
