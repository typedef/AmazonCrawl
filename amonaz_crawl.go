package main

import (
	"fmt"
	"net/http"
	"strings"
	"log"
	"io/ioutil"
	"github.com/PuerkitoBio/goquery"
	 "encoding/xml"
	 "os"
	 "runtime"
	// "strconv"
)

type ItemNode struct{
    Loc    string   `xml:"loc"`
    Lastmod   string `xml:"lastmod"`
}
type Result struct {
    XMLName xml.Name `xml:"sitemapindex"`
    ItemNode []ItemNode `xml:"sitemap"`
}

type UrlNode struct{
	Loc		string	`xml:"loc"`
	Lastmod	 string	`xml:"lastmod"`
	Changefreq string	`xml:"changefreq"`
	Priority	string	`xml:"priority"`
}
type ItemDetail struct {
	
	XMLName xml.Name `xml:"urlset"`
	UrlNodeList [] UrlNode `xml:"url"`
}

type AmazonCrawl struct {
	
	stop chan bool
//	popOdd chan string
	output chan string
	outputstop chan bool
	 f  *os.File
}

const (
	outputBufferLen = 256
	goroutineCnt	    = 8
)


func newAmazonCrawl(fileName string) *AmazonCrawl{
	ret := new(AmazonCrawl)
	ret.output = make(chan string, outputBufferLen)
	ret.outputstop = make(chan bool)
	ret.f,_ = os.Create(fileName)
	return ret
}



func (this *AmazonCrawl)ParseOutput(){
	fmt.Println("start output")
	for {
		select {
		case statu := <-this.outputstop:
			fmt.Println("outputstop!!!!!!!",statu )
			return
		default:
			for itemInfo:= range this.output{
				//fmt.Println("output:",itemInfo)
				this.f.WriteString(itemInfo)
		}
		}
	}
	fmt.Println("end ParseOutput")
}

func ParseHtmlPage(url string) (ItemInfoPrice string){

	//fmt.Println("start paser ",  url)
	var doc *goquery.Document
	 var e error
	 var ret string
	 
	if doc, e = goquery.NewDocument(url); e != nil {
	    panic(e.Error())
	  }
	bNewLine:=false
	doc.Find("div.buying .parseasinTitle #btAsinTitle").Each(func(i int, s* goquery.Selection){
		ret = s.Text()
		bNewLine = true
  	})
      // Find the review items (the type of the Selection would be *goquery.Selection)
      doc.Find("#actualPriceValue .priceLarge").Each(func(i int, s *goquery.Selection){
		  ret = ret + "\t" + s.Text()
      })
	 if bNewLine {
		 ret = ret + "\n"
	 }
	 fmt.Println(ret)
	 return ret
}

func (this *AmazonCrawl)DoAction(nodeList  [] UrlNode){
	for _,a := range nodeList{
 		ret := ParseHtmlPage(a.Loc)
 		this.output <- ret
	 }
}
 				 
func (this *AmazonCrawl)ParseSiteMap(url string){
	
	res, err := http.Get(url)
	if err != nil {
	    log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
	    log.Fatal(err)
		
	}
	  v  := Result{} 
	  err = xml.Unmarshal([]byte(data), &v)
	  if err != nil {
	      fmt.Printf("error: %v", err)
	      return
	  }
	  
	  fmt.Printf("XMLName: %#v\n", v.XMLName)
	  visitePage := 0
	  for _, node:= range v.ItemNode[1:] {
		var res *http.Response 
		var err error
		if (strings.Index(node.Loc,"reviewdetail") == -1){
	  		res, err = http.Get(node.Loc)
		  	if err != nil {
		  	    log.Fatal(err)
		  	}
		}else{
			continue
		}
		data, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err == nil {
			v := ItemDetail{}
	  	  	err = xml.Unmarshal([]byte(data), &v)
	  	  	if err != nil {
	  	    	   log.Fatal(err)
			  }
			   Len := len(v.UrlNodeList)
			  for i:=1; i< goroutineCnt;i++{
			   fmt.Println(Len*(i-1)/goroutineCnt, Len*i/goroutineCnt)
			   go this.DoAction(v.UrlNodeList[ Len * (i-1) /goroutineCnt: Len * i /goroutineCnt])
				 
		   }	
		}else{
			fmt.Println(node.Loc)
			log.Fatal(err)
		}
	  }
	  this.stop <- true
	  fmt.Println("visitePage:", visitePage)
}

func (this *AmazonCrawl)Start(url string){
	//go this.ParseHtmlPageThread()
	go this.ParseOutput()
	this.ParseSiteMap(url)
	fmt.Println("end start")
	return
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	aCrawl := newAmazonCrawl("amazon_item_price.txt")
	aCrawl.Start("http://www.amazon.cn/sitemap_feed_index1.xml")
	//aCrawl.End()
	//ParseSiteMap("http://www.amazon.cn/sitemap_feed_index1.xml", writeF)
	//ParseHtmlPage("http://www.amazon.cn/Die-Berufung-Des-Deutschen-Ordens-Gegen-Die-Preussen-Rethwisch-Conrad/dp/1168854989")
}
