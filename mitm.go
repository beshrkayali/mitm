// http_request_change_headers.go
package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var urlr = regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

func parse(url string) string {
	var content bytes.Buffer

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	title := document.Find("h1.graf").Text()

	content.WriteString(fmt.Sprintf("%s\n", title))
	content.WriteString(fmt.Sprintf(strings.Repeat("=", utf8.RuneCountInString(title))))
	content.WriteString(fmt.Sprintf(("\n\n")))

	document.Find(".section-content p.graf.graf--p").Each(func(i int, s *goquery.Selection) {
		content.WriteString(fmt.Sprintf("%s\n\n", s.Text()))
	})

	return content.String()
}

func mitm(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	switch {
	case urlr.MatchString(url):
		fmt.Fprintf(w, parse(url))
	}
}

func main() {
	r := mux.NewRouter()

	port := "8888"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	fmt.Printf("Running on port %s", port)

	r.HandleFunc("/", mitm).Methods("GET").Queries("url", "{.*}")

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
