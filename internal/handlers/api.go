// =============================================================================
// api.go = JavaScript向けの受付(画面ではなくデータを返す)
//
// ▼ 画面を返す受付との違い
//
//	page.go … HTMLを丸ごと返す。ページが切り替わる。
//	api.go  … データだけ返す。ページを切り替えずに一部だけ書き換えられる。
//
//	使い分けの目安:
//	  ページ移動を伴う操作(詳細ページへ移動、フォーム送信後の遷移) → page.go
//	  その場で追加・削除・切り替え(いいね、フィルタ絞り込み)       → api.go
//
// ▼ JavaScript側との対応
//
//	送受信の作法(整理券を付ける、エラーを拾う)は main.js の api がやるので、
//	JS側は api.post("/api/○○", { ... }) と書くだけでよい。
//
// =============================================================================
package handlers

import (
	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes = JavaScript向けのURLを登録する。
func RegisterAPIRoutes(r *gin.Engine) {
	// r.Group("/api") = 「この先のURLは全部 /api で始まる」というまとめ方。
	// 1つずつ "/api/..." と書かなくて済み、書き間違いも減る。
	api := r.Group("/api")

	// ★ここから書きはじめる
	// api.GET("/posts", listPosts)
	// api.POST("/posts", createPosts)

	_ = api // ★URLを1つでも登録したら、この行は消す
}
