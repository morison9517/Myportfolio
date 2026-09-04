// =============================================================================
// admin.go = 管理ページ(問い合わせの一覧)とそのログイン
//
// ▼ URL
//
//	GET  /admin/login                  … ログイン画面
//	POST /admin/login                  … 合言葉の確認
//	POST /admin/logout                 … ログアウト
//
//	★ここから下はログイン必須(middleware.RequireAdmin)
//	GET  /admin                        … 問い合わせ一覧
//	GET  /admin/inquiries/:id          … 返信画面
//	POST /admin/inquiries/:id/reply    … 返信を送る
//	POST /admin/inquiries/:id/delete   … 問い合わせを消す
//
// ▼ ★利用者の新規登録は無い
//
//	入るのは自分1人だけなので、DBのusers表は作らず
//	.env の ADMIN_EMAIL / ADMIN_PASSWORD と見比べている。
//	将来「会員機能」を作るときは、ここをDBのユーザーに差し替える。
//
// =============================================================================
package handlers

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"case_gin/internal/config"
	"case_gin/internal/database"
	"case_gin/internal/mailer"
	"case_gin/internal/middleware"
	"case_gin/internal/models"
	"case_gin/internal/view"
)

// RegisterAdminRoutes = このファイルが担当するURLを登録する。
func RegisterAdminRoutes(r *gin.Engine, cfg *config.Config) {
	r.GET("/admin/login", adminLoginPage)

	r.POST("/admin/login", func(c *gin.Context) {
		adminLogin(c, cfg)
	})

	r.POST("/admin/logout", adminLogout)

	// ★r.Group の2つ目に middleware.RequireAdmin() を渡すと、
	//   このまとまりの中のURL全部にログイン必須がかかる。
	//   1つずつ書くより、付け忘れの事故が起きない。
	admin := r.Group("/admin", middleware.RequireAdmin())

	admin.GET("", inquiryList)
	admin.GET("/inquiries/:id", inquiryReplyPage)

	admin.POST("/inquiries/:id/reply", func(c *gin.Context) {
		inquiryReply(c, cfg)
	})

	admin.POST("/inquiries/:id/delete", inquiryDelete)
}

