package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

const url = "http://192.168.100.1/cgi-bin/status_cgi"

// parseHTML breaks down the HTML table to extract the modem data.
func parseHTML(stream io.Reader) ([]map[string]string, error) {
	tokenizer := html.NewTokenizer(stream)

	// Tracking state through the doc
	tableCount := 0
	curRecord := map[string]string(nil)
	firstRow := true
	inTD := false
	countTD := 0

	// Storage for our tabular data
	var data []map[string]string

	// Storage for column header labels
	var labels []string

	labels = append(labels, "Channel")

	done := false

	for !done {
		token := tokenizer.Next()

		switch token {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				done = true
			} else {
				return nil, tokenizer.Err()
			}

		case html.TextToken:
			if tableCount == 2 && inTD {
				text := string(tokenizer.Text())

				if firstRow {
					labels = append(labels, text)
				} else {
					if curRecord == nil {
						curRecord = make(map[string]string)
					}
					curRecord[labels[countTD]] = text
					//fmt.Printf("%s %s\n", labels[countTD], text)
				}
			}

		case html.StartTagToken:
			tagName, _ := tokenizer.TagName()
			if string(tagName) == "table" {
				tableCount++
			} else if string(tagName) == "td" {
				inTD = true
			}

		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			if string(tagName) == "td" {
				inTD = false
				countTD++
			} else if string(tagName) == "tr" {
				if tableCount == 2 {
					if curRecord != nil {
						data = append(data, curRecord)
						curRecord = nil
					}
					firstRow = false
					countTD = 0
				}
			}
		}
	}

	return data, nil
}

// getData attempts to get the HTML data from the modem, then passes it
// to a parser.
func getData(url string) ([]map[string]string, error) {
	client := http.Client{
		Timeout: time.Duration(5) * time.Second,
	}

	response, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	data, err := parseHTML(response.Body)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// main runs an infinite monitoring loop.
func main() {
	for {
		data, err := getData(url)

		if err != nil {
			fmt.Println(err)
		} else {
			//fmt.Printf("%#v\n", data)
			for _, elem := range data {
				fmt.Printf("%s,%s\n", elem["Channel"], elem["SNR"])
			}
		}

		time.Sleep(time.Duration(60) * time.Second)
	}
}
