package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var baseURl string
var mainUrl string
var directory string
var baseURlFilename string
var wg sync.WaitGroup
var savedFileNames = make([]string, 1)

func main() {

	//get web url and directory from command-line, using go run main.go [weburl] [directory]
	mainUrl = os.Args[1]
	directory = os.Args[2]
	if !strings.HasSuffix(directory, "\\") {
		directory = directory + "\\"
	}
	fmt.Printf("Program started !!!\n")

	//implement Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		fmt.Printf("- Ctrl+C pressed in Terminal\n")
		num := runtime.NumGoroutine()
		fmt.Printf("- %v ongoing routines stopped\n\n Program terminated", num)
		os.Exit(0)
		//defer deleteFiles()
	}()

	fmt.Printf("\nweb url: %v\n", mainUrl)

	//check if url is valid
	validURl, urlErr := url.ParseRequestURI(mainUrl)
	checkError(urlErr)

	//check if direcotry is valid
	dirErr := os.MkdirAll(directory, os.ModePerm)
	checkError(dirErr)
	fmt.Printf("directory: %v\n", directory)

	//get the domain url
	domainName := strings.Split(validURl.Path, "/")[1]
	baseURl = validURl.Scheme + "://" + validURl.Host + "/" + domainName
	fmt.Printf("domain url: %v\n\n", baseURl)

	//read and save the very fist page
	readPageHtml := readPage(mainUrl)
	baseURlFilename := urlToValidFilename(mainUrl)
	wg.Add(1)
	writeToFile(baseURlFilename+".html", readPageHtml)
	getHrefFromHtml(readPageHtml)
	wg.Wait()
	fmt.Print("Program completed !!!")

}

/*function to make filenames valid for saving*/
func urlToValidFilename(url string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(url, "/", "_"), ":", "_"), "\\", "_"), "<", "_"), ">", "_"), "*", "_"), "|", "_"), "?", "_")
}

/*function to write to file*/
func writeToFile(filename string, page string) {
	//fmt.Println(filename)
	file, err := os.Create(directory + filename)
	checkError(err)
	defer file.Close()
	_, err2 := io.WriteString(file, page)
	checkError(err2)
}

/*function to handle err*/
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

/*function to get href links from html document*/
func getHrefFromDoc(n *html.Node) {

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				if strings.HasPrefix(a.Val, baseURl) {
					subUrl := urlToValidFilename(a.Val)
					if _, err := os.Stat(directory + subUrl + ".html"); os.IsNotExist(err) {
						filename := subUrl + ".html"
						fmt.Printf("sub url: %v\nfilename: %v\nfile: %v\n\n", a.Val, filename, directory+filename)
						docPages := readPage(a.Val)
						savedFileNames = append(savedFileNames, filename)
						writeToFile(filename, docPages)
						//parallel calls
						wg.Add(1)
						go getHrefFromHtml(docPages)
					}
				}
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getHrefFromDoc(c)
	}
}

func getHrefFromHtml(requestHtml string) {
	docHtml, err := html.Parse(strings.NewReader(requestHtml))
	checkError(err)
	getHrefFromDoc(docHtml)
	defer wg.Done()

}

/*function to get the page content of url*/
func readPage(pageUrl string) string {
	resp, err := http.Get(pageUrl)
	checkError(err)
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	return string(bytes)
}
