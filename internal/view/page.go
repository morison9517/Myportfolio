// =============================================================================
// page.go = 画面に渡す情報を組み立てる場所
//
// ▼ ★Ginのハマりどころ
//
//	Goのテンプレートは「渡したものしか見えない」。
//	ログイン中の人の名前も、整理券も、メッセージも、渡さなければ画面に出せない。
//	かといって全ページのハンドラに毎回同じ4行を書くのは事故のもと。
//
//	そこで「毎回必要なもの」をここで自動的に詰める。
//	ハンドラ側はそのページ固有の情報だけ書けばよくなる:
//
//	    c.HTML(200, "index.html", view.Page(c, gin.H{
//	        "Title": "ホーム",
//	    }))
//
//	これで画面からは {{ .Title }} も {{ .CurrentUser }} も両方使える。
//
// =============================================================================
package view

import (
	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/middleware"
)

// 設定は起動時に1回だけ受け取って覚えておく。
var appConfig *config.Config

// Setup = テンプレートを準備して、設定を覚える。router.go から呼ばれる。
func Setup(cfg *config.Config, templateDir string) (*Renderer, error) {
	appConfig = cfg

	// 開発中だけ、HTMLを保存するたびに読み込み直す設定にする。
	return NewRenderer(templateDir, !cfg.IsProduction())
}

// Page = 全ページ共通の情報に、そのページ固有の情報を足して返す。
//
// 画面から使える共通の値:
//
//	{{ .SiteName }}    … サイトの表示名
//	{{ .AuthEnabled }} … ログイン機能がONか
//	{{ .CurrentUser }} … 今ログインしている人(未ログインなら空)
//	{{ .CSRFToken }}   … フォームに入れる整理券
//	{{ .Flashes }}     … 「保存しました」などのメッセージ
func Page(c *gin.Context, data gin.H) gin.H {
	if data == nil {
		data = gin.H{}
	}

	// ★プロダクト名が決まったらここを変える(全ページのタイトルとヘッダーに反映される)。
	data["SiteName"] = "Gin Template Demo"

	data["AuthEnabled"] = appConfig.AuthEnabled
	data["CurrentUser"] = middleware.CurrentUser(c)
	data["CSRFToken"] = middleware.CSRFToken(c)
	data["Flashes"] = middleware.TakeFlashes(c)

	return data
}
