package main

// Импортируем всё, что нам может понадобиться
import (
"fmt"
"log"
"net/http"
"io/ioutil"
"os"
	"regexp"
)

func getRoot() string {
	path := os.Getenv("ROOT_PATH")
	if path == "" {
		path = "/var/www/"
	}
	return path
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Run and listen 8383")
	http.ListenAndServe(":8383", nil)
}

func handler(iWrt http.ResponseWriter, iReq *http.Request) {
	var lGet = iReq.URL.Path[1:]
	if lGet == "" || lGet == "/" {
		lGet = "index.html"
	}

	exp := "(.*):[:digit:]*"
	r := regexp.MustCompile(exp)
	host := r.FindString(iReq.Host)
	if host != ""{
		host = host[:len(host) - 1]
	} else {
		host = iReq.Host
	}
	log.Println("Host: " + host)

	if host == ""  || host == "localhost" {
		lGet = getRoot() + lGet
	} else {
		lGet = getRoot() + host + "/" + lGet
	}

	lData := readFile(lGet)
	fmt.Fprintln(iWrt, lData)
}

func readFile(iFileName string) string {
	log.Println("readFile: " + iFileName)
	lData, err := ioutil.ReadFile(iFileName)
	var lOut string
	if !os.IsNotExist(err) {
		lOut = string(lData)
	} else {
		lOut = "404"
	}
	return lOut
}