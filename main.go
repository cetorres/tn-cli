package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/jroimartin/gocui"
)

const (
	APP_TITLE		   = "TabNews"
	APP_VERSION    = "1.0"
	APP_COPYRIGHT  = "(c) 2022 Carlos E. Torres"
	APP_GITHUB		 = "https://github.com/cetorres/tn-cli"
	LIST_VIEW			 = "list"
	READER_VIEW		 = "reader"
	BOTTOM_VIEW		 = "bottom"
	VERSION_VIEW	 = "version"
	RECENT_ITEMS_VIEW = "recentitems"
	RELEVANT_ITEMS_VIEW = "relevantitems"
	OPEN_ARTICLE_VIEW = "openarticleview"
	PAGE_NUMBER_VIEW  = "pagenumberview"
	READER_PADDING = 1	
	BOTTOM_HELP	   = "1-2: muda filtro, ←/→: muda páginas, ↑/↓: cima/baixo, r: recarrega, tab: alterna quadros, i: info, q: sair"
	APP_LOGO			 = `
 _                    _ _ 
| |_ _ __         ___| (_)
| __| '_ \ _____ / __| | |
| |_| | | |_____| (__| | |
 \__|_| |_|      \___|_|_|														
	`
)

var (
	viewArr = []string{LIST_VIEW, READER_VIEW}
	active  = 0
	contents = []Content{}
	selected = 0
	cachedContents = make(map[int][]Content)
	cachedArticles = make(map[string]*Article)
)

