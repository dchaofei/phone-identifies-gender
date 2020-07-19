package main

import (
	"fmt"
	"github.com/dchaofei/phone-identifies-gender/control"
	"log"
	"sync"
)

type Run struct {
	phoneCh chan string
	control *control.Control
}

func NewRun(phoneCh chan string, control *control.Control) *Run {
	return &Run{
		phoneCh: phoneCh,
		control: control,
	}
}

func (r *Run) run(wg *sync.WaitGroup) {
	for {
		phone := <-r.phoneCh
		gender, err := r.control.Gender(phone)
		if err != nil {
			log.Printf("设备 %s %s", r.control.Serial(), err)
			return
		}
		fmt.Printf("%s %s\n", phone, gender)
		if err = r.control.Reset(); err != nil {
			log.Printf("设备 %s reset %s", r.control.Serial(), err)
			return
		}
		wg.Done()
	}
}
