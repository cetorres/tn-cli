# tn-cli
**TabNews CLI** is a terminal based application (TUI) for the Brazilian news website [TabNews](https://(tabnews.com.br)).

It is made in Go using the libraries:

- [gocui](https://github.com/jroimartin/gocui) for the TUI. This is one the best TUI libraries I've even seen. Supports even mouse interaction!
- [go-term-markdown](https://github.com/MichaelMure/go-term-markdown) for Markdown rendering.

I created this TUI app inspired by the [lazygit](https://github.com/jesseduffield/lazygit) great Git terminal client, that also uses **gocui**.

## Demo

![Demo](tn-cli-demo.gif)

## Installation

### Via source code
- Clone the repo or download and extract the zip file.
- Run: `go build` inside the directory to compile and create the executable.
- Run: `./tn-cli`

### Pre-built binaries
- Download the binary for your platform from the releases.
  
## Roadmap

The project needs help with these items.

- [x] Main interface created
- [x] Accessing contents API
- [x] Load TabNews **recents** posts via API
- [x] Load TabNews **relevant** posts via API
- [x] Show articles list on the left side view
- [x] Show article content on the right side reader view
- [x] Load article pages with arrows
- [x] Caching loaded data for faster access
- [x] Fix known bug on the left side list scroll
- [x] Show a bottom line with available commands and app version
- [x] Improve the reader view to show contents better, with `markdown` and links.
- [ ] Add your feature to the list...

## List of commands / key bindings

| Action | Key |
|--------|-----|
| Refresh Content | <kbd>R</kbd> |
| Quit | <kbd>Ctrl+C</kbd>, <kbd>Q</kbd> |
| Scroll Up | <kbd>↑</kbd> |
| Scroll Down | <kbd>↓</kbd> |
| Toggle Views | <kbd>tab</kbd> |
| Change to Relevant | <kbd>1</kbd> |
| Change to Recent | <kbd>2</kbd> |
| Next Page | <kbd>→</kbd> |
| Previous Page | <kbd>←</kbd> |

