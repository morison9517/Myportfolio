// =============================================================================
// error.go = エラー画面(400 / 401 / 404 / 500)を返す担当
//
// ▼ ★使い方(各ハンドラから呼ぶ)
//
//	 見つからなかったとき:
//	     handlers.ShowError(c, http.StatusNotFound)
//	     return
//
//	 ★ShowError を呼んだら必ず return すること。
//	   呼んだだけでは処理は止まらないので、後ろの行が動いてしまう。
//
// ▼ 存在しないURLは自動でこの404になる(下の RegisterErrorRoutes)
//
// ▼ 文言を直したい・種類を増やしたいとき
//
//	下の errorPages の表を書き換えるだけ。HTML(error.html)は共通の1枚なので
//	触る必要はない。
//
// =============================================================================
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"case_gin/internal/view"
)

// errorPage = 画面に出す3行。
//
//	Name    … エラーの種類(英語)
//	Message … エラーの内容(日本語)
type errorPage struct {
	Name    string
	Message string
}

// errorPages = 状態番号ごとの文言。★文言を直すのはここ。
var errorPages = map[int]errorPage{
	http.StatusBadRequest: {
		Name:    "Bad Request",
		Message: "リクエストの内容に誤りがあります。",
	},
	http.StatusUnauthorized: {
		Name:    "Unauthorized",
		Message: "このページを表示する権限がありません。",
	},
	http.StatusNotFound: {
		Name:    "Not Found",
		Message: "指定されたページが見つかりませんでした。",
	},
	http.StatusInternalServerError: {
		Name:    "Internal Server Error",
		Message: "サーバー側で問題が発生しました。時間をおいて試してください。",
	},
}

// fallbackErrorPage = 表に無い番号が来たときの文言。
//
// ★表に無いからといって何も返さないと、真っ白な画面になってしまう。
var fallbackErrorPage = errorPage{
	Name:    "Error",
	Message: "問題が発生しました。",
}

// RegisterErrorRoutes = このファイルが担当するURLを登録する。
//
// ★NoRoute = 「どのURLにも当てはまらなかったとき」の受け皿。
//
//	これを登録しておくと、存在しないURL(例:/hoge)を開いた人に
//	Ginの素っ気ない「404 page not found」ではなくこのページが出る。
func RegisterErrorRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		ShowError(c, http.StatusNotFound)
	})
}

// ShowError = エラー画面を返して、以降の処理を止める。
//
// ★JavaScript向けのURL(/api/...)にはHTMLではなくJSONを返す。
//
//	main.js は {"error": "..."} の形を期待しているので、
//	ここでHTMLを返すと画面にエラー内容を出せなくなる。
func ShowError(c *gin.Context, status int) {
	page, ok := errorPages[status]
	if !ok {
		page = fallbackErrorPage
	}

	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatusJSON(status, gin.H{"error": page.Message})
		return
	}

	c.HTML(status, "error.html", view.Page(c, gin.H{
		// ★Title を渡すと <title> が「404 | mrrn.jp」になる。
		"Title": strconv.Itoa(status),

		"ErrorCode":    status,
		"ErrorName":    page.Name,
		"ErrorMessage": page.Message,
	}))

	c.Abort()
}
