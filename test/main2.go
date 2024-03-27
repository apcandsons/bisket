package main

import (
	"log"
	"time"
)

type A struct {
	State string
}

func main() {
	a := &A{State: "initialized"} // Corrected initialization
	ch := make(chan string)
	go func() {
		time.Sleep(2 * time.Second)
		a.State = "complete"
		time.Sleep(5 * time.Second)
		ch <- "done"
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Println("State: " + a.State)
		}
	}()

	<-ch // Corrected receiving from channel
	log.Println("State: " + a.State)
}
