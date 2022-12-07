package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/jroimartin/gocui"
)

const (
	APP_TITLE		   = "TabNews CLI"
	APP_VERSION    = "1.0"
	APP_COPYRIGHT  = "(c) 2022 Carlos E. Torres"
	APP_GITHUB		 = "https://github.com/cetorres/tn-cli"
	URL_RECENTS    = "https://www.tabnews.com.br/recentes/rss"
	URL_API				 = "https://www.tabnews.com.br/api/v1"
	URL_CONTENTS   = URL_API + "/contents"
	LIST_VIEW			 = "list"
	READER_VIEW		 = "reader"
	BOTTOM_VIEW		 = "bottom"
	VERSION_VIEW	 = "version"
	READER_PADDING = 1
	PAGE_SIZE			 = 40
	BOTTOM_HELP	   = "1-2: seleciona quadro, ←/→: carrega páginas, ↑/↓: cima/baixo, r: recarrega, tab: alterna quadros, i: info, q: sair"
	APP_LOGO			 = `
 _                    _ _ 
| |_ _ __         ___| (_)
| __| '_ \ _____ / __| | |
| |_| | | |_____| (__| | |
 \__|_| |_|      \___|_|_|														
	`
)

type Content struct {
	ID                string      `json:"id"`
	OwnerID           string      `json:"owner_id"`
	ParentID          interface{} `json:"parent_id"`
	Slug              string      `json:"slug"`
	Title             string      `json:"title"`
	Status            string      `json:"status"`
	SourceURL         interface{} `json:"source_url"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	PublishedAt       time.Time   `json:"published_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	Tabcoins          int         `json:"tabcoins"`
	OwnerUsername     string      `json:"owner_username"`
	ChildrenDeepCount int         `json:"children_deep_count"`
}

type Article struct {
	ID                string      `json:"id"`
	OwnerID           string      `json:"owner_id"`
	ParentID          interface{} `json:"parent_id"`
	Slug              string      `json:"slug"`
	Title             string      `json:"title"`
	Body              string      `json:"body"`
	Status            string      `json:"status"`
	SourceURL         interface{} `json:"source_url"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	PublishedAt       time.Time   `json:"published_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	OwnerUsername     string      `json:"owner_username"`
	Tabcoins          int         `json:"tabcoins"`
	ChildrenDeepCount int         `json:"children_deep_count"`
}

var (
	viewArr = []string{LIST_VIEW, READER_VIEW}
	active  = 0
	contents = []Content{}
	selected = 0
	currentPage = 1
	cachedContents = make(map[int][]Content)
	cachedArticles = make(map[string]*Article)
)

func DownloadContent() ([]Content, error) {
	// Return cached results if exist
	if len(contents) > 0 && len(cachedContents) > 0 {
		content := cachedContents[currentPage]
		if len(content) > 0 {
			return content, nil
		}
	}

	// Perform HTTP request to load results
	resp, err := http.Get(fmt.Sprintf("%s%s%d%s%d", URL_CONTENTS, "?per_page=", PAGE_SIZE, "&page=", currentPage))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()

	var content = []Content{}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	decoder.Decode(&content)

	// Save page results into cache
	cachedContents[currentPage] = content

	return content, nil
}

func DownloadArticle(username string, slug string, id string) (*Article, error) {
	// Return cached result if exist
	if len(cachedArticles) > 0 {
		article := cachedArticles[id]
		if article != nil {
			return article, nil
		}
	}

	// Perform HTTP request to load results
	resp, err := http.Get(URL_CONTENTS + "/" + username + "/" + slug)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()

	var article = Article{}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	decoder.Decode(&article)

	// Save article into cache
	cachedArticles[id] = &article

	return &article, nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func selectView(g *gocui.Gui, v *gocui.View, viewId int) error {
	name := viewArr[viewId]

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	if viewId == 0 {
		g.Cursor = false
	} else {
		g.Cursor = true
	}

	active = viewId
	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	return selectView(g, v, nextIndex)
}

func selectListView(g *gocui.Gui, v *gocui.View) error {
	return selectView(g, v, 0)
}

func selectReaderView(g *gocui.Gui, v *gocui.View) error {
	return selectView(g, v, 1)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Set up list view
	if v, err := g.SetView(LIST_VIEW, 0, 0, maxX/3-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[ " + APP_TITLE + " ]"
		v.Editable = false
		v.Autoscroll = false
		v.Wrap = false

		if _, err = setCurrentViewOnTop(g, LIST_VIEW); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			LoadList(g, v)
			return nil
		})
	}

	// Set up reader view
	if v, err := g.SetView(READER_VIEW, maxX/3, 0, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[ ]"
		v.Wrap = true
		v.Autoscroll = false
	}

	// Set up version view
	versionStr := "v." + APP_VERSION
	versionViewW := len(versionStr)+2
	if v, err := g.SetView(VERSION_VIEW, maxX-versionViewW, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Wrap = false
		v.Autoscroll = false
		v.FgColor = gocui.ColorYellow

		fmt.Fprintln(v, versionStr)
	}

	// Set up bottom help view
	if v, err := g.SetView(BOTTOM_VIEW, -1, maxY-2, maxX-versionViewW-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Wrap = false
		v.Autoscroll = false
		v.FgColor = gocui.ColorBlue

		fmt.Fprintln(v, BOTTOM_HELP)
	}	
	
	return nil
}

func UpdateList(g *gocui.Gui, v0 *gocui.View) error {
	v, _ := g.View(LIST_VIEW)
	g.Update(func(g *gocui.Gui) error {
		LoadList(g, v)
		return nil
	})
	return nil
}

