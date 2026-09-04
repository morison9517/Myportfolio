// =============================================================================
// auth.go = 「ログインしている人だけ通す」ドア
//
//	ログイン状態そのものは session.go が持っている。
//	ここはそれを見て、通すか追い返すかを決めるだけ。
//
// ▼ 使い方(ルートの登録に挟むだけ)
//
//	admin := r.Group("/admin", middleware.RequireAdmin())
//	admin.GET("", inquiryList)
//
// =============================================================================
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin = ログイン中の人だけ通す。画面を返すページ用。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) {
			c.Next()
			return
		}

		Flash(c, "warning", "このページを見るにはログインが必要です。")
		c.Redirect(http.StatusFound, "/admin/login")

		// ★Abort() が必須。
		//   これが無いと、追い返したあとに本来のページの処理も動いてしまう。
		c.Abort()
	}
}