// adminLoginPage = ログイン画面を出す。
func adminLoginPage(c *gin.Context) {
	// 既に入っている人をログイン画面に留めても仕方がないので一覧へ送る。
	if middleware.IsAdmin(c) {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	c.HTML(http.StatusOK, "admin_login.html", view.Page(c, gin.H{
		"Title": "ログイン",
	}))
}

// adminLogin = 合言葉を確かめて、合っていればログイン状態にする。
func adminLogin(c *gin.Context, cfg *config.Config) {
	ip := c.ClientIP()

	// ★合言葉を確かめる前に、締め出し中かどうかを見る。
	//   ここを後回しにすると、締め出していても試行そのものは通ってしまい、
	//   総当たりを止められない(仕掛けは loginlimit.go)。
	if wait := loginLockRemaining(ip); wait > 0 {
		minutes := int(wait.Minutes()) + 1

		log.Printf("[admin] 締め出し中のアクセス (%s / あと約%d分)", ip, minutes)
		middleware.Flash(c, "error", fmt.Sprintf(
			"ログインの失敗が続いたため、しばらくお待ちください。(あと約%d分)", minutes))
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if !isAdminCredential(email, password, cfg) {
		recordLoginFailure(ip)

		// ★誰かが試している形跡を残す。
		//   本番でログを見たとき、総当たりされているかどうかが分かる。
		log.Printf("[admin] ログイン失敗 (%s)", ip)

		// ★どちらが違うかは言わない。
		//   「メールアドレスは合っている」と分かると、
		//   総当たりで試す相手に手がかりを与えてしまう。
		message := "メールアドレスかパスワードが違います。"

		if remaining := remainingLoginAttempts(ip); remaining > 0 {
			message += fmt.Sprintf("(あと%d回間違えると、しばらくログインできなくなります)", remaining)
		} else {
			message = fmt.Sprintf(
				"ログインの失敗が続いたため、%d分間ログインできません。",
				int(loginLockDuration.Minutes()))
		}

		middleware.Flash(c, "error", message)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	// ★成功したら失敗の記録を消す。
	//   消さないと「4回間違えて5回目に成功」した次の失敗で締め出される。
	clearLoginFailures(ip)

	if err := middleware.LoginAdmin(c); err != nil {
		log.Printf("[admin] ログイン状態を保存できません: %v", err)
		middleware.Flash(c, "error", "ログインできませんでした。時間をおいて試してください。")
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

// adminLogout = ログアウトしてトップページへ戻す。
func adminLogout(c *gin.Context) {
	if err := middleware.LogoutAdmin(c); err != nil {
		log.Printf("[admin] ログアウトできません: %v", err)
	}

	middleware.Flash(c, "success", "ログアウトしました。")
	c.Redirect(http.StatusFound, "/")
}

// isAdminCredential = 入力が .env の合言葉と一致するか。
//
// ▼ ★設定が空のときは必ず false
//
//	.env を書き忘れたときに「空の入力で誰でも入れる」状態になるのを防ぐ。
//
// ▼ ★単純な == で比べない
//
//	== は違いが見つかった時点で終わるため、返事までの時間差から
//	正解を1文字ずつ推測される可能性がある。
//	ConstantTimeCompare は中身に関係なく同じ時間で比べる。
func isAdminCredential(email, password string, cfg *config.Config) bool {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return false
	}

	// ★メールアドレスの大文字小文字は区別しない(入力揺れで入れなくなるのを防ぐ)。
	emailOK := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(email)),
		[]byte(strings.ToLower(cfg.AdminEmail)),
	) == 1

	passwordOK := subtle.ConstantTimeCompare(
		[]byte(password),
		[]byte(cfg.AdminPassword),
	) == 1

	// ★片方が違っても両方を確かめてから返す。
	//   途中で return すると、そこでも時間差が生まれてしまう。
	return emailOK && passwordOK
}

// inquiryList = 届いた問い合わせを新しい順に並べて出す。
func inquiryList(c *gin.Context) {
	var inquiries []models.Inquiry

	// Order("created_at desc") = 新しいものが上。
	if err := database.DB.Order("created_at desc").Find(&inquiries).Error; err != nil {
		log.Printf("[admin] 問い合わせを読み出せません: %v", err)
		ShowError(c, http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin.html", view.Page(c, gin.H{
		"Title":     "お問い合わせ一覧",
		"Inquiries": inquiries,
	}))
}

// 返信の本文の上限。問い合わせ本文と同じにそろえてある。
const maxReplyLength = 2000

// findInquiry = URLの番号(:id)から問い合わせを1件取り出す。
//
// 見つからなければエラー画面を出して false を返す。
// ★false が返ったら、呼び出し側はそのまま return すること。
//
//	画面はもう出ているので、続けて書くと二重に返事をすることになる。
func findInquiry(c *gin.Context) (models.Inquiry, bool) {
	var inquiry models.Inquiry

	// ★URLの番号は利用者が自由に書き換えられる。
	//   数字以外が来ることを前提にして受け止める。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		ShowError(c, http.StatusNotFound)
		return inquiry, false
	}

	if err := database.DB.First(&inquiry, id).Error; err != nil {
		// ★「無い」と「DBの故障」は分けて扱う。
		//   まとめて404にすると、DBが壊れていることに気づけない。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ShowError(c, http.StatusNotFound)
		} else {
			log.Printf("[admin] 問い合わせを読み出せません: %v", err)
			ShowError(c, http.StatusInternalServerError)
		}
		return inquiry, false
	}

	return inquiry, true
}

// inquiryReplyPage = 1件の問い合わせと、返信の入力欄を出す。
func inquiryReplyPage(c *gin.Context) {
	inquiry, ok := findInquiry(c)
	if !ok {
		return
	}

	// ★返信済みなら「読むだけ」の画面になる(出し分けは admin_reply.html)。
	//   見出しも変える。
	title := "返信"
	if inquiry.RepliedAt != nil {
		title = "返信内容"
	}

	c.HTML(http.StatusOK, "admin_reply.html", view.Page(c, gin.H{
		"Title":   title,
		"Inquiry": inquiry,
	}))
}

