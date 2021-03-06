package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

type entry struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
}

type QueryResult struct {
	Title   string
	Entries []*entry
}

type JsonFormat struct {
	Entries []*entry `json:"entries"`
	Error   string   `json:"error"`
}

func main() {
	app := cli.NewApp()
	app.Name = "hngo"
	app.Usage = "Command Line Client for Hacker News (https://news.ycombinator.com/)"
	app.ArgsUsage = "[category]"
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "number, n",
			Value: 10,
			Usage: "number of output lines",
		},
		cli.BoolFlag{
			Name:  "json",
			Usage: "output as json",
		},
	}

	app.Action = func(c *cli.Context) {
		category := c.Args().First()
		number := c.Int("number")

		url := buildUrl(category)
		result, err := crawl(url, number)
		if c.Bool("json") {
			showResultAsJson(result, err)
		} else {
			showResult(result, url)
		}
	}

	app.Run(os.Args)
}

func buildUrl(keyword string) string {
	return fmt.Sprintf("https://news.ycombinator.com/%s", url.QueryEscape(keyword))
}

func crawl(url string, number int) (QueryResult, error) {
	entries := []*entry{}
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return QueryResult{
			Title:   "",
			Entries: entries,
		}, err
	}
	doc.Find("table.itemlist tr").Each(func(_ int, tr *goquery.Selection) {
		cells := [3]interface{}{"", "", 0}
		tr.Find("a.storylink").Each(func(_ int, s *goquery.Selection) {
			cells[0] = s.Text()
			href, _ := s.Attr("href")
			if href != "" {
				cells[1] = href
			}
		})

		entry := entry{
			Title:         strings.TrimSpace(cells[0].(string)),
			URL:           cells[1].(string),
		}
		if entry.URL != "" {
			entries = append(entries, &entry)
		}
	})

	if number > len(entries) {
		number = len(entries)
	}

	return QueryResult{
		Title:   doc.Find("title").Text(),
		Entries: entries[:number],
	}, nil
}

func maxTitleWidth(entries []*entry) int {
	width := 0
	for _, e := range entries {
		count := runewidth.StringWidth(e.Title)
		if count > width {
			width = count
		}
	}
	return width
}

func maxURLWidth(entries []*entry) int {
	width := 0
	for _, e := range entries {
		count := utf8.RuneCountInString(e.URL)
		if count > width {
			width = count
		}
	}
	return width
}

func showResult(result QueryResult, url string) {
	entries := result.Entries
	if len(entries) == 0 {
		fmt.Println("カテゴリが見つかりませんでした ʕ◔ϖ◔ʔ")
		fmt.Printf("  url: %s\n\n", url)
		return
	}

	fmt.Printf("%s : %d 件\n",
		result.Title,
		len(entries),
	)
	fmt.Printf("  url: %s\n\n", url)

	titleWidth := maxTitleWidth(entries)
	titleFmt := fmt.Sprintf("%%-%ds", titleWidth)

	urlWidth := maxURLWidth(entries)
	urlFmt := fmt.Sprintf("%%-%ds", urlWidth)

	fmt.Fprintf(color.Output, " %s | %s \n",
		color.BlueString(titleFmt, "Title"),
		fmt.Sprintf(urlFmt, "Url"),
	)
	fmt.Println(strings.Repeat("-", titleWidth+urlWidth+16))
	for _, e := range entries {
		fmt.Fprintf(color.Output, " %s | %s \n",
			color.BlueString(runewidth.FillRight(e.Title, titleWidth)),
			fmt.Sprintf(urlFmt, e.URL),
		)
	}
}

func showResultAsJson(result QueryResult, err error) {
	enc := json.NewEncoder(os.Stdout)
	if err != nil {
		enc.Encode(JsonFormat{Entries: []*entry{}, Error: err.Error()})
		return
	}
	err = enc.Encode(JsonFormat{Entries: result.Entries, Error: ""})
	if err != nil {
		fmt.Print(err)
	}
}
