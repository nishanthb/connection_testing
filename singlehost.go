package main

import (
	"flag"
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	var hostname, port string
	var timeout, count, delay, repeat int
	flag.StringVar(&hostname, "hostname", "localhost", "Hostname to test")
	flag.StringVar(&port, "port", "8763", "Port to check")
	flag.IntVar(&timeout, "timeout", 20, "Timeout for tcp connections")
	flag.IntVar(&count, "count", 1000, "Count of concurrent connections")
	flag.IntVar(&delay, "delay", 1000, "Delay between successive runs")
	flag.IntVar(&repeat, "repeat", 100, "Number of times to repeat test")
	flag.Parse()
	wg := new(sync.WaitGroup)
	c := make(chan string, count)
	for iter := 0; iter < repeat; iter++ {
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func(hostname string, port string, timeout int, wg *sync.WaitGroup) {
				conn, err := net.DialTimeout("tcp", hostname+":"+port, time.Duration(timeout)*time.Millisecond)
				if err != nil {
					c <- fmt.Sprintf("ERROR: %s\n", err.Error())
					wg.Done()
				} else {
					defer conn.Close()
					wg.Done()
				}
			}(hostname, port, timeout, wg)
		}
		wg.Wait()
		for i := 0; i < count; i++ {
			fmt.Printf("%s", <-c)
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}