// inquiryReply = 返信メールを送る。
func inquiryReply(c *gin.Context, cfg *config.Config) {
	inquiry, ok := findInquiry(c)
	if !ok {
		return
	}

	// ★返信済みのものには送らせない。
	//
	//   画面は「読むだけ」になっていて送信ボタンが無いが、
	//   URLを直接叩けばここまで来られる。
	//   通してしまうと、控えてある返信内容が上書きされて消える。
	if inquiry.RepliedAt != nil {
		middleware.Flash(c, "warning", "このお問い合わせには既に返信済みです。")
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	body := strings.TrimSpace(c.PostForm("body"))

	// ★入力に問題があるときは、書いた内容を持ったまま画面に戻す。
	//   リダイレクトすると入力が消えて、最初から書き直しになる。
	if body == "" || len([]rune(body)) > maxReplyLength {
		message := "返信内容を入力してください。"
		if body != "" {
			message = fmt.Sprintf("返信内容は%d文字以内で入力してください。", maxReplyLength)
		}

		middleware.Flash(c, "error", message)
		c.HTML(http.StatusOK, "admin_reply.html", view.Page(c, gin.H{
			"Title":     "返信",
			"Inquiry":   inquiry,
			"ReplyBody": body,
		}))
		return
	}

	err := mailer.Send(mailer.Message{
		To:      inquiry.Email,
		Subject: "お問い合わせへのご返信",
		Body:    buildReplyBody(inquiry, body),
	})
	if err != nil {
		log.Printf("[admin] 返信を送れません(宛先: %s): %v", inquiry.Email, err)

		middleware.Flash(c, "error", "返信を送信できませんでした。時間をおいて試してください。")
		c.HTML(http.StatusOK, "admin_reply.html", view.Page(c, gin.H{
			"Title":     "返信",
			"Inquiry":   inquiry,
			"ReplyBody": body,
		}))
		return
	}

	// ★メールの設定が無いときは mailer が「送らずに成功」を返す。
	//   そのまま「送信しました」と出すと嘘になるので、正直に伝える。
	if !cfg.MailEnabled() {
		middleware.Flash(c, "warning",
			"メールの設定が無いため、実際には送信されていません(内容はログに出力しました)。")

		// ★実際に送っていないので「返信済み」にはしない。
		//   ここで印を付けると、届いていない相手を返信済みとして
		//   見落とすことになる。
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	// 「返信済み」の印と、送った内容の控えを残す。
	//
	// ★Updates で必要な列だけ書き換える。
	//   Save だと全部の列を書き直すので、その間に別の場所から
	//   変更があったときに上書きしてしまう。
	//
	// ★控え(reply_body)を残すのは、管理ページで「何と答えたか」を
	//   後から読み返せるようにするため。メールは送ったら手元に残らない。
	now := time.Now()
	if err := database.DB.Model(&inquiry).Updates(map[string]any{
		"replied_at": now,
		"reply_body": body,
	}).Error; err != nil {
		// ★ここで失敗してもメールは届いている。
		//   利用者に「送れませんでした」と伝えると二重に送ることになるので、
		//   印が付かなかったことだけをログに残す。
		log.Printf("[admin] 返信済みの印を付けられません(id: %d): %v", inquiry.ID, err)
	}

	middleware.Flash(c, "success", fmt.Sprintf("%s 様に返信を送信しました。", inquiry.Name))
	c.Redirect(http.StatusFound, "/admin")
}

// buildReplyBody = 返信メールの本文を組み立てる。
//
//	お礼のあいさつ
//	返信の内容
//	====== の区切り
//	いただいた問い合わせのコピー
//
// ★問い合わせのコピーを付けているのは、受け取った人が
//
//	「何についての返事か」を思い出せるようにするため。
func buildReplyBody(inquiry models.Inquiry, reply string) string {
	return fmt.Sprintf(`%s 様

お問い合わせいただきありがとうございます。
以下のとおりご回答いたします。

======================================================================

%s

======================================================================
以下は、いただいたお問い合わせの内容です。

受付日時: %s

お名前:
%s

メールアドレス:
%s

お問い合わせ内容:
%s

----------------------------------------------------------------------
※このメールは送信専用のアドレスからお送りしています。
　ご返信いただいても内容を確認できません。
　追加のご質問は、お手数ですが下記のフォームからお願いいたします。
　https://mrrn.jp/#contact
`,
		inquiry.Name,
		reply,
		inquiry.CreatedAt.Format("2006年1月2日 15:04"),
		inquiry.Name,
		inquiry.Email,
		inquiry.Body,
	)
}

// inquiryDelete = 問い合わせを1件消す。
//
// ★消す前の確認は画面側(admin.js のポップアップ)で行っている。
//
//	ただしURLを直接叩けばここに来られるので、
//	「確認したか」をサーバー側であてにしてはいけない。
//	消してよいのはログイン済みの人だけ、という点だけが本当の守りになる
//	(それは RequireAdmin が担当している)。
func inquiryDelete(c *gin.Context) {
	inquiry, ok := findInquiry(c)
	if !ok {
		return
	}

	if err := database.DB.Delete(&inquiry).Error; err != nil {
		log.Printf("[admin] 問い合わせを消せません: %v", err)
		middleware.Flash(c, "error", "削除できませんでした。時間をおいて試してください。")
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	middleware.Flash(c, "success", fmt.Sprintf("%s 様のお問い合わせを削除しました。", inquiry.Name))
	c.Redirect(http.StatusFound, "/admin")
}
