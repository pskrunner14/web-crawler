package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/llimllib/loglevel"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// MaxDepth constant that defines the
// max. depth of traversing a link to
// another webpage and so on.
const MaxDepth = 3

func main() {

	log.SetPriorityString("info")
	log.SetPrefix("crawler")

	log.Debug(os.Args)

	if len(os.Args) < 2 {
		log.Fatalf("Missing URL arg")
	}
	RecurDownloader(os.Args[1], 0)
}

// Link defines the structure of a
// link retreived from the webpage.
type Link struct {
	url   string
	text  string
	depth int
}

// String member function for the Link struct.
// returns the string representation of a Link
func (link Link) String() string {
	spacer := strings.Repeat("\t", link.depth)
	return fmt.Sprintf("%s%s (%d) - %s", spacer, link.text, link.depth, link.url)
}

// Valid member function for the Link struct.
// returns a bool true if the format of the
// link is valid and vice versa
func (link Link) Valid() bool {
	if link.depth >= MaxDepth {
		return false
	}
	if len(link.text) == 0 {
		return false
	}
	if len(link.url) == 0 || strings.Contains(strings.ToLower(link.url), "javascript") {
		return false
	}
	return true
}

// HTTPError defines the structure of an HTTP Error
// if any, retreived from the request made by the crawler.
type HTTPError struct {
	original string
}

// Error member function for the HTTPError struct.
// returns the original error string stored in the HTTPError
func (httperror HTTPError) Error() string {
	return httperror.original
}

// LinkReader reads and parses the links on a page.
// param resp the http response/ page that was downloaded
// param depth the depth of the page passed in
// returns an array of Links
func LinkReader(resp *http.Response, depth int) []Link {

	page := html.NewTokenizer(resp.Body)
	links := []Link{}

	var start *html.Token
	var text string

	for {
		_ = page.Next()
		token := page.Token()
		if token.Type == html.ErrorToken {
			break
		}
		if start != nil && token.Type == html.TextToken {
			text = fmt.Sprintf("%s%s", text, token.Data)
		}
		if token.DataAtom == atom.A {
			switch token.Type {
			case html.StartTagToken:
				if len(token.Attr) > 0 {
					start = &token
				}
			case html.EndTagToken:
				if start == nil {
					log.Warnf("Lind end found without Start:%s", text)
					continue
				}
				link := NewLink(*start, text, depth)
				if link.Valid() {
					links = append(links, link)
					log.Debugf("Link Found %v", link)
				}
				start = nil
				text = ""
			}
		}
	}
	log.Debug(links)
	return links
}

// NewLink parses a link into readable format.
// param tag the html token of the link
// param text the string received from the link
// param depth the depth of the page the link is on
// returns a parsed Link
func NewLink(tag html.Token, text string, depth int) Link {
	link := Link{text: strings.TrimSpace(text), depth: depth}
	for i := range tag.Attr {
		if tag.Attr[i].Key == "href" {
			link.url = strings.TrimSpace(tag.Attr[i].Val)
		}
	}
	return link
}

// RecurDownloader recursively downloads webpages for parsing.
// param url the URL that identifies the webpage to be crawled
// param depth the depth of the webpage it is downloading from
func RecurDownloader(url string, depth int) {
	page, err := Downloader(url)
	if err != nil {
		log.Error(err)
		return
	}
	links := LinkReader(page, depth)
	for _, link := range links {
		fmt.Println(link)
		if depth + 1 < MaxDepth {
			RecurDownloader(link.url, depth + 1)
		}
	}
}

// Downloader downloads a webpage for parsing.
// param url the URL of the webpage to be crawled
// returns the http response from the request
// returns the error if any recieved from the request
func Downloader(url string) (resp *http.Response, err error) {
	log.Debugf("Downloading %s", url)
	resp, err = http.Get(url)
	if err != nil {
		log.Debugf("Error: %s", err)
		return
	}
	if resp.StatusCode > 299 {
		err = HTTPError{fmt.Sprintf("Error (%d): %s", resp.StatusCode, url)}
		log.Debug(err)
		return
	}
	return
}
