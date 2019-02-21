package main

import (
	"fmt"
	"phone_gender/device"
	"runtime"
	"sync"
)

var wg sync.WaitGroup

type devicesStruct struct {
	device *device.Device
}

func (d devicesStruct) run(ch chan string, wg *sync.WaitGroup) {
	for {
		phone := <-ch
		gender := d.device.Gender(phone)
		fmt.Printf("%s %s\n", phone, gender)
		d.device.Reset(phone)
		wg.Done()
	}
}

func main() {
	phones := []string{"18339258680", "13057518987", "133456789"}
	ch := make(chan string, len(device.Devices))
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	for _, d := range device.Devices {
		de := &devicesStruct{d}
		go de.run(ch, &wg)
	}

	for _, p := range phones {
		wg.Add(1)
		ch <- p
	}
	wg.Wait()
}
