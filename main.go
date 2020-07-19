package main

import (
	"github.com/dchaofei/phone-identifies-gender/control"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	phoneCh := make(chan string)
	for _, c := range control.Controls {
		go NewRun(phoneCh, c).run(wg)
	}

	phones := []string{
		"18339258680",
		"133456789",
		"18339258680",
	}
	for _, p := range phones {
		wg.Add(1)
		phoneCh <- p
	}
	wg.Wait()
}
