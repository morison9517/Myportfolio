// =============================================================================
// page.go = 画面(HTML)を返す受付
//
//	ブラウザ「/ をください」 → ここ → 「index.html をどうぞ」
//
// ▼ ★ファイルを分けている理由
//
//	URLの登録を1つのファイルに全部書くと、増えたときに見通しが悪くなる。
//	そこで「種類ごとにファイルを分けて、そのファイルの中でURLを登録する」
//	形にしてある。
//
//	    page.go → 画面を返すURL      ← このファイル
//	    api.go  → JavaScript向けのURL
//
//	router.go は、それぞれの登録係を呼ぶだけ。普段は触らない。
//
// ▼ ★トップページはまだ空いています
//
//	下の見本のように r.GET("/", index) のコメントを外し、
//	web/templates/pages/index.html を置けばトップページが出ます。
//
// =============================================================================
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/view"
)

// RegisterPageRoutes = このファイルが担当するURLを登録する。
//
// ★画面を1枚増やすときの手順
//  1. web/templates/pages/ にHTMLを1枚置く
//  2. ここに r.GET("/URL", 関数名) を1行足す
//  3. その関数を下に書く
func RegisterPageRoutes(r *gin.Engine) {
	r.GET("/", index)
	r.GET("/health", health)
}

// index = トップページ。
//
//	view.Page(c, ...) が、ここに書いた情報 + 全ページ共通の情報(サイト名・整理券・
//	メッセージ)をまとめてくれる。ページ側は固有の情報だけ書けばよい。
//
//	"index.html" は web/templates/pages/index.html のこと。
//
//	★ここでは Title を渡していない。
//	  渡すと <title> が「ホーム | mrrn.jp」になってしまうので、
//	  元サイトと同じ「mrrn.jp」のままにするため、あえて空にしている。
//	  下層ページを作るときは "Title": "Products" のように渡す。
func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", view.Page(c, nil))
}

// health = 動作確認用。
//
// gin.H{...} を返すと、Ginが自動でJSONに変換して返す。
//
// AWSやNginxが「アプリが生きているか」を定期的に確認しに来る先。
// 開発中も「画面が出ない…アプリ自体は動いてる?」の切り分けに使える。
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
