package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-ping/ping"

	"github.com/miekg/dns"
)

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
			if strings.Contains(err.Error(), "missing") {
				printError(err.Error())
				os.Exit(1)
			}
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
