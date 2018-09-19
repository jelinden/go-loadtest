package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Addresses struct {
	Addresses map[string]Address
}
type Address struct {
	URL      string
	Times    []float64
	Requests int
	Failed   int
}

var client *http.Client
var mutex = &sync.Mutex{}

const requests = 300

func main() {
	m := make(map[string]Address)
	m["https://www.uutispuro.fi/fi"] = Address{URL: "https://www.uutispuro.fi/fi"}
	m["https://www.google.fi"] = Address{URL: "https://www.google.fi"}
	m["https://www.kauppalehti.fi"] = Address{URL: "https://www.kauppalehti.fi"}
	m["https://m.kauppalehti.fi"] = Address{URL: "https://m.kauppalehti.fi"}
	m["https://jelinden.fi"] = Address{URL: "https://jelinden.fi"}

	addresses := &Addresses{Addresses: m}

	client = &http.Client{
		Timeout: time.Second * 10,
	}

	c := make(chan bool, 10)
	for k := 0; k < requests; k++ {
		go getAddress(addresses, c)
	}

	for k := 0; k < requests; k++ {
		<-c
	}
	close(c)
	printResults(*addresses)
}

func getAddress(addresses *Addresses, c chan bool) {
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

	c <- true
}

func get(address Address) *Address {
	t := time.Now()
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
