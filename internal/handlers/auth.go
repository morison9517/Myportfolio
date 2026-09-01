// =============================================================================
// auth.go = ログイン・ログアウト・新規登録
//
// ▼ Gin側の受け取り方
//
//	GET  = 画面を見に来た  → HTMLを返す
//	POST = 入力を送ってきた → 中身を照合する
//	同じURLでも、登録するときに GET用/POST用 で別々の関数を指定して分ける。
//
// ▼ 入力の取り出し方
//
//	c.PostForm("username") の "username" は
//	HTML側の <input name="username"> と1対1で対応する。
//	★ここがズレると値が届かない(つまずきの定番)。
//
// =============================================================================
package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"case_gin/internal/database"
	"case_gin/internal/middleware"
	"case_gin/internal/models"
	"case_gin/internal/view"
)

// RegisterAuthRoutes = ログイン関係のURLを登録する。
//
// ▼ Group("/auth") = URLの頭に共通の文字を付けるまとめ役
//
//	下で "/login" と書いたものが、実際には "/auth/login" になる。
//	関連するURLがまとまるので、あとから場所を変えるのも1か所で済む。
func RegisterAuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")

	auth.GET("/register", registerPage)
	auth.POST("/register", registerSubmit)

	auth.GET("/login", loginPage)
	auth.POST("/login", loginSubmit)

	// ★ログアウトが POST な理由
	//   GETでログアウトできると、他サイトに <img src="/auth/logout"> と
	//   書かれただけで勝手にログアウトさせられてしまう。
	//   「状態を変える操作はPOST」が基本ルール。
	auth.POST("/logout", logoutSubmit)
}

// -----------------------------------------------------------------------------
// 新規登録
// -----------------------------------------------------------------------------

func registerPage(c *gin.Context) {
	// もうログインしている人が登録画面に来たら、トップへ返す。
	if middleware.CurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "register.html", view.Page(c, gin.H{
		"Title": "新規登録",
	}))
}

func registerSubmit(c *gin.Context) {
	if middleware.CurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	passwordConfirm := c.PostForm("password_confirm")

	// 問題点をまとめて集めて最後に表示する。
	// 「1個直すと次のエラーが出る」を繰り返させないため。
	var problems []string

	switch {
	case username == "":
		problems = append(problems, "ユーザー名を入力してください。")
	// ★len() は「バイト数」を数えるので、日本語だと1文字=3バイトになる。
	//   文字数を数えたいときは []rune に変換してから len を使う。
	case len([]rune(username)) > 80:
		problems = append(problems, "ユーザー名は80文字以内で入力してください。")
	}

	if len(password) < 8 {
		problems = append(problems, "パスワードは8文字以上で入力してください。")
	}

	if password != passwordConfirm {
		problems = append(problems, "パスワードが一致しません。")
	}

	if username != "" && usernameExists(username) {
		problems = append(problems, "そのユーザー名はすでに使われています。")
	}

	if len(problems) > 0 {
		for _, message := range problems {
			middleware.Flash(c, "error", message)
		}

		// username を渡すので、入力した名前は消えずに残る。
		c.HTML(http.StatusOK, "register.html", view.Page(c, gin.H{
			"Title":    "新規登録",
			"Username": username,
		}))
		return
	}

	user := models.User{Username: username}
	if err := user.SetPassword(password); err != nil {
		middleware.Flash(c, "error", "登録に失敗しました。もう一度お試しください。")
		c.Redirect(http.StatusFound, "/auth/register")
		return
	}

	// ★ここで初めてDBに書き込まれる。
	//   &user の & は「コピーではなく本人を渡す」という意味。
	//   これが無いと、DBが振った番号(ID)が手元の user に反映されない。
	if err := database.DB.Create(&user).Error; err != nil {
		middleware.Flash(c, "error", "登録に失敗しました。もう一度お試しください。")
		c.Redirect(http.StatusFound, "/auth/register")
		return
	}

	_ = middleware.Login(c, user.ID)
	middleware.Flash(c, "success", "ようこそ、"+user.Username+" さん!")
	c.Redirect(http.StatusFound, "/")
}

func usernameExists(username string) bool {
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// -----------------------------------------------------------------------------
// ログイン
// -----------------------------------------------------------------------------

func loginPage(c *gin.Context) {
	if middleware.CurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", view.Page(c, gin.H{
		"Title": "ログイン",
		// ログイン必須ページから飛ばされて来た場合、戻り先を持ち回す。
		"Next": c.Query("next"),
	}))
}

func loginSubmit(c *gin.Context) {
	if middleware.CurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	var user models.User

	// Where(...).First(...) = 条件に合う最初の1件を取る。
	// ★見つからないときもエラーとして返ってくる(gorm.ErrRecordNotFound)。
	//   「エラー = 異常」ではないので、区別して扱う。
	err := database.DB.Where("username = ?", username).First(&user).Error

	// ★「ユーザー名が違います」と書かない理由
	//   どの名前が実在するかを攻撃者に教えてしまうため、あえてぼかす。
	if err != nil || !user.CheckPassword(password) {
		if err != nil && err != gorm.ErrRecordNotFound {
			// DB自体の異常はログに残す(ログインの失敗とは別問題なので)。
			log.Printf("[auth] ログイン時のDBエラー: %v", err)
		}

		middleware.Flash(c, "error", "ユーザー名またはパスワードが正しくありません。")
		c.HTML(http.StatusOK, "login.html", view.Page(c, gin.H{
			"Title":    "ログイン",
			"Username": username,
			"Next":     c.PostForm("next"),
		}))
		return
	}

	_ = middleware.Login(c, user.ID)
	middleware.Flash(c, "success", "ログインしました。")

	c.Redirect(http.StatusFound, safeRedirect(c.PostForm("next")))
}

// safeRedirect = ログイン後の遷移先が安全か確かめる。
//
// ログイン画面のURLには /auth/login?next=/mypage のように戻り先が付く。
// この next を無条件に信じると、?next=https://偽サイト.com というリンクを
// 配られてログイン直後に飛ばされる。自サイト内("/"始まり)だけを許可する。
func safeRedirect(target string) string {
	const fallback = "/"

	if target == "" {
		return fallback
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return fallback
	}

	// 自サイト内なら /mypage の形なので、ドメイン名(Host)は空になる。
	if parsed.Scheme != "" || parsed.Host != "" {
		return fallback
	}

	// "//evil.com" は一見内部リンクに見えるが、外部サイトに飛ぶ書き方なので弾く。
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") {
		return fallback
	}

	return target
}

// -----------------------------------------------------------------------------
// ログアウト
// -----------------------------------------------------------------------------

func logoutSubmit(c *gin.Context) {
	_ = middleware.Logout(c)

	// ★Logout でメモを全部消しているので、
	//   ここで先にメッセージを積んでも一緒に消えてしまう。順番が大事。
	middleware.Flash(c, "success", "ログアウトしました。")

	c.Redirect(http.StatusFound, "/")
}
