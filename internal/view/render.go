// =============================================================================
// render.go = 「型紙(base.html)+ 各ページ」を組み合わせてHTMLを作る仕組み
//
// ★このファイルは土台です。基本的に触りません。
//
//	画面を足すときにやることは「web/templates/pages/ にファイルを1枚置く」だけ。
//	ここに登録を書き足す必要はありません(自動で見つけます)。
//
// -----------------------------------------------------------------------------
// ▼ なぜこんな仕組みが必要なのか(Goのテンプレートの事情)
//
//	Goの標準のやり方は「型紙と中身を一緒にセットで読み込む」というもので、
//	全部のHTMLをまとめて1回で読み込むと、こうなる:
//
//	    index.html   … 中身(content)はこれです
//	    login.html   … 中身(content)はこれです  ← 上書き
//	    register.html… 中身(content)はこれです  ← さらに上書き
//
//	同じ名前の「中身」が3つあるので、最後の1つだけが残り、
//	どのURLを開いても新規登録画面が出る、という現象になる。
//
//	そこで「型紙 + そのページ1枚」の組み合わせを、ページの枚数だけ別々に
//	作っておく。index用のセット、login用のセット…と分けておけば混ざらない。
//
//	    セットA = base.html + index.html
//	    セットB = base.html + login.html
//	    セットC = base.html + register.html
//
//	それをやっているのが下の build() です。
//
// =============================================================================
package view

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin/render"
)

// フォルダの役割
//
//	layouts/  … 型紙。全ページ共通のヘッダー・フッターなど(base.html)
//	pages/    … 各ページの中身。ここにファイルを置くと自動でURLに使えるようになる
//	partials/ … 型紙が長くなってきたときに部品を切り出す置き場(最初は空でOK)
const (
	layoutFile  = "layouts/base.html"
	pagesGlob   = "pages/*.html"
	partialGlob = "partials/*.html"
)

// Renderer = ページごとのテンプレートのセットを持っておく係。
type Renderer struct {
	dir   string
	debug bool

	mu   sync.RWMutex
	sets map[string]*template.Template

	// ★書き間違いはページごとに覚えておく。
	//   こうしておくと、誰かが1枚壊しても、他の人のページは普通に表示できる。
	//   (全部まとめて止めると、6人で作業しているとき全員の手が止まる)
	errs map[string]error
}

// funcMap = テンプレートの中から使える自作の関数。
//
// 例えば下のように足すと、HTMLの中で {{ upper .Name }} と書けるようになる。
//
//	"upper": strings.ToUpper,
var funcMap = template.FuncMap{}

// NewRenderer = テンプレートを読み込んで準備する。起動時に1回だけ呼ぶ。
func NewRenderer(dir string, debug bool) (*Renderer, error) {
	r := &Renderer{dir: dir, debug: debug}
	if err := r.build(); err != nil {
		return nil, err
	}
	return r, nil
}

// build = 「型紙 + ページ1枚」のセットを、ページの枚数だけ作る。
func (r *Renderer) build() error {
	layout := filepath.Join(r.dir, filepath.FromSlash(layoutFile))

	// 部品置き場。空でもエラーにはならない。
	partials, err := filepath.Glob(filepath.Join(r.dir, filepath.FromSlash(partialGlob)))
	if err != nil {
		return err
	}

	pages, err := filepath.Glob(filepath.Join(r.dir, filepath.FromSlash(pagesGlob)))
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("ページが1枚もありません: %s", filepath.Join(r.dir, pagesGlob))
	}

	sets := make(map[string]*template.Template, len(pages))
	errs := make(map[string]error)

	for _, page := range pages {
		// "index.html" のようなファイル名で引けるようにしておく。
		// ハンドラ側の c.HTML(200, "index.html", ...) のこの名前と対応する。
		name := filepath.Base(page)

		// 読み込む順番:型紙 → 部品 → そのページ
		files := make([]string, 0, len(partials)+2)
		files = append(files, layout)
		files = append(files, partials...)
		files = append(files, page)

		// ★New("base.html") の名前は、1枚目のファイル名と揃える決まり。
		//   ここを揃えないと「そんなテンプレートは無い」と言われる。
		tmpl, err := template.New(filepath.Base(layout)).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			// このページだけ「エラー」として覚えておき、他のページは作り続ける。
			errs[name] = err
			continue
		}

		sets[name] = tmpl
	}

	r.mu.Lock()
	r.sets = sets
	r.errs = errs
	r.mu.Unlock()

	// 起動時にターミナルにも出しておく(ブラウザを開く前に気づけるように)。
	for name, err := range errs {
		log.Printf("[template] %s に書き間違いがあります: %v", name, err)
	}

	return nil
}

// Instance = Ginがページを表示するたびに呼ぶ関数。
//
// ここが「c.HTML(200, "index.html", data)」の受け口になっている。
func (r *Renderer) Instance(name string, data any) render.Render {
	// ★開発中は毎回読み込み直す。
	//   こうしておくと、HTMLを保存してブラウザを再読み込みするだけで反映される。
	//   (読み込み直さないと、アプリを再起動するまで古い画面が出続ける)
	//   本番では起動時の1回だけなので、この処理は動かない。
	if r.debug {
		if err := r.build(); err != nil {
			return templateError{err}
		}
	}

	r.mu.RLock()
	tmpl, ok := r.sets[name]
	parseErr := r.errs[name]
	r.mu.RUnlock()

	// このページに書き間違いがある場合は、その内容をブラウザに出す。
	// ★ファイル名と行番号が出るので、そこを見れば直せる。
	if parseErr != nil {
		return templateError{parseErr}
	}

	if !ok {
		return templateError{fmt.Errorf(
			"%s が見つかりません。web/templates/pages/ にファイルがあるか確認してください", name)}
	}

	// 型紙(base.html)から描き始める。中身は型紙の中から呼ばれる。
	return render.HTML{Template: tmpl, Name: filepath.Base(layoutFile), Data: data}
}

// -----------------------------------------------------------------------------
// templateError = テンプレートの書き間違いを、ブラウザにそのまま表示する係。
//
//	これが無いとログを見に行かないと原因が分からない。
//	画面に「base.html の23行目」と出れば、その場で直せる。
//
// -----------------------------------------------------------------------------
type templateError struct{ err error }

func (e templateError) Render(w http.ResponseWriter) error {
	e.WriteContentType(w)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, "テンプレートのエラー\n\n%v\n", e.err)
	return nil
}

func (e templateError) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}
