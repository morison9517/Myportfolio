// =============================================================================
// contact.go = 問い合わせフォームの受付
//
// ▼ 流れ
//
//	画面(#contact)で送信
//	  ↓ contact.js が api.post("/api/contact", {...})
//	ここ
//	  ↓ 1. 中身を確かめる(空、長すぎ、メールの形)
//	  ↓ 2. DBに保存する          ← ★先に保存する
//	  ↓ 3. メールを2通送る(送信者へのコピー / 自分への通知)
//	{"message": "..."} を返す
//
// ▼ ★保存を先にする理由
//
//	メールは相手のサーバー次第で失敗する。
//	先に保存しておけば、メールが届かなくても問い合わせ自体は残り、
//	管理ページ(/admin)で読める。
//	逆にすると「メールが飛ばなかった日の問い合わせが消える」ことになる。
//
// =============================================================================
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/database"
	"case_gin/internal/mailer"
	"case_gin/internal/models"
)

// 入力の上限。DBの列の長さと合わせてある。
//
// ★HTML側の maxlength だけに頼らないこと。
//
//	あれは親切心の機能で、送信そのものは細工すれば通せてしまう。
//	最後に守るのは必ずサーバー側。
const (
	maxNameLength  = 100
	maxEmailLength = 255
	maxBodyLength  = 2000
)

// mailRule = メール本文の区切り線。
//
// ★このファイルと admin.go が出す全てのメールで同じものを使う。
//
//	どれも同じ送信元から届くので、線の形が揃っていないと
//	別々のところから来たように見えてしまう。
const mailRule = "──────────────────────────────"

// contactForm = 画面から送られてくるJSONの形。
type contactForm struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Body  string `json:"body"`
}

// RegisterContactRoutes = このファイルが担当するURLを登録する。
func RegisterContactRoutes(r *gin.Engine, cfg *config.Config) {
	r.POST("/api/contact", func(c *gin.Context) {
		createInquiry(c, cfg)
	})
}

// createInquiry = 問い合わせを受け取る。
func createInquiry(c *gin.Context, cfg *config.Config) {
	var form contactForm

	// ★ShouldBindJSON はJSONの形が違うときだけエラーになる。
	//   中身が空かどうかは見てくれないので、この後で自分で確かめる。
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "送信内容を読み取れませんでした。"})
		return
	}

	// 前後の空白を落としてから確かめる。
	// これが無いと、スペースだけの入力が「入力あり」として通ってしまう。
	name := strings.TrimSpace(form.Name)
	email := strings.TrimSpace(form.Email)
	body := strings.TrimSpace(form.Body)

	if message := validate(name, email, body); message != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	inquiry := models.Inquiry{
		Name:     name,
		Email:    email,
		Body:     body,
		RemoteIP: c.ClientIP(),
	}

	if err := database.DB.Create(&inquiry).Error; err != nil {
		log.Printf("[contact] 保存できません: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存できませんでした。時間をおいて試してください。",
		})
		return
	}

	// ▼ ここから先が失敗しても、問い合わせは保存済みなので成功として返す。
	//   利用者に「送れませんでした」と伝えて再送させるほうが害が大きい
	//   (同じ問い合わせが何件も届くことになる)。
	sendCopyToSender(inquiry)
	sendNoticeToOwner(inquiry, cfg)

	c.JSON(http.StatusOK, gin.H{
		"message": "お問い合わせを受け付けました。ありがとうございます。",
	})
}

// validate = 入力の中身を確かめる。問題があればその文言を返す(無ければ空文字)。
//
// ★文言をそのまま画面に出すので、何を直せばよいか分かる言い方にする。
func validate(name, email, body string) string {
	switch {
	case name == "":
		return "お名前を入力してください。"
	case len([]rune(name)) > maxNameLength:
		return fmt.Sprintf("お名前は%d文字以内で入力してください。", maxNameLength)

	case email == "":
		return "メールアドレスを入力してください。"
	case len(email) > maxEmailLength:
		return fmt.Sprintf("メールアドレスは%d文字以内で入力してください。", maxEmailLength)

	case body == "":
		return "お問い合わせ内容を入力してください。"
	case len([]rune(body)) > maxBodyLength:
		return fmt.Sprintf("お問い合わせ内容は%d文字以内で入力してください。", maxBodyLength)
	}

	// ★メールアドレスの形を確かめる。
	//   ここを通しても「実在する」保証はない(届くかどうかは送ってみるまで
	//   分からない)が、打ち間違いのほとんどはここで気づける。
	if _, err := mail.ParseAddress(email); err != nil {
		return "メールアドレスの形式が正しくありません。"
	}

	// ★ヘッダーに改行を混ぜて別の宛先を足す細工(メールヘッダインジェクション)を防ぐ。
	//   名前とアドレスはメールのヘッダーに入るので、改行が混ざると危ない。
	if strings.ContainsAny(name, "\r\n") || strings.ContainsAny(email, "\r\n") {
		return "入力に使えない文字が含まれています。"
	}

	return ""
}

// sendCopyToSender = 問い合わせた本人に、内容のコピーを送る。
func sendCopyToSender(inquiry models.Inquiry) {
	body := fmt.Sprintf(`%s 様

お問い合わせありがとうございます。
以下の内容で受け付けました。改めてご連絡いたします。

%s
受付日時: %s

お名前:
%s

メールアドレス:
%s

お問い合わせ内容:
%s

%s
※このメールは送信専用のアドレスから自動でお送りしています。
　ご返信いただいても内容を確認できません。
　追加のご連絡は、お手数ですが下記のフォームからお願いいたします。
　https://mrrn.jp/#contact
`,
		inquiry.Name,
		mailRule,
		inquiry.CreatedAt.Format("2006年1月2日 15:04"),
		inquiry.Name,
		inquiry.Email,
		inquiry.Body,
		mailRule,
	)

	err := mailer.Send(mailer.Message{
		To:      inquiry.Email,
		Subject: "お問い合わせを受け付けました",
		Body:    body,
	})
	if err != nil {
		// ★失敗しても利用者には成功として返している(上のコメント参照)。
		//   代わりにここに残しておく。管理ページには問い合わせ自体が出る。
		log.Printf("[contact] コピーを送れません(宛先: %s): %v", inquiry.Email, err)
	}
}

// sendNoticeToOwner = 自分の受信箱に新着を知らせる。
func sendNoticeToOwner(inquiry models.Inquiry, cfg *config.Config) {
	if cfg.ContactNotifyTo == "" {
		return
	}

	body := fmt.Sprintf(`新しいお問い合わせが届きました。

受付日時: %s

お名前:
%s

メールアドレス:
%s

お問い合わせ内容:
%s

%s
このメールに返信はしないでください
一覧: /admin
`,
		inquiry.CreatedAt.Format("2006年1月2日 15:04"),
		inquiry.Name,
		inquiry.Email,
		inquiry.Body,
		mailRule,
	)

	// ★ReplyTo は付けない。
	//
	//   付けると、受信箱で「返信」を押したときの宛先が問い合わせた人になる。
	//   返信は管理ページ(/admin)から行う運用なので、
	//   メールソフトから直接返せてしまうと経路が2つになり、
	//   「返信済み」の印(RepliedAt)も付かないまま話が進んでしまう。
	err := mailer.Send(mailer.Message{
		To:      cfg.ContactNotifyTo,
		Subject: fmt.Sprintf("[お問い合わせ] %s 様より", inquiry.Name),
		Body:    body,
	})
	if err != nil {
		log.Printf("[contact] 自分宛の通知を送れません: %v", err)
	}
}
