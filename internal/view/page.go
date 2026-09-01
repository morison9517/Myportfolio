// =============================================================================
// page.go = 画面に渡す情報を組み立てる場所
//
// ▼ ★Ginのハマりどころ
//
//	Goのテンプレートは「渡したものしか見えない」。
//	整理券も、メッセージも、渡さなければ画面に出せない。
//	かといって全ページのハンドラに毎回同じ数行を書くのは事故のもと。
//
//	そこで「毎回必要なもの」をここで自動的に詰める。
//	ハンドラ側はそのページ固有の情報だけ書けばよくなる:
//
//	    c.HTML(200, "index.html", view.Page(c, gin.H{
//	        "Title": "ホーム",
//	    }))
//
//	これで画面からは {{ .Title }} も {{ .CSRFToken }} も両方使える。
//
// =============================================================================
package view

import (
	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/middleware"
)

// Setup = テンプレートを準備する。router.go から呼ばれる。
func Setup(cfg *config.Config, templateDir string) (*Renderer, error) {
	// 開発中だけ、HTMLを保存するたびに読み込み直す設定にする。
	return NewRenderer(templateDir, !cfg.IsProduction())
}

// Page = 全ページ共通の情報に、そのページ固有の情報を足して返す。
//
// 画面から使える共通の値:
//
//	{{ .SiteName }}  … サイトの表示名
//	{{ .CSRFToken }} … フォームに入れる整理券
//	{{ .Flashes }}   … 「保存しました」などのメッセージ
func Page(c *gin.Context, data gin.H) gin.H {
	if data == nil {
		data = gin.H{}
	}

	// ★サイト名はここ1か所(全ページのタイトルとヘッダーに反映される)。
	data["SiteName"] = "mrrn.jp"

	data["CSRFToken"] = middleware.CSRFToken(c)
	data["Flashes"] = middleware.TakeFlashes(c)

	return data
}
