### GODNS子域名爆破

[源代码地址](https://github.com/hatmagic/GODNS)

简介：

```

（一）godns是个自己编写的子域名爆破工具，简单实用。

（二）使用go语言的高并发编写，速度够快(天下武功唯快不破！)。

（三）编译成二进制可执行程序，支持linux，mac，windows。


```

### 出发点

最近把go语言的基础语法都学完了，就找一点大佬的github分析下他们写的工具的过程，通过分析他们写的代码，可以自己借鉴，学会他们写工具的思路。即使大部分代码我都不是很懂，但是我把他们的代码分成小的部分，一部分一部分的去分析，最后学到了很多编写思路和其中的原理。从前只能使用别人分享的工具，现在自己知道原理以后可以自己定制属于自己的工具。最主要的是通过编写自动化的漏洞验证程序，可以省很多的时间，可以用这些省下来的时间去学习更多的漏洞姿势，提高挖洞的效率，机器能干的事为什么还要去动手？


<!--more-->

### 代码分析


```
package main

//引入需要使用的包，都是golang内置的包
import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

/*
*	定义godns的结构体
*	file用来记录外部调用的字典
*	ch，wh 这两个管道是为了进程通讯，ch是读入字典的一行数据，wh是输出数据
*	wg 是通过sync.WatiGroup管理goroutine（golang的并发
*	opt 是获取用户的输入
*/
type GoDns struct {
	file string
	num  chan int
	ch   chan string
	wh   chan string
	wg   sync.WaitGroup
	opt  Option
}
/*
*	接收用户的输入
*	file用户输入的字典
*	domain用户输入的域名
*	outfile输出的文件
*	notfound是否显示所有信息
*/

type Option struct {
	file   string
	domain string
	outfile  string
	notfound bool
}


```

GetWordlist()是逐行读取字典文件的内容，通过管道传递给DnsLookUp()进行查询操作，最后关闭ch管道，防止死锁。

```

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


```

最主要的函数，从ch管道得到数据和domain进行拼接，然后查询进行查询。因为dns查询会有超时，所以每次ch读取一个数据就独自分配个goroutine让它独自运行，然后再通过wh管道写入到一个文件中。

```

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

```

把结果输出到一个文件中

```

func (this *GoDns) WriteToFile() {
	// file, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE, 0755)
	file, _ := os.OpenFile(this.opt.outfile, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	for line := range this.wh {
		file.WriteString(line)
	}
}


```

获取用户输入并保存在GoDns结构体中

```

func (this *GoDns) WriteToFile() {
	// file, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE, 0755)
	file, _ := os.OpenFile(this.opt.outfile, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	for line := range this.wh {
		file.WriteString(line)
	}
}


```

### 运行实例：

简单的输出help信息：

![](/images/godns_01.png)

以网易子域名为例：

```
godns -w one.txt -d 163.com -o hosts.txt

```
![](/images/godns_02.png)

输出的文件：(域名和ip地址之间有一个空格，方便切割)

![](/images/godns_03.png)

切割输出文件：

切割并保存成test.txt文件

```
cat hosts.txt | cut -d " " -f 2 | sort -u >test.txt

```

![](/images/godns_06.png)


用bash命令切割，另保存成文件，再通过自动化扫描工具，对得到的ip进行端口扫描或者漏洞扫描,例如通过扫描最近爆出的weblogic命令执行漏洞？如果扫到补天提交一下不是美滋滋？

![](/images/godns_05.png)


结束！能看到这里真的很感谢你，万分感谢！！！