func selectView(g *gocui.Gui, v *gocui.View, viewId int) error {
	name := viewArr[viewId]

	g.SetCurrentView(name)

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
		v.Title = APP_TITLE
		v.Editable = false
		v.Autoscroll = false
		v.Wrap = false

		g.SetCurrentView(LIST_VIEW)

		g.Update(func(g *gocui.Gui) error {
			LoadList(g, v)
			return nil
		})
	}

	// Set up new and relevant views (selectable)
	if v, err := g.SetView(RELEVANT_ITEMS_VIEW, 10, -1, 21, 1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
		v.Frame = false
		v.Highlight = true
		v.SelFgColor = gocui.ColorCyan
		fmt.Fprintln(v, "Relevantes")
	}
	if v, err := g.SetView(RECENT_ITEMS_VIEW, 21, -1, 30, 1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
		v.Frame = false
		// v.Highlight = true
		v.SelFgColor = gocui.ColorCyan
		fmt.Fprintln(v, "Recentes")
	}

	// Set up page number view
	// pageNumberStr := fmt.Sprintf("Page %d", currentPage)
	// if v, err := g.SetView(PAGE_NUMBER_VIEW, maxX/3-3-len(pageNumberStr), maxY-3, maxX/3-2, maxY-1); err != nil {
	// 	if err != gocui.ErrUnknownView {
	// 		return err
	// 	}
	// 	v.Wrap = false
	// 	v.Frame = false
	// 	fmt.Fprintln(v, pageNumberStr)
	// }
	UpdatePageNumber(g)

	// Set up reader view
	if v, err := g.SetView(READER_VIEW, maxX/3, 0, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = ""
		v.Wrap = true
		v.Autoscroll = false
	}

	// Set up open article view
	openArticleStr := "a: abrir na web"
	if v, err := g.SetView(OPEN_ARTICLE_VIEW, maxX-len(openArticleStr)-3, maxY-3, maxX-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
		v.Frame = false
		v.Autoscroll = false
		v.FgColor = gocui.ColorCyan

		fmt.Fprintln(v, openArticleStr)
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

func RefreshContent(g *gocui.Gui, v0 *gocui.View) error {
	v, _ := g.View(LIST_VIEW)

	// Clear caches
	cachedContents = make(map[int][]Content)
	cachedArticles = make(map[string]*Article)
	ClearDiskCache()

	// Reset positions
	selected = 0
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	// Load content
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

	v2, _ := g.View(READER_VIEW)
	v2.Clear()
	v2.Title = ""
	fmt.Fprintln(v2, "Carregando...")

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

func LoadContent(g *gocui.Gui, v *gocui.View) error {
	v2, _ := g.View(READER_VIEW)
	v2.Clear()
	fmt.Fprintln(v2, "Carregando...")

	// View title (username)
	tabcoin := "tabcoins"
	if contents[selected].Tabcoins == 0 || contents[selected].Tabcoins == 1 {
		tabcoin = "tabcoin"
	}
	tz, _ := time.LoadLocation("Local")
	v2.Title = fmt.Sprintf("%s (%d %s) (%s)", contents[selected].OwnerUsername, contents[selected].Tabcoins, tabcoin, contents[selected].PublishedAt.In(tz).Format(time.RFC822))

	// Markdown options
	maxX, _ := v2.Size()
	opts := []markdown.Options{
		// needed when going through gocui
		markdown.WithImageDithering(markdown.DitheringWithBlocks),
	}	

	// Load article
	g.Update(func(g *gocui.Gui) error {
		article, err := DownloadArticle(contents[selected].OwnerUsername, contents[selected].Slug, contents[selected].ID)

		if err != nil {
			fmt.Fprintln(v2, "Não possível carregar artigo.")
			return nil
		}

		v2.Clear()

		// Article title
		result1 := markdown.Render(contents[selected].Title, maxX-READER_PADDING, READER_PADDING, opts...)
		_, _ = v2.Write(result1)
		fmt.Fprintln(v2, "")
		
		// Article body
		result2 := markdown.Render(article.Body, maxX-READER_PADDING, READER_PADDING, opts...)
		_, _ = v2.Write(result2)
		return nil
	})

	v2.SetCursor(0, 0)
	v2.SetOrigin(0, 0)
	return nil
}

func LoadRelevant(g *gocui.Gui, v0 *gocui.View) error {
	v, _ := g.View(LIST_VIEW)
	v2, _ := g.View(RELEVANT_ITEMS_VIEW)
	v3, _ := g.View(RECENT_ITEMS_VIEW)

	// Select view
	v2.Highlight = true
	v3.Highlight = false

	// Reset list view
	cachedContents = make(map[int][]Content)
	selected = 0
	currentPage = 1
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	// Load list
	currentStrategy = "relevant"
	LoadList(g, v)

	return nil
}

func LoadRecent(g *gocui.Gui, v0 *gocui.View) error {
	v, _ := g.View(LIST_VIEW)
	v2, _ := g.View(RELEVANT_ITEMS_VIEW)
	v3, _ := g.View(RECENT_ITEMS_VIEW)

	// Select view
	v2.Highlight = false
	v3.Highlight = true

	// Reset list view
	cachedContents = make(map[int][]Content)
	selected = 0
	currentPage = 1
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	// Load list
	currentStrategy = "new"
	LoadList(g, v)

	return nil
}

func UpdatePageNumber(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	g.DeleteView(PAGE_NUMBER_VIEW)
	pageNumberStr := fmt.Sprintf("Page %d", currentPage)
	if v, err := g.SetView(PAGE_NUMBER_VIEW, maxX/3-3-len(pageNumberStr), maxY-3, maxX/3-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
		v.Frame = false
		fmt.Fprintln(v, pageNumberStr)
	}
	return nil
}

func HandleClickVersion(g *gocui.Gui, v *gocui.View) error {
	openbrowser(APP_GITHUB)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	SaveCacheToDisk()
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
	if active == 0 {
		if selected >= len(contents)-1 {
			return nil
		}
	}

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
		if selected >= len(contents)-1 {
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

	UpdatePageNumber(g)
	
	return nil
}

func LoadNextPage(g *gocui.Gui, v *gocui.View) error {
	currentPage += 1

	selected = 0
	LoadList(g ,v)
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	UpdatePageNumber(g)
	
	return nil
}

func ShowVersion() {
	fmt.Println(APP_LOGO)
	fmt.Println(APP_TITLE + " CLI", APP_VERSION)
	fmt.Println(APP_COPYRIGHT)
}

func ShowInfo(g *gocui.Gui, v *gocui.View) error {
	v2, _ := g.View(READER_VIEW)
	v2.Clear()
	v2.Title = "Info"

	v2.SetCursor(0, 0)
	v2.SetOrigin(0, 0)

	infoText := fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s\n%s\n\n%s", 
		APP_LOGO,
		APP_TITLE + " CLI",
		"Versão " + APP_VERSION,
		APP_COPYRIGHT, 
		"Cliente de terminal para o website TabNews (https://tabnews.com.br).",
		"Desenvolvido em Go, usando bibliotecas gocui e go-term-markdown.",
		"GitHub: https://github.com/cetorres/tn-cli")

	fmt.Fprintln(v2, infoText)
	return nil
}

func OpenArticleWeb(g *gocui.Gui, v *gocui.View) error {
	openbrowser(URL_SITE + "/" + contents[selected].OwnerUsername + "/" + contents[selected].Slug)
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

	LoadCacheToDisk()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'r', gocui.ModNone, RefreshContent); err != nil {
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

	if err := g.SetKeybinding("", '1', gocui.ModNone, LoadRelevant); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", '2', gocui.ModNone, LoadRecent); err != nil {
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
	
	if err := g.SetKeybinding(RELEVANT_ITEMS_VIEW, gocui.MouseLeft, gocui.ModNone, LoadRelevant); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(RECENT_ITEMS_VIEW, gocui.MouseLeft, gocui.ModNone, LoadRecent); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(OPEN_ARTICLE_VIEW, gocui.MouseLeft, gocui.ModNone, OpenArticleWeb); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'a', gocui.ModNone, OpenArticleWeb); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}