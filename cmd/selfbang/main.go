package main

import (
	"encoding/json"
	html "html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	text "text/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

type bang struct {
	C  string
	D  string
	R  int
	S  string
	SC string
	U  string
}

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	mu    sync.RWMutex
	bangs = make(map[string]bang)
)

func main() {
	go func() {
		for c := time.Tick(24 * time.Hour); ; <-c {
			func() {
				req, err := http.NewRequest("GET", "https://duckduckgo.com/bang.js", nil)
				if err != nil {
					log.Println(err)
					return
				}

				resp, err := client.Do(req)
				if err != nil {
					log.Println(err)
					return
				}

				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Println(err)
					}
				}()

				if resp.StatusCode != 200 {
					log.Println(err)
					return
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}

				var v []struct {
					C  string `json:"c"`
					D  string `json:"d"`
					R  int    `json:"r"`
					S  string `json:"s"`
					SC string `json:"sc"`
					T  string `json:"t"`
					U  string `json:"u"`
				}

				if err := json.Unmarshal(body, &v); err != nil {
					log.Println(err)
					return
				}

				var b = make(map[string]bang)
				for _, object := range v {
					b[object.T] = bang{
						C:  object.C,
						D:  object.D,
						R:  object.R,
						S:  object.S,
						SC: object.SC,
						U:  object.U,
					}
				}

				mu.Lock()
				defer mu.Unlock()
				bangs = b
			}()
		}
	}()

	app := fiber.New()

	h := html.Must(html.ParseFiles("./index.html"))
	app.Get("/", func(c *fiber.Ctx) error {
		if _, ok := c.Queries()["q"]; !ok {
			c.Response().Header.Add("content-type", "text/html; charset=utf-8")
			return h.Execute(c, c.BaseURL())
		}

		query := strings.TrimSpace(c.Query("q"))
		if query == "" {
			return c.Redirect("https://google.com")
		}

		re := regexp.MustCompile(`![^!\s]{1,}`)
		match := re.FindString(query)
		if match == "" {
			return c.Redirect("https://www.google.com/search?q=" + url.QueryEscape(query))
		}

		mu.RLock()
		val, ok := bangs[strings.TrimPrefix(match, "!")]
		mu.RUnlock()
		if !ok {
			return c.Redirect("https://www.google.com/search?q=" + url.QueryEscape(query))
		}

		remaining := strings.TrimSpace(re.ReplaceAllString(query, ""))
		if remaining == "" {
			return c.Redirect("https://" + val.D)
		}

		return c.Redirect(strings.Replace(val.U, "{{{s}}}", url.PathEscape(remaining), 1))
	})

	t := text.Must(text.ParseFiles("./opensearch.xml"))
	app.Get("/opensearch.xml", func(c *fiber.Ctx) error {
		c.Response().Header.Add("content-type", "application/opensearchdescription+xml")
		return t.Execute(c, c.BaseURL())
	})

	app.Static("/", "./public")
	app.Listen(":3000")
}
