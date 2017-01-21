package main

import(
	"fmt"
	"net/http"
	"net"
	"os"
	"time"
	"sync"
	"strconv"
	"crypto/tls"
	"bufio"
	"flag"
)

func main() {
        os.Setenv("TZ","UTC")
	var timeout int
	var port int
	var url string
	var hostheader string
	var filename string
	var insecure bool
	var https bool
	var delay int
	var query string

	flag.StringVar(&url,"url","","Url to get list of hosts from")
	flag.IntVar(&delay,"delay",100,"Delay between successive iterations")
	flag.BoolVar(&insecure,"insecure",true,"Run insecure https calls")
	flag.StringVar(&query,"query","/status","Query to check on the hosts")
	flag.BoolVar(&https,"https",false,"Run http/https, true/false")
	flag.IntVar(&port,"port",80,"Port to run queries on hosts")
	flag.StringVar(&filename,"filename","list","File to get list of hosts from")
	flag.StringVar(&hostheader,"hostheader","","Host header to pass")
	flag.IntVar(&timeout,"timeout",200,"Timeout for http calls to hosts")
	flag.Parse()
	
	f,err := os.Open(filename)
	if err != nil { 
		fmt.Println("Unable to open file for reading: ",err.Error())
		os.Exit(1)
	}

	var hosts []string

	reader := bufio.NewScanner(f)
	for reader.Scan() { 
		hosts = append(hosts,reader.Text())
	}
	if url != "" {
		res,err := http.Get(url)
		defer res.Body.Close()
		if err != nil {
			fmt.Println("Errored out while getting ",url," : ", err.Error())
			if len(hosts) > 0 { 
				fmt.Println("Working with hosts from ", filename)
			} else { 
				fmt.Println("Exiting")
				os.Exit(1)
			}
		} 
		reader := bufio.NewScanner(res.Body)
		for reader.Scan() { 
			hosts = append(hosts,reader.Text())
		}
	}
	

	for {
                var wg sync.WaitGroup
                wg.Add(len(hosts))
		urlscheme := "http://"

		for _,name := range hosts { 
			go func(name string, wg *sync.WaitGroup) {
				client := http.Client{
					Timeout: time.Duration(timeout) * time.Millisecond,
				}
				if  https == true { 
					urlscheme = "https://"
					if insecure == true { 
						tr := &http.Transport{
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						}
						client = http.Client{
							Transport: tr,
							Timeout: time.Duration(timeout) * time.Millisecond,
						}
					}
				
				}
				myurl := urlscheme + name
				if port != 80 { 
					myurl = myurl + ":" +  strconv.Itoa(port)
				}
				myurl = myurl + query
				mystr := fmt.Sprint("Getting " + myurl)
				hostarr,_ := net.LookupHost(name)
				if err != nil { 
					mystr = mystr + err.Error()
					wg.Done()
					//break
				}
				mystr = mystr + fmt.Sprintf(" %v ",hostarr)
				mytime := time.Now()
				//res,err := client.Get(myurl)
				req,_ := http.NewRequest("GET",myurl,nil)
				if hostheader != "" { 
					//req.Header.Add("Host",hostheader)
					req.Host = hostheader
				}
				res,err := client.Do(req)
				if err != nil { 
					mystr = mystr + fmt.Sprint(": failed with err ", err.Error(), " at ", time.Now(), " , duration ",time.Duration(time.Since(mytime)))
					wg.Done()
				} else {
					mystr = mystr + fmt.Sprint(" Call succeeded at ",time.Now(), ", duration ",time.Duration(time.Since(mytime)))
					wg.Done()
					res.Body.Close()
					mystr = mystr + " CODE: "+strconv.Itoa(res.StatusCode)
				}
				fmt.Println(mystr)
			
			}(name,&wg)
		}
		wg.Wait()
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}
