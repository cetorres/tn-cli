package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	goose "github.com/advancedlogic/GoOse"
	"github.com/jroimartin/gocui"
	"github.com/mmcdole/gofeed"
)

const (
	APP_TITLE		  = "TabNews CLI"
	APP_VERSION   = "1.0.0"
	APP_COPYRIGHT = "(c) 2022 Carlos E. Torres"
	URL_RECENTS   = "https://www.tabnews.com.br/recentes/rss"
	LIST_VIEW			= "list"
	READER_VIEW		= "reader"
)

type Event struct {
	Id        int
	Title     string
	Author    string
	Url       string
	Summary   string
	Published string
	Content   string
}

var (
	viewArr = []string{LIST_VIEW, READER_VIEW}
	active  = 0
	feedContent = []Event{}
	selected = 0
)

func CheckUrl(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL(url)
}

func DownloadEvents(url string) ([]Event, error) {
	feed, err := CheckUrl(url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to retrieve news from: '%v'", url))
	}

	var events []Event
	for _, item := range feed.Items {
		e := Event{}
		e.Title = item.Title
		if item.Author != nil {
			e.Author = item.Author.Name
		} else {
			// Get author username from the link
			re := regexp.MustCompile(`(?:br\/)(\w+)(?:\/)`)
			match := re.FindStringSubmatch(item.Link)
			e.Author = match[1]
		}
		e.Url = item.Link
		if len(item.Description) > 0 {
			e.Summary = trim(item.Description)
		} else {
			e.Summary = "No summary available"
		}
		e.Published = item.Published
		e.Content = item.Content

		events = append(events, e)
	}

	return events, nil
}

func trim(desc string) string {
	var re = regexp.MustCompile(`(<.*?>)`)

	// remove html
	desc = re.ReplaceAllString(desc, ``)

	// remove spaces
	desc = strings.TrimSpace(desc)

	return desc
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	active = nextIndex
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(LIST_VIEW, 0, 0, maxX/3-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = APP_TITLE
		v.Editable = false
		v.Autoscroll = false
		v.Wrap = false

		if _, err = setCurrentViewOnTop(g, LIST_VIEW); err != nil {
			return err
		}

		fmt.Fprintln(v, "Carregando...")

		g.Update(func(g *gocui.Gui) error {
			LoadList(g, v, URL_RECENTS)
			return nil
		})
	}

	if v, err := g.SetView(READER_VIEW, maxX/3, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[ ]"
		v.Wrap = true
		v.Autoscroll = false
	}
	
	return nil
}

func UpdateList(g *gocui.Gui, v0 *gocui.View) error {
	v, _ := g.View(LIST_VIEW)
	g.Update(func(g *gocui.Gui) error {
		LoadList(g, v, URL_RECENTS)
		return nil
	})
	return nil
}

func GetContent(url string) ([]string, error) {
	g := goose.New()
	article, err := g.ExtractFromURL(url)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(article.CleanedText, "\n\n")

	return lines, nil
}

func LoadList(g *gocui.Gui, v *gocui.View, url string) error {
	feed, err := DownloadEvents(url)
	feedContent = feed

	if err != nil {
		fmt.Fprintln(v, "Não foi possível carregar o conteúdo.")
		return err
	}

	v.Clear()

	for i, item := range feed {
		s := fmt.Sprintf("%d. %v", i+1, item.Title)
		fmt.Fprintln(v, s)
	}

	LoadContent(g, nil)
	
	return nil
}

func LoadContent(g *gocui.Gui, v *gocui.View) error {
	v2, _ := g.View(READER_VIEW)
	v2.Clear()

	v2.Title = "[ " + feedContent[selected].Author + " ]"
	
	fmt.Fprintln(v2, feedContent[selected].Title + "\n")

	g.Update(func(g *gocui.Gui) error {
		goo := goose.New()
		content, _ := goo.ExtractFromURL(feedContent[selected].Url) 
		fmt.Fprintln(v2, content.CleanedText)
		return nil
	})

	v2.SetCursor(0, 0)
	v2.SetOrigin(0, 0)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func GoUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}

	if active == 0 {
		if selected == 0 {
			return nil
		}
		selected = selected - 1
		LoadContent(g, nil)
	}
	return nil
}

func GoDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}

	if active == 0 {
		if selected == len(feedContent)-1 {
			return nil
		}
		selected = selected + 1
		LoadContent(g, nil)
	}
	return nil
}


func main() {
	if len(os.Args) == 2 {
		if os.Args[1] == "--version" {
			fmt.Println(APP_TITLE, APP_VERSION)
			fmt.Println(APP_COPYRIGHT)
			return
		}
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.ASCII = false
	g.FgColor = gocui.ColorWhite
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'r', gocui.ModNone, UpdateList); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, GoUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, GoDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}