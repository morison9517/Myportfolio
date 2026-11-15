// =============================================================================
// product_markdown.go = 作品ページの本文(Markdown)を HTML にする
//
// 変換は goldmark(v2)に任せ、このサイト用に4か所だけ出し方を変えている。
//
//	見出し   … # も ## として出す。ページの h1 は作品名の1つだけにするため
//	写真     … 写真だけの段落は <figure> にし、押すと拡大できる形にする。
//	           同じ段落に2枚以上並べると、横並びになる
//	画像の場所 … "Chorus/home.png" と書けば /static/images/products/Chorus/home.png を指す
//	リンク   … 外部のページ(http〜)は新しいタブで開く
//
// ★Markdown の中に直接書いた HTML は出さない(goldmark の既定のまま)。
//
//	javascript: のような危険なリンクも出さない。
//	見た目を増やしたいときは、HTML を書けるようにするのではなく、ここに規則を足す。
//
// ★変換するのは .md を読んだときだけ(product_load.go)。ページを開くたびではない。
// =============================================================================
package models

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"html/template"
	"io"
	"net/url"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/util"
)

// productImageBase = 本文の写真の置き場所。"Chorus/home.png" はここから見た場所。
const productImageBase = "/static/images/products/"

var (
	markdownParser = parser.New()

	markdownRenderer = html.New(
		// ★日本語の段落の途中で改行したとき、文字の間に空白が入らないようにする。
		//   既定のままだと、改行がブラウザで半角の空白として表示される。
		html.WithLineBreakStrategy(html.CSSText3LineBreakStrategy),

		// ★ここに書いた種類だけ、goldmark の既定の出し方を置き換える。
		html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
			ast.KindHeading:   html.NodeRendererFunc(renderHeading),
			ast.KindParagraph: html.NodeRendererFunc(renderParagraph),
			ast.KindImage:     html.NodeRendererFunc(renderImage),
			ast.KindLink:      html.NodeRendererFunc(renderLink),
		}),
	)
)

// renderMarkdown = 本文の Markdown を HTML にする。
func renderMarkdown(source []byte) (template.HTML, error) {
	doc := markdownParser.Parse(source)

	var buf bytes.Buffer
	if err := markdownRenderer.Render(&buf, source, doc); err != nil {
		return "", err
	}

	// ★ここで template.HTML にするので、画面側ではエスケープされない。
	//   中身はこのファイルの規則と goldmark が作ったものだけ。
	return template.HTML(buf.String()), nil
}

// ---------------------------------------------------------------------------
// 見出し
// ---------------------------------------------------------------------------

// renderHeading = # も ## と同じ h2 にする。
//
// ★ページの h1 は作品名の1つだけにする決まり(読み上げソフトが h1 を目印にするため)。
func renderHeading(w io.Writer, _ []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	bw := w.(util.BufWriter)
	level := max(node.(*ast.Heading).Level, 2)

	if entering {
		_, _ = fmt.Fprintf(bw, "<h%d>", level)
	} else {
		_, _ = fmt.Fprintf(bw, "</h%d>\n", level)
	}

	return ast.WalkContinue, nil
}

// ---------------------------------------------------------------------------
// 段落と写真
// ---------------------------------------------------------------------------

// renderParagraph = 写真だけの段落は写真の枠に、それ以外は普通の <p> にする。
func renderParagraph(w io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	bw := w.(util.BufWriter)

	if images := photoParagraph(node, source); images != nil {
		if entering {
			writePhotos(bw, source, images)
		}
		// ★中身(画像)はもう書いたので、goldmark にはたどらせない。
		return ast.WalkSkipChildren, nil
	}

	// 箇条書きの中の段落(項目どうしの間に空行が無いもの)は <p> で包まない。
	// ★goldmark の既定の動きと同じ。
	if html.IsInTightBlock(node) {
		if !entering && node.NextSibling() != nil && node.FirstChild() != nil {
			_ = bw.WriteByte('\n')
		}
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = bw.WriteString("<p>")
	} else {
		_, _ = bw.WriteString("</p>\n")
	}

	return ast.WalkContinue, nil
}

// photoParagraph = 段落の中身が画像だけなら、その画像を返す。文字が混ざっていれば nil。
//
// ★画像どうしの間の空白や改行は、文字として数えない。
//
//	![a](1.png) ![b](2.png)  も、2行に分けて書いても、どちらも「画像だけ」。
func photoParagraph(node ast.Node, source []byte) []*ast.Image {
	var images []*ast.Image

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Image:
			images = append(images, c)
		case *ast.Text:
			if strings.TrimSpace(c.Value.Value(source)) != "" {
				return nil
			}
		default:
			return nil
		}
	}

	return images
}

