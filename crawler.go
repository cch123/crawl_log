package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
"time"

	"regexp"

	"github.com/gocolly/colly"
)

var f *os.File

var visited = map[string]bool{}

var fileName = "./c_log"

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// open file, read links already crawled
func init() {
	var err error
	if !pathExists(fileName) {
		f, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = f.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	links := strings.Split(string(contents), "\n")

	for _, link := range links {
		if link != "" {
			visited[link] = true
		}
	}

	f, err = os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var separater = "#######################################################"

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("www.0daydown.com"),
	)

	detailRegex, _ := regexp.Compile(`.*//www.0daydown.com/\d+/\d+.html$`)
	listRegex, _ := regexp.Compile(`.*//www.0daydown.com/category/tutorials/page/\d+`)
	panRegex, _ := regexp.Compile(`pan.baidu.com`)

	c.OnHTML("article", func(e *colly.HTMLElement) {
		// Print article
		if !panRegex.Match([]byte(e.Text)) {
			return
		}

		fmt.Println(e.Text)
		fmt.Println(separater)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "http://www.0daydown.com/09/680172.html" {
			return
		}

		if link == "http://www.0daydown.com/05/7977.html" {
			return
		}

		// 已访问过的详情页或列表页，跳过
		if visited[link] && (detailRegex.Match([]byte(link)) || listRegex.Match([]byte(link))) {
			return
		}

		// 匹配下列两种 url 模式的，才去 visit
		// http://www.0daydown.com/category/tutorials/page/6
		// http://www.0daydown.com/07/896908.html
		if !detailRegex.Match([]byte(link)) && !listRegex.Match([]byte(link)) {
			return
		}

		visited[link] = true
		f.WriteString(link + "\n")

		time.Sleep(time.Millisecond*2)
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
		r.Headers.Set("Host", "www.0daydown.com")
	})

	c.Visit("http://www.0daydown.com/category/tutorials")
}
