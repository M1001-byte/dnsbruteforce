package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
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
	Ping       bool   `default:"false" help:"Verifiy web server (80) and ping response (icmp)"`
	Threads    int    `default:"100" help:"Number of threads to use"`
	MaxRetries int    `default:"10" help:"Number of max retries"`
	Outpout    string `default:"{domain}-outpout.txt" help:"Outpout to save result"`
}

type DnsBrute struct {
	domain      string
	dnsServer   string
	Ping        bool
	wordlist    string
	threads     int
	max_retries int
}

func printArgs(args []string) {
	Domain, Wordlist, Threads, DnsServer, maxRetries, outpout, Ping := args[0], args[1], args[2], args[3], args[4], args[5], args[6]
	logo := figure.NewColorFigure("dnsbrute", "ogre", "red", true)
	logo.Print()

	fmt.Printf(" \n Author: %v", green("M1001-Byte\n"))
	fmt.Printf(" Domain: %s\n", yellow(Domain))
	fmt.Printf(" Wordlists: %s\n", yellow(Wordlist))
	fmt.Printf(" Threads: %s\n", yellow(Threads))
	fmt.Printf(" DnsServer: %s\n", yellow(DnsServer))
	fmt.Printf(" Ping http and icmp: %s\n", yellow(Ping))
	fmt.Printf(" Maximium Retries: %v\n", yellow(maxRetries))
	fmt.Printf(" Outpout: %s\n\n", yellow(outpout))

}

func printError(error string) {
	fmt.Printf("\n[ %s ] %s\n", red("ERROR"), error)
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
	dnsbrute.Ping = args.Ping

	outpout = args.Outpout

	if strings.Contains(outpout, "domain") {
		outpout = fmt.Sprintf("%s-outpout.txt", dnsbrute.domain)
	}

	printArgs([]string{args.Domain, args.Wordlist, fmt.Sprintf("%d", args.Threads), args.DnsServer, fmt.Sprintf("%d", args.MaxRetries), outpout, fmt.Sprintf("%v", dnsbrute.Ping)})

	// semaforo, para controlar las cantidad de goroutines a ejecutarse
	semaphore := make(chan struct{}, dnsbrute.threads)

	data, err := os.ReadFile(dnsbrute.wordlist)
	if err != nil {
		printError(err.Error())
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
	if dnsbrute.Ping {
		fmt.Printf("\n Checking http status code and ping response (icmp).")

		// get and add status code
		for index, value := range domain_list {
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
	}

	saveFilePrint(domain_list, outpout)
	fmt.Printf("\r Total:  %v   %v: %v %v: %v %v: %v\n", cyan(total_lines), green("SUCCESS"), success, yellow("NXDOMAIN"), nxdomain, red("ERROR"), error)

}
