package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/go-ping/ping"

	"github.com/alexflint/go-arg"
	"github.com/miekg/dns"
)

// COLORS
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()

var args struct {
	Domain     string `arg:"required,positional" help:"Base domain"`
	Wordlist   string `arg:"required, --wordlist" help:"Wordlist contain wildcards"`
	DnsServer  string `default:"1.1.1.1:53" help:"DNS server address. Format: ip:port"`
	Threads    int    `default:"100" help:"Number of threads to use"`
	MaxRetries int    `default:"10" help:"Number of max retries"`
	Outpout    string `default:"{domain}-outpout.txt" help:"Outpout to save result"`
}

type DnsBrute struct {
	domain      string
	dnsServer   string
	wordlist    string
	threads     int
	max_retries int
}

func check_domain(testDomain string, dnsServer string, maxRetries int) (string, []string, error) {
	var answer []dns.RR
	var domain = ""
	var ipAddress []string

	m := new(dns.Msg)

	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	c := dns.Client{
		Timeout: 10 * time.Second,
	}

	for i := 0; i < maxRetries; i++ {
		in, _, err := c.Exchange(m, dnsServer)
		answer = in.Answer
		// comprueba si hay errores
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				continue
			}
			return domain, ipAddress, errors.New(err.Error())
		} else {
			// obtiene la direccion ip
			if len(answer) > 0 {
				for _, val := range answer {
					// Asercion
					if address, status := val.(*dns.A); status {
						ipAddress = append(ipAddress, fmt.Sprint(address.A.String()))
						domain = testDomain
					}
				}
			} else {
				// DOMAIN NOT EXIST
				return domain, ipAddress, errors.New("NXDOMAIN")
			}
		}
		break
	}
	return domain, ipAddress, nil
}

func saveFilePrint(content []string, file string) {
	fmt.Printf("\n\n")
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		return
	}
	for i, v := range content {
		splitResult := strings.Split(v, ":")
		domain, ip, statusCode, pinger := splitResult[0], splitResult[1], splitResult[2], splitResult[3]

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		str := fmt.Sprintf("%s IP'S:[%s] StatusCode: %v Ping: %v\n", domain, ip, statusCode, pinger)
		f.WriteString(str)

		fmt.Printf("[ %v ] %v IP'S: [%v] StatusCode: %v Ping: %v\n", i, green(domain), yellow(ip), cyan(statusCode), cyan(pinger))

	}
	println()
}

func getStatusCode(domain string) (int, error) {
	req, _ := http.NewRequest("GET", "http://"+domain, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	statusCode := resp.StatusCode
	if err != nil {
		return 0, err
	} else {
		return statusCode, nil
	}
}

func checkPing(domain string) bool {
	pinger, err := ping.NewPinger(domain)
	if err != nil {
		return false
	}
	pinger.Count = 1
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return false
	}
	return true
}

func printArgs(args []string) {
	Domain, Wordlist, Threads, DnsServer, maxRetries, outpout := args[0], args[1], args[2], args[3], args[4], args[5]
	logo := figure.NewColorFigure("dnsbrute", "ogre", "red", true)
	logo.Print()

	fmt.Printf(" \n Author: %v", green("M1001-Byte\n"))
	fmt.Printf(" Domain: %s\n", yellow(Domain))
	fmt.Printf(" Wordlists: %s\n", yellow(Wordlist))
	fmt.Printf(" Threads: %s\n", yellow(Threads))
	fmt.Printf(" DnsServer: %s\n", yellow(DnsServer))
	fmt.Printf(" Maximium Retries: %v\n", yellow(maxRetries))
	fmt.Printf(" Outpout: %s\n\n", yellow(outpout))

}

func main() {

	var dnsbrute DnsBrute
	var wg sync.WaitGroup
	var domain_list []string

	var success int = 0
	var nxdomain int = 0
	var error int = 0

	var startTime = time.Now()

	var outpout string

	arg.MustParse(&args)

	dnsbrute.domain = args.Domain
	dnsbrute.wordlist = args.Wordlist
	dnsbrute.threads = args.Threads
	dnsbrute.dnsServer = args.DnsServer
	dnsbrute.max_retries = args.MaxRetries

	outpout = args.Outpout

	if strings.Contains(outpout, "domain") {
		outpout = fmt.Sprintf("%s-outpout.txt", dnsbrute.domain)
	}

	printArgs([]string{args.Domain, args.Wordlist, fmt.Sprintf("%d", args.Threads), args.DnsServer, fmt.Sprintf("%d", args.MaxRetries), outpout})

	// semaforo, para controlar las cantidad de goroutines a ejecutarse
	semaphore := make(chan struct{}, dnsbrute.threads)

	data, err := os.ReadFile(dnsbrute.wordlist)
	if err != nil {
		fmt.Printf("\n[ %s ] %s\n", red("ERROR"), err.Error())
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	total_lines := len(lines) - 1

	// last line is blank
	for index, line := range lines[0:total_lines] {
		wg.Add(1)
		semaphore <- struct{}{}
		test_domain := fmt.Sprintf("%s.%s", line, dnsbrute.domain)

		go func(domain string, dnsServer string, max_retries int) {
			defer wg.Done()

			test_domain, ipAddress, err := check_domain(domain, dnsServer, max_retries)
			elapseTime := time.Since(startTime)
			estimatedTime := elapseTime / time.Duration(index+1) * time.Duration(total_lines-index-1)

			fmt.Printf("\r Progress:  %v - %v  (%v - %v)  %v: %v %v: %v %v: %v ",
				cyan(index), cyan(len(lines)-1),
				elapseTime.Round(time.Second), estimatedTime.Round(time.Second),
				green("SUCCESS"), success, yellow("NXDOMAIN"), nxdomain, red("ERROR"), error)

			if err == nil {
				if len(test_domain) > 0 {
					success++
					domain_ip := fmt.Sprintf("%s:%s", test_domain, strings.Join(ipAddress, ","))
					domain_list = append(domain_list, domain_ip)
				}
			} else if strings.Contains("NXDOMAIN", err.Error()) {
				nxdomain++
			} else {
				error++
			}
			<-semaphore // Liberar el "token" del semáforo (indicar que la goroutine ha terminado)
		}(test_domain, dnsbrute.dnsServer, dnsbrute.max_retries)
	}
	wg.Wait()
	fmt.Printf("\n Checking http status code and ping response (icmp).")
	// get and add status code
	for index, value := range domain_list {
		fmt.Sprintf("..")
		domain_ip := strings.Split(value, ":")
		statusCode, err := getStatusCode(domain_ip[0])
		ping := checkPing(domain_ip[0])
		// HTTP STATUS CODE
		if err != nil {
			domain_list[index] += fmt.Sprintf(":%v", "ERROR")
		} else {
			domain_list[index] += fmt.Sprintf(":%d", statusCode)
		}
		// PINGER
		if ping {
			domain_list[index] += fmt.Sprintf(":%v", "OK")
		} else {
			domain_list[index] += fmt.Sprintf(":%v", "ERROR")
		}
	}
	saveFilePrint(domain_list, outpout)
	fmt.Printf("\r Total:  %v   %v: %v %v: %v %v: %v\n", cyan(total_lines), green("SUCCESS"), success, yellow("NXDOMAIN"), nxdomain, red("ERROR"), error)

}
