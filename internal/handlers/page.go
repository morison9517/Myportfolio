// =============================================================================
// page.go = 画面(HTML)を返す受付
//
//	ブラウザ「/ をください」 → ここ → 「index.html をどうぞ」
//
// ▼ ★ファイルを分けている理由(チーム開発でいちばん効く工夫)
//
//	URLの登録を1つのファイルにまとめて書くと、6人が同じ行を編集することになり、
//	git で必ず衝突(コンフリクト)する。
//	そこで「担当ごとにファイルを分けて、自分のファイルの中で自分のURLを登録する」
//	形にしてある。
//
//	    page.go → 画面を返すURL      ← このファイル
//	    auth.go → ログイン関係のURL
//	    api.go  → JavaScript向けのURL
//
//	router.go は、それぞれの登録係を呼ぶだけ。普段は触らない。
//
// ▼ ★トップページはまだ空いています
//
//	今 "/" を開くとデモページが出ますが、それは「ここにまだ "/" が
//	無いから」です。下の見本のように r.GET("/", index) を足せば、
//	自分たちの画面に入れ替わります。
//	デモを消す作業は要りません(仕組みは internal/demo/demo.go)。
//
// =============================================================================
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterPageRoutes = このファイルが担当するURLを登録する。
//
// ★画面を1枚増やすときの手順
//  1. web/templates/pages/ にHTMLを1枚置く
//  2. ここに r.GET("/URL", 関数名) を1行足す
//  3. その関数を下に書く
func RegisterPageRoutes(r *gin.Engine) {
	// ★ここから書きはじめる(コメントを外して、下の index も有効にする)
	// r.GET("/", index)

	r.GET("/health", health)
}

// =============================================================================
// ★トップページの見本(コメントを外して使う)
//
//	view.Page(c, ...) が、ここに書いた情報 + 共通の情報 をまとめてくれる。
//	"index.html" は web/templates/pages/index.html のこと。
//
//	func index(c *gin.Context) {
//		c.HTML(http.StatusOK, "index.html", view.Page(c, gin.H{
//			"Title": "ホーム",
//		}))
//	}
//
//	※ 使うときは上の import に "case_gin/internal/view" を足すこと。
//	  動く見本は internal/demo/demo.go にあります。
//
// =============================================================================

// health = 動作確認用。
//
// gin.H{...} を返すと、Ginが自動でJSONに変換して返す。
//
// AWSやNginxが「アプリが生きているか」を定期的に確認しに来る先。
// 開発中も「画面が出ない…アプリ自体は動いてる?」の切り分けに使える。
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
