// =============================================================================
// mailer.go = メールを送る係
//
// ▼ 使い方
//
//	mailer.Send(mailer.Message{
//	    To:      "someone@example.com",
//	    Subject: "お問い合わせありがとうございます",
//	    Body:    "……",
//	    ReplyTo: "contact@example.com", // 省略可
//	})
//
// ▼ ★設定が無いときは何もしない
//
//	.env の SMTP_HOST / MAIL_FROM が空なら、送らずに「送っていません」と
//	ログに出すだけで終わる(エラーにはしない)。
//	開発中にメールを飛ばさないための逃げ道であり、
//	問い合わせ自体はDBに保存されているので失われない。
//
// ▼ ★なぜ外部ライブラリを使わないのか
//
//	Goには net/smtp が最初から入っていて、これだけで送れる。
//	送る量が少ないうちは追加のライブラリは要らない。
//
// =============================================================================
package mailer

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	netmail "net/mail"
	"net/smtp"
	"strings"
	"time"

	"case_gin/internal/config"
)

// 設定は起動時に1回だけ受け取って、ここに持っておく。
// database.DB と同じ考え方(アプリ全体で1つを使い回す)。
var cfg *config.Config

// Setup = 設定を渡す。router.New() から呼ばれる。
func Setup(c *config.Config) {
	cfg = c
}

// Message = 送るメール1通分。
type Message struct {
	To      string
	Subject string
	Body    string

	// ReplyTo = 受け取った人が「返信」を押したときの宛先。
	//
	//	空なら送信元(MAIL_FROM)に返る。
	//	問い合わせの通知では、ここに問い合わせた人のアドレスを入れておくと、
	//	自分の受信箱から「返信」を押すだけでその人に返事が書ける。
	ReplyTo string
}

// Send = メールを1通送る。
//
// ★設定が無いときは nil(成功扱い)を返す。
//
//	呼び出し側が毎回「設定はあるか?」を気にしなくて済むようにしている。
func Send(m Message) error {
	if cfg == nil || !cfg.MailEnabled() {
		log.Printf("[mail] 設定が無いので送りません(宛先: %s / 件名: %s)", m.To, m.Subject)
		return nil
	}

	if strings.TrimSpace(m.To) == "" {
		return errors.New("宛先が空です")
	}

	addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)

	// ★465番は「最初から暗号化して繋ぐ」方式(SSL/TLS)。
	//   587番は「普通に繋いでから暗号化に切り替える」方式(STARTTLS)。
	//   契約先の案内でどちらを使うか決まるので、両方に対応しておく。
	client, err := dial(addr, cfg.SMTPHost, cfg.SMTPPort == "465")
	if err != nil {
		return fmt.Errorf("送信サーバーに繋げません: %w", err)
	}
	defer client.Close()

	if cfg.SMTPPort != "465" {
		// STARTTLS に対応していれば暗号化に切り替える。
		// ★暗号化しないとIDとパスワードがそのまま流れる。
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: cfg.SMTPHost}); err != nil {
				return fmt.Errorf("暗号化に切り替えられません: %w", err)
			}
		}
	}

	if cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("ログインできません(IDかパスワードを確認してください): %w", err)
		}
	}

	if err := client.Mail(cfg.MailFrom); err != nil {
		return fmt.Errorf("送信元が受け付けられません: %w", err)
	}
	if err := client.Rcpt(m.To); err != nil {
		return fmt.Errorf("宛先が受け付けられません: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(build(m)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return err
	}

	// ★成功したことも残しておく。
	//   これが無いと「送ったのに届かない」ときに、
	//   アプリが送ったのか送っていないのかが切り分けられない。
	log.Printf("[mail] 送信しました(宛先: %s / 件名: %s)", m.To, m.Subject)
	return nil
}

// dial = 送信サーバーに繋ぐ。
//
// ★時間制限を付けているのが要点。
//
//	付けないと、相手のサーバーが黙り込んだときに延々と待ち続け、
//	問い合わせフォームの「送信中…」が終わらなくなる。
func dial(addr, host string, useTLS bool) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	// 繋いだ後のやり取り全体にも制限を付ける。
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	if useTLS {
		conn = tls.Client(conn, &tls.Config{ServerName: host})
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

// build = メールの中身(ヘッダー + 本文)を組み立てる。
//
// ▼ ★日本語をそのままヘッダーに書いてはいけない
//
//	件名や差出人名の欄は、決められた形に変換しないと
//	受信側で文字化けする(本文は Content-Type で指定できるので大丈夫)。
//	mime.QEncoding がその変換をしてくれる。
//
// ▼ ★改行は \n ではなく \r\n
//
//	メールの決まりごと。\n だけだと受け付けないサーバーがある。
func build(m Message) []byte {
	var b strings.Builder

	// ★差出人の欄は自分で組み立てない。
	//
	//   以前は「名前 <アドレス>」と文字をつなげていたが、
	//   名前に . や , などの記号が入ると、決まりに反する書き方になる。
	//   例) mrrn.jp <contact@mrrn.jp>
	//       → 記号を含む名前は "…" で囲まなければならない
	//   受け取り側によっては、この形を拒否したり黙って捨てたりする。
	//
	//   net/mail の Address が、囲む必要の判断も、日本語の変換も
	//   まとめてやってくれる。
	from := (&netmail.Address{
		Name:    cfg.MailFromName,
		Address: cfg.MailFrom,
	}).String()

	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + m.To + "\r\n")

	if m.ReplyTo != "" {
		b.WriteString("Reply-To: " + m.ReplyTo + "\r\n")
	}

	b.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", m.Subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")

	// ★本文は base64 に変換して送る。
	//
	//   日本語をそのままの形(8bit)で流すこともできるが、それには
	//   相手のサーバーが 8BITMIME に対応している必要がある。
	//   対応していないサーバーは「受け取りました」と答えたうえで
	//   中身を壊したり捨てたりすることがあり、
	//   送った側からは成功に見えるのに届かない、という状態になる。
	//
	//   base64 は英数字と記号だけになるので、どのサーバーでも安全に通る。
	b.WriteString("Content-Transfer-Encoding: base64\r\n")

	// ヘッダーと本文の区切りは空行1つ。
	b.WriteString("\r\n")

	// 本文の改行も \r\n にそろえる。
	body := strings.ReplaceAll(m.Body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\n", "\r\n")

	b.WriteString(wrapBase64(base64.StdEncoding.EncodeToString([]byte(body))))

	return []byte(b.String())
}

// wrapBase64 = base64 の長い1行を76文字ごとに折り返す。
//
// ★メールの1行は1000文字までという決まりがある。
//
//	長い問い合わせを変換すると簡単に超えるので、必ず折る。
//	76文字は昔からの慣習の値。
func wrapBase64(encoded string) string {
	const lineLength = 76

	var b strings.Builder

	for len(encoded) > lineLength {
		b.WriteString(encoded[:lineLength])
		b.WriteString("\r\n")
		encoded = encoded[lineLength:]
	}

	b.WriteString(encoded)
	b.WriteString("\r\n")

	return b.String()
}
