package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type GoDns struct {
	file string
	num  chan int
	ch   chan string
	wh   chan string
	wg   sync.WaitGroup
	opt  Option
}
type Option struct {
	file   string
	domain string
	//ips      bool
	outfile  string
	notfound bool
}

func (this *GoDns) GetWordlist() {
	var lines []string
	file, err := os.Open(this.file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	this.num <- len(lines)
	for _, line := range lines {
		this.ch <- line
		//fmt.Println("test ", line)
	}
	//fmt.Println(lines)
	close(this.ch)
}

func (this *GoDns) DnsLookUp() {
	defer this.wg.Done()
	host := fmt.Sprintf("%s.%s", <-this.ch, this.opt.domain)
	ret, err := net.LookupHost(host)
	if this.opt.notfound {
		if err != nil {
			log.Printf("NotFound : %s", host)
			return
		}
	}
	if err != nil {
		return
	}
	if this.opt.outfile != "" {
		this.wh <- fmt.Sprintf("%s %s\n", host, ret[0])
	}
	fmt.Printf("Found : %s\t%s\n", host, ret[0])
}
func (this *GoDns) WriteToFile() {
	// file, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE, 0755)
	file, _ := os.OpenFile(this.opt.outfile, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	for line := range this.wh {
		file.WriteString(line)
	}
}

func getOption(o *Option) {
	flag.StringVar(&o.file, "w", "", "Your wordlist file")
	flag.StringVar(&o.domain, "d", "", "Your Domain")
	flag.StringVar(&o.outfile, "o", "", "out file")
	// flag.BoolVar(&o.ips, "ips", false, "If show ip addresss ")
	flag.BoolVar(&o.notfound, "sa", false, "Print show all result ")
	flag.Parse()
}

func banner(o Option) {
	fmt.Println("=====================================================")
	fmt.Println("Name: GODNS")
	fmt.Printf("[+] Domain:\t%s\n", o.domain)
	fmt.Printf("[+] WordList:\t%s\n", o.file)
	if o.outfile != "" {
		fmt.Printf("[+] OutPut File:\t%s\n", o.outfile)
	}
	fmt.Println("Author: www.magicz.cn")
	fmt.Println("=====================================================")
}

func main() {
	opt := Option{}
	getOption(&opt)
	banner(opt)
	godns := &GoDns{
		file: opt.file,
		ch:   make(chan string),
		num:  make(chan int),
		wh:   make(chan string),
		opt:  opt,
	}
	if godns.opt.outfile != "" {
		go godns.WriteToFile()
	}
	go godns.GetWordlist()
	gonum := <-godns.num
	godns.wg.Add(gonum)
	for i := 0; i < gonum; i++ {
		go godns.DnsLookUp()
	}
	godns.wg.Wait()
}
