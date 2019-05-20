package test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	// go func() {
	// 	var safe func()
	// 	safe = func() {
	// 		defer func() {
	// 			if err := recover(); err != nil {
	// 				fmt.Printf("recover")
	// 			}
	// 		}()
	// 		time.AfterFunc(time.Second, safe)
	// 		proc()
	// 	}
	// 	safe()
	// }()

	// select {}

}
func proc() {
	panic("ok")
}

func TestWaitGroup(t *testing.T) {
	wg := sync.WaitGroup{}
	c := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(num int, close <-chan struct{}) {
			defer wg.Done()
			<-close
			fmt.Println(num)
		}(i, c)
	}

	if WaitTimeout(&wg, time.Second*5) {
		close(c)
		fmt.Println("timeout exit")
	}
	time.Sleep(time.Second * 10)

}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	ch := make(chan bool)

	go time.AfterFunc(timeout, func() {
		ch <- true
	})

	go func() {
		wg.Wait()
		ch <- false
	}()

	return <-ch
}