// writePhotos = 写真を書き出す。1枚なら縦に1つ、2枚以上なら横並びの箱に入れる。
//
// ★見た目は product-detail.css、拡大表示は photo-viewer.js。
//
//	.product-photo-link というクラス名は photo-viewer.js が目印にしているので変えないこと。
func writePhotos(w util.BufWriter, source []byte, images []*ast.Image) {
	if len(images) > 1 {
		_, _ = w.WriteString(`<div class="product-photo-row">` + "\n")
	}

	for _, image := range images {
		src := stdhtml.EscapeString(imageURL(image.Destination.Value(source)))
		alt := stdhtml.EscapeString(plainText(image, source))
		caption := stdhtml.EscapeString(image.Title.Value(source))

		// ★<a> にしておくと、JavaScript が動かないときも押せば写真そのものが開く。
		_, _ = fmt.Fprintf(w,
			`<figure class="product-photo"><a class="product-photo-link" href="%s" data-caption="%s">`+
				`<img src="%s" alt="%s" loading="lazy" decoding="async">`+
				`<span class="visually-hidden">(拡大して表示)</span></a>`,
			src, caption, src, alt)

		if caption != "" {
			_, _ = fmt.Fprintf(w, "<figcaption>%s</figcaption>", caption)
		}

		_, _ = w.WriteString("</figure>\n")
	}

	if len(images) > 1 {
		_, _ = w.WriteString("</div>\n")
	}
}

// renderImage = 文章の途中に置かれた画像(写真だけの段落でないもの)。拡大はしない。
func renderImage(w io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	image := node.(*ast.Image)

	_, _ = fmt.Fprintf(w.(util.BufWriter), `<img src="%s" alt="%s" loading="lazy" decoding="async">`,
		stdhtml.EscapeString(imageURL(image.Destination.Value(source))),
		stdhtml.EscapeString(plainText(image, source)))

	return ast.WalkSkipChildren, nil
}

// imageURL = Markdown に書いた画像の場所を、ブラウザが読める URL に直す。
//
//	"Chorus/home.png"          → "/static/images/products/Chorus/home.png"
//	"/static/images/ogp.jpg"   → そのまま(/ から始まるもの)
//	"https://example.com/a.png" → そのまま
//
// ★javascript: などの危険なものは空にする。
func imageURL(dest string) string {
	if html.IsDangerousURL(dest) {
		return ""
	}

	if strings.HasPrefix(dest, "/") || strings.Contains(dest, "://") {
		return dest
	}

	// ★ファイル名の空白や日本語を、URL で使える形(%20 など)にする。
	return (&url.URL{Path: productImageBase + dest}).EscapedPath()
}

// plainText = 画像の alt 部分(![ここ](...))を、飾りを外した文字だけにする。
func plainText(node ast.Node, source []byte) string {
	var b strings.Builder

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			b.WriteString(text.Value.Value(source))
		} else {
			b.WriteString(plainText(child, source))
		}
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// リンク
// ---------------------------------------------------------------------------

// renderLink = 外部のページ(http〜)へのリンクは新しいタブで開く。
//
// ★見た目(外部リンクの印)と「新しいタブで開きます」の読み上げは、
//
//	詳細ページ下の「リンク」と同じにしてある(product.html)。
func renderLink(w io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	bw := w.(util.BufWriter)
	link := node.(*ast.Link)
	dest := link.Destination.Value(source)

	external := strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://")

	if entering {
		_, _ = bw.WriteString(`<a href="`)
		if !html.IsDangerousURL(dest) {
			// ★URL の書き出しは goldmark の係に任せる(" や < を安全な形にしてくれる)。
			_, _ = html.ContextLinkURLWriter(rc).WriteString(dest)
		}
		_ = bw.WriteByte('"')

		if external {
			_, _ = bw.WriteString(` target="_blank" rel="noopener noreferrer"`)
		}

		_ = bw.WriteByte('>')

		return ast.WalkContinue, nil
	}

	if external {
		_, _ = bw.WriteString(`<span class="visually-hidden">(新しいタブで開きます)</span>` +
			`<i class="fa-solid fa-arrow-up-right-from-square external-link-icon" aria-hidden="true"></i>`)
	}

	_, _ = bw.WriteString("</a>")

	return ast.WalkContinue, nil
}
