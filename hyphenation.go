package hyphenation

import (
	"log"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	"github.com/speedata/hyphenation"
)

type Hyphenation struct {
	hyphenator *hyphenation.Lang
}

type HyphenationOption interface {
	parser.Option
	SetHyphenationOption(*Hyphenation)
}

type withHyphenationFile struct {
	value string
}

func (o *withHyphenationFile) SetParserOption(c *parser.Config) {
	// This option doesn't affect the parser, so do nothing here.
}

func (o *withHyphenationFile) SetHyphenationOption(h *Hyphenation) {
	filename := o.value
	r, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	l, err := hyphenation.New(r)
	if err != nil {
		log.Fatal(err)
	}

	h.hyphenator = l
}

func WithHyphenationFile(value string) HyphenationOption {
	return &withHyphenationFile{
		value: value,
	}
}

func NewHyphenation(opts ...HyphenationOption) goldmark.Extender {
	h := &Hyphenation{}
	for _, opt := range opts {
		opt.SetHyphenationOption(h)
	}
	return h
}

func (h *Hyphenation) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHyphenationHTMLRenderer(h.hyphenator), 500),
	))
}

type HyphenationHTMLRenderer struct {
	hyphenator *hyphenation.Lang
}

func NewHyphenationHTMLRenderer(hyphenator *hyphenation.Lang) *HyphenationHTMLRenderer {
	return &HyphenationHTMLRenderer{
		hyphenator: hyphenator,
	}
}

func (r *HyphenationHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindText, r.Render)
}

func (r *HyphenationHTMLRenderer) Render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	text := string(node.Text(source))
	hyphenationPoints := r.hyphenator.Hyphenate(text)

	var hyphenatedText strings.Builder
	lastIndex := 0
	for _, point := range hyphenationPoints {
		hyphenatedText.WriteString(text[lastIndex:point])
		hyphenatedText.WriteString("&shy;")
		lastIndex = point
	}
	hyphenatedText.WriteString(text[lastIndex:])

	_, err := w.WriteString(hyphenatedText.String())
	return ast.WalkSkipChildren, err
}