func LoadList(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	v.Highlight = false
	fmt.Fprintln(v, "Carregando...")

	g.Update(func(g *gocui.Gui) error {
		content, err := DownloadContent()
		contents = content

		v.Clear()

		if err != nil {
			fmt.Fprintln(v, "Não foi possível carregar o conteúdo.")
			fmt.Fprintln(v, err.Error())
			return err
		}

		if len(contents) == 0 {
			fmt.Fprintln(v, "Conteúdo vazio.")
			return err
		}

		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for i, item := range contents {
			maxW, _ := v.Size()
			idStr := fmt.Sprintf("%d. ", ((i+1) + ((currentPage-1)*PAGE_SIZE)))
			spacesToAdd := strings.Repeat(" ", maxW)
			s := fmt.Sprintf("%s%v", idStr, item.Title + spacesToAdd)
			fmt.Fprintln(v, s)
		}

		LoadContent(g, nil)
		return nil
	})
	
	return nil
}

func openbrowser(url string) {
  var err error
  switch runtime.GOOS {
  case "linux":
    err = exec.Command("xdg-open", url).Start()
  case "windows":
    err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
  case "darwin":
    err = exec.Command("open", url).Start()
  default:
    err = fmt.Errorf("unsupported platform")
  }
  if err != nil {
    panic(err)
  }
}

func HandleClickVersion(g *gocui.Gui, v *gocui.View) error {
	openbrowser(APP_GITHUB)
	return nil
}

func LoadContent(g *gocui.Gui, v *gocui.View) error {
	v2, _ := g.View(READER_VIEW)
	v2.Clear()

	// View title (username)
	v2.Title = "[ " + contents[selected].OwnerUsername + " ]"

	// Markdown options
	maxX, _ := v2.Size()
	opts := []markdown.Options{
		// needed when going through gocui
		markdown.WithImageDithering(markdown.DitheringWithBlocks),
	}

	// Article title
	result := markdown.Render(contents[selected].Title, maxX-READER_PADDING, READER_PADDING, opts...)
	_, _ = v2.Write(result)
	fmt.Fprintln(v2, "")

	// Article body
	g.Update(func(g *gocui.Gui) error {
		article, err := DownloadArticle(contents[selected].OwnerUsername, contents[selected].Slug, contents[selected].ID)

		if err != nil {
			fmt.Fprintln(v2, "Não possível carregar artigo.")
			return nil
		}
		
		result := markdown.Render(article.Body, maxX-READER_PADDING, READER_PADDING, opts...)
		_, _ = v2.Write(result)
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
		if selected == len(contents)-1 {
			return nil
		}
		selected = selected + 1
		LoadContent(g, nil)
	}
	return nil
}

func ScrollUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		v.SetCursor(cx, cy-1)
		v.SetOrigin(ox, oy-1)
	}
	return nil
}

func ScrollDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		v.SetCursor(cx, cy+1)
		v.SetOrigin(ox, oy+1)
	}
	return nil
}

func SelectListItem(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		selectListView(g, v)

		_, cy := v.Cursor()
		if cy > PAGE_SIZE-1 {
			return nil
		}
		selected = cy
		LoadContent(g, nil)
	}
	return nil
}

func LoadPreviewsPage(g *gocui.Gui, v *gocui.View) error {
	if currentPage > 2 {
		currentPage -= 1
	} else {
		currentPage = 1
	}

	selected = 0
	LoadList(g ,v)
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)
	
	return nil
}

func LoadNextPage(g *gocui.Gui, v *gocui.View) error {
	currentPage += 1

	selected = 0
	LoadList(g ,v)
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)
	
	return nil
}

func ShowVersion() {
	fmt.Println(APP_LOGO)
	fmt.Println(APP_TITLE, APP_VERSION)
	fmt.Println(APP_COPYRIGHT)
}

func ShowInfo(g *gocui.Gui, v *gocui.View) error {
	v2, _ := g.View(READER_VIEW)
	v2.Clear()
	v2.Title = "[ Info ]"

	infoText := fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n%s", 
		APP_LOGO,
		APP_COPYRIGHT, 
		"Cliente de terminal para o website TabNews (https://tabnews.com.br).",
		"Desenvolvido em Go, usando bibliotecas gocui e go-term-markdown.",
		"GitHub: https://github.com/cetorres/tn-cli")

	fmt.Fprintln(v2, infoText)
	return nil
}

func main() {
	if len(os.Args) == 2 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			ShowVersion()
			return
		}
	}

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.ASCII = false
	g.FgColor = gocui.ColorWhite
	g.SelFgColor = gocui.ColorGreen
	g.Mouse = true

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

	if err := g.SetKeybinding("", '1', gocui.ModNone, selectListView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", '2', gocui.ModNone, selectReaderView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'i', gocui.ModNone, ShowInfo); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(LIST_VIEW, gocui.KeyArrowLeft, gocui.ModNone, LoadPreviewsPage); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(LIST_VIEW, gocui.KeyArrowRight, gocui.ModNone, LoadNextPage); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(LIST_VIEW, gocui.MouseLeft, gocui.ModNone, SelectListItem); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(READER_VIEW, gocui.MouseLeft, gocui.ModNone, selectReaderView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(READER_VIEW, gocui.MouseWheelUp, gocui.ModNone, ScrollUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(READER_VIEW, gocui.MouseWheelDown, gocui.ModNone, ScrollDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(VERSION_VIEW, gocui.MouseLeft, gocui.ModNone, HandleClickVersion); err != nil {
		log.Panicln(err)
	}	

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}