package main

import(
	"fmt"
	"net"
	"os"
	"flag"
	"sync"
	"strconv"
	"time"
	"bufio"
	"net/http"
)

func main() {
//	os.Setenv("TZ","UTC")
	var servers []string
	var file string
	var url string
	var port int
	var timeout int
	var count int
	var delay int
	var verbose int

	var errors map[string]string

	flag.StringVar(&file,"filename","","File name")
	flag.IntVar(&port,"port",80,"Port")
	flag.IntVar(&timeout,"timeout",250,"Timeout")
	flag.IntVar(&verbose,"verbose",0,"Verbose")
	flag.IntVar(&count,"count",0,"Count")
	flag.IntVar(&delay,"delay",250,"Delay between checks")
	flag.StringVar(&url,"url","","Url")

	flag.Parse()

	if port == 0 || count < 0 || timeout == 0 { 
		fmt.Println("Invalid options entered")
		os.Exit(1)
	}
	
	if file!= "" {
		fh,err := os.Open(file)
		defer fh.Close()
		if err != nil { 
			fmt.Printf("Unable to open file for reading: %s",err.Error())
			os.Exit(1)
		}
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			servers = append(servers,scanner.Text())
		}
		if err = scanner.Err(); err != nil { 
			fmt.Println("Scanner errored out: ",err.Error())
			os.Exit(1)
		}

	}

	if url != "" { 
		res,err := http.Get(url)
		if err != nil { 
			fmt.Println("Errored out while getting url: ",err.Error())
			os.Exit(1);
		}
		scanner := bufio.NewScanner(res.Body)
		defer res.Body.Close()
		for scanner.Scan() { 
			servers = append(servers,scanner.Text())
		}
		if err = scanner.Err(); err != nil { 
			fmt.Println("Scanner errored out: ",err.Error())
			os.Exit(1)
		}
		
	}
	if len(servers) == 0 { 
		fmt.Println("No elements found in servers")
		os.Exit(1)
	} else { 
		if verbose > 2 { 
			fmt.Printf("%#v\n",servers)
		}
	}
	ctr := 0
	
	for {
		var wg sync.WaitGroup
		wg.Add(len(servers))

		ctr++
		if count == 0 {
			ctr = 0
		}
		if ctr > count {
			break
		}	
		for _,host := range servers { 
			go func(host string, wg *sync.WaitGroup) {
				var mstr string
				if verbose > 2 { 
					 mstr = fmt.Sprintf("Attempting connect to %s on port %d with timeout %d: ",host,port,timeout)
				}
				now := time.Now()
				d,err := net.DialTimeout("tcp",host + ":" + strconv.Itoa(port),time.Duration(timeout) * time.Millisecond)
				if err != nil { 
					errors[host] = err.Error() + " at " + fmt.Sprintf("%s",now)
					fmt.Println(err.Error())
					wg.Done()
				} else { 
					if verbose == 2 { 
						fmt.Printf("%s",".")
					}
					if verbose > 2 { 
						mstr = mstr + fmt.Sprintln("Connected in ",time.Since(now))
						fmt.Printf("%s",mstr)
					}
					defer d.Close()
					wg.Done()
				}
			}(host,&wg)

		}
		wg.Wait()
		if len(errors) != 0 { 
			fmt.Printf("%#v\n",errors)
		}
		errors = map[string]string{}
		if verbose > 1 { 
			fmt.Println()
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

}
