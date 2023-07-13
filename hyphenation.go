package hyphenation

import (
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	hyphenator "github.com/speedata/hyphenation"
)

type HyphenationConfig struct {
	hyphenator *hyphenator.Lang
}

type HyphenationOption interface {
	SetHyphenationOption(*HyphenationConfig)
}

type withHyphenationFile struct {
	value *os.File
}

func (o *withHyphenationFile) SetHyphenationOption(h *HyphenationConfig) {
	l, err := hyphenator.New(o.value)
	if err != nil {
		panic(err)
	}

	h.hyphenator = l
}

func WithHyphenationFile(value *os.File) HyphenationOption {
	return &withHyphenationFile{
		value: value,
	}
}

func NewHyphenation(opts ...HyphenationOption) goldmark.Extender {
	h := &HyphenationConfig{}
	for _, opt := range opts {
		opt.SetHyphenationOption(h)
	}

	return h
}

func (h *HyphenationConfig) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHyphenationHTMLRenderer(h.hyphenator), 500),
	))
}

type HyphenationHTMLRenderer struct {
	hyphenator *hyphenator.Lang
}

func NewHyphenationHTMLRenderer(hyphenator *hyphenator.Lang) *HyphenationHTMLRenderer {
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
