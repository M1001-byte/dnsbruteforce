
# DNS BRUTE FORCE

This script performs a brute force attack on DNS servers to discover active subdomains associated with a given domain. It uses a list of keywords to generate potential subdomains quickly and efficiently, querying the DNS server to verify their existence.

![dnsbrute](https://i.imgur.com/OFRvL56.png)


## Installation
```bash
go install github.com/m1001-byte/dnsbruteforce@latest
```

    
## Usage/Examples

```bash
Usage: dnsbruteforce --wordlist WORDLIST [--dnsserver DNSSERVER] [--threads THREADS] [--maxretries MAXRETRIES] [--outpout OUTPOUT] DOMAIN

Positional arguments:
  DOMAIN                 Base domain

Options:
  --wordlist WORDLIST    Wordlist contain wildcards
  --dnsserver DNSSERVER
                         DNS server address. Format: ip:port [default: 1.1.1.1:53]
  --threads THREADS      Number of threads to use [default: 100]
  --maxretries MAXRETRIES
                         Number of max retries [default: 10]
  --outpout OUTPOUT      Outpout to save result [default: {domain}-outpout.txt]
  --help, -h             display this help and exit 
```

```bash
./dnsbrute --wordlist best-wordlist.txt --threads 100 --dnsserver 1.1.1.1:53 "google.com"
```
Example outpout file:




## License

[MIT](https://choosealicense.com/licenses/mit/)
