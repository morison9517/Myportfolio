// =============================================================================
// auth.go = 「今アクセスしているのは誰か」を毎回調べる仕掛け
//
//	ブラウザのメモには「あなたは3番の人」としか書かれていない。
//	その番号から実際のユーザー情報をDBから取ってくるのがここ。
//
// =============================================================================
package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"case_gin/internal/database"
	"case_gin/internal/models"
)

// この名前で「今の人」を1回のやり取りの間だけ持ち歩く。
const keyCurrentUser = "currentUser"

// LoadUser = 毎回のアクセスで「今の人」を調べて持たせる。
//
// ★ログインしていない人も普通に通す。
//
//	ここは「調べるだけ」の係。通す/通さないの判断は RequireLogin の役目。
func LoadUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := UserID(c)
		if id == 0 {
			// メモが無い = 未ログイン。そのまま次へ。
			c.Next()
			return
		}

		var user models.User
		if err := database.DB.First(&user, id).Error; err != nil {
			// 番号は書いてあるのにDBに居ない(退会した等)。
			// 古いメモを持たせたままだと毎回DBを探しに行くので捨てる。
			_ = Logout(c)
			c.Next()
			return
		}

		c.Set(keyCurrentUser, &user)
		c.Next()
	}
}

// CurrentUser = 今アクセスしている人を取り出す。未ログインなら nil。
//
//	user := middleware.CurrentUser(c)
//	if user != nil { ... }
func CurrentUser(c *gin.Context) *models.User {
	if value, exists := c.Get(keyCurrentUser); exists {
		if user, ok := value.(*models.User); ok {
			return user
		}
	}
	return nil
}

// RequireLogin = 「ログイン中の人だけ通す」ドア。画面を返すページ用。
//
// 会員専用ページを作りたいときは、ルートの登録にこれを挟むだけ:
//
//	r.GET("/mypage", middleware.RequireLogin(), myPage)
func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c) != nil {
			c.Next()
			return
		}

		Flash(c, "warning", "このページを見るにはログインが必要です。")

		// 元々見たかったページを ?next= に付けておくと、
		// ログイン後にそこへ戻してあげられる(処理は handlers/auth.go)。
		next := url.QueryEscape(c.Request.URL.RequestURI())
		c.Redirect(http.StatusFound, "/auth/login?next="+next)

		// ★Abort() が必須。
		//   これが無いと、リダイレクトを出したあとに本来のページの処理も動いてしまう。
		c.Abort()
	}
}

// RequireLoginAPI = 同じく「ログイン中だけ通す」ドアだが、JSONを返すAPI用。
//
// APIでログイン画面にリダイレクトすると、JavaScript側がHTMLを受け取って
// 「JSONとして読めない」と混乱するので、素直に「401 = 未ログイン」を返す。
//
//	api.POST("/todos", middleware.RequireLoginAPI(), createTodo)
func RequireLoginAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c) != nil {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "ログインが必要です。",
		})
	}
}
