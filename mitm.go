package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var urlr = regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

type Segment struct {
	Content template.HTML
}

type Page struct {
	// Title    string
	Segments []Segment
}

func minifmt(t string, ctx ...string) string {
	ictx := make([]interface{}, len(ctx))
	for i, v := range ctx {
		ictx[i] = v
	}

	return fmt.Sprintf(t, ictx...)
}

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

	// title := document.Find("h1.graf").Text()

	// if title == "" {
	// 	title = document.Find("h1.elevate-h1").Text()
	// }

	var segments []Segment

	document.Find(".section-content .graf.graf--p, .section-content .graf.graf--pre, .section-content .postList, .progressiveMedia-image, blockquote, h1 , h2, h3").Each(func(i int, s *goquery.Selection) {
		// c := fmt.Sprintf("<code>%s</code>", s.Text())
		c := ""
		nodename := goquery.NodeName(s)

		switch {
		case s.HasClass("graf--p"):
			c = minifmt("<p>%s</p>", s.Text())
		case nodename == "blockquote":
			c = minifmt("<blockquote>%s</blockquote>", s.Text())
		case nodename == "h1" || nodename == "h2" || nodename == "h3":
			c = minifmt("<%s>%s</%s>", nodename, s.Text(), nodename)
		case s.HasClass("graf--pre"):
			c = minifmt("<pre><code>%s</code></pre>", s.Text())
		case s.HasClass("postList"):
			c = "<p>FIXME</p>"
		case s.HasClass("progressiveMedia-image"):
			src, _ := s.Attr("data-src");
			c = minifmt("<img src='%s' />", src)
		}

		segments = append(segments, Segment{
			Content: template.HTML(c),
		})
	})

	parsed_page := Page{
		// Title:    title,
		Segments: segments,
	}

	tmpl := template.Must(template.New("page").Parse(`
<html><head><style>.content { margin: 50px 100px; line-height: 30px; } img {display: block; max-height: 300px; max-width: 600px; margin: 20px auto} blockquote { font-size: 1.5em; }</style></head>
<body><div class="content">{{range .Segments}}{{.Content}}{{end}}</div></body></html>`))

	tmpl.Execute(&content, parsed_page)

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
