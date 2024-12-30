package main

import (
	"fmt"

	"github.com/dhaavi/checkser"
)

func liveUpdates(scan *checkser.Scan, progressFunc func() string) (done func()) {
	stop := make(chan struct{})
	finalized := make(chan struct{})

	// Print initial line.
	fmt.Printf("%s", progressFunc())

	// Print updates.
	go func() {
		for {
			select {
			case <-scan.LiveUpdateSignal():
				fmt.Printf("\r%s", progressFunc())
			case <-stop:
				fmt.Printf("\r%s\n", progressFunc())
				close(finalized)
				return
			}
		}
	}()

	return func() {
		close(stop)
		<-finalized
	}
}
