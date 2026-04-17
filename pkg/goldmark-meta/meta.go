// 这个包用来解析markdown的yaml部分的meta数据.
package goldmark_meta

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

type data struct {
	Map   map[string]interface{}
	Error error
	Node  gast.Node
}

var contextKey = parser.NewContextKey()

// Option interface sets options for this extension.
type Option interface {
	metaOption()
}

// Get returns a YAML metadata.
func Get(pc parser.Context) map[string]interface{} {
	v := pc.Get(contextKey)
	if v == nil {
		return nil
	}
	d := v.(*data)
	return d.Map
}

type metaParser struct {
}

var defaultParser = &metaParser{}

// NewParser returns a BlockParser that can parse YAML metadata blocks.
func NewParser() parser.BlockParser {
	return defaultParser
}

func isSeparator(line []byte) bool {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	for i := 0; i < len(line); i++ {
		if line[i] != '-' {
			return false
		}
	}
	return true
}

func (b *metaParser) Trigger() []byte {
	return []byte{'-'}
}

func (b *metaParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	linenum, _ := reader.Position()
	if linenum != 0 {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	if isSeparator(line) {
		return gast.NewTextBlock(), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *metaParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if isSeparator(line) && !util.IsBlank(line) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (b *metaParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		buf.Write(segment.Value(reader.Source()))
	}
	d := &data{}
	d.Node = node
	meta := map[string]interface{}{}
	if err := yaml.Unmarshal(buf.Bytes(), &meta); err != nil {
		d.Error = err
	} else {
		d.Map = meta
	}

	pc.Set(contextKey, d)

	if d.Error == nil {
		node.Parent().RemoveChild(node.Parent(), node)
	}
}

func (b *metaParser) CanInterruptParagraph() bool {
	return false
}

func (b *metaParser) CanAcceptIndentedLine() bool {
	return false
}

type astTransformer struct {
	transformerConfig
}

type transformerConfig struct {
	// Renders metadata as an html table.
	Table bool

	// Stores metadata in ast.Document.Meta().
	StoresInDocument bool
}

type transformerOption interface {
	Option

	// SetMetaOption sets options for the metadata parser.
	SetMetaOption(*transformerConfig)
}

func newTransformer(opts ...transformerOption) parser.ASTTransformer {
	p := &astTransformer{
		transformerConfig: transformerConfig{
			Table:            false,
			StoresInDocument: false,
		},
	}
	for _, o := range opts {
		o.SetMetaOption(&p.transformerConfig)
	}
	return p
}

func (a *astTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	dtmp := pc.Get(contextKey)
	if dtmp == nil {
		return
	}
	d := dtmp.(*data)
	if d.Error != nil {
		msg := gast.NewString([]byte(fmt.Sprintf("<!-- %s -->", d.Error)))
		msg.SetCode(true)
		d.Node.AppendChild(d.Node, msg)
		return
	}

	if a.StoresInDocument {
		for k, v := range d.Map {
			node.AddMeta(k, v)
		}
	}
}

type meta struct {
	options []Option
}

var Meta = &meta{}

// New returns a new Meta extension.
func New(opts ...Option) goldmark.Extender {
	e := &meta{
		options: opts,
	}
	return e
}

// implements goldmark.Extender.
func (e *meta) Extend(m goldmark.Markdown) {
	topts := []transformerOption{}
	for _, opt := range e.options {
		if topt, ok := opt.(transformerOption); ok {
			topts = append(topts, topt)
		}
	}
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewParser(), 0),
		),
	)
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(newTransformer(topts...), 0),
		),
	)
}
