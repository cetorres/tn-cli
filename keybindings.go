package main

import (
	"log"

	"github.com/jroimartin/gocui"
)

func SetKeyBindings(g *gocui.Gui) {
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

	if err := g.SetKeybinding("", 'w', gocui.ModNone, GoUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, GoDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 's', gocui.ModNone, GoDown); err != nil {
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

	if err := g.SetKeybinding(LIST_VIEW, 'a', gocui.ModNone, LoadPreviewsPage); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(LIST_VIEW, gocui.KeyArrowRight, gocui.ModNone, LoadNextPage); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(LIST_VIEW, 'd', gocui.ModNone, LoadNextPage); err != nil {
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

	if err := g.SetKeybinding(READER_VIEW, 'a', gocui.ModNone, OpenArticleWeb); err != nil {
		log.Panicln(err)
	}
}