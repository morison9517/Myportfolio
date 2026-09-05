// =============================================================================
// recaptcha.go = 問い合わせフォームのスパム対策(reCAPTCHA v3)
//
// ▼ 流れ
//
//	画面           Googleから合言葉(トークン)をもらう
//	  ↓ 問い合わせと一緒に送る
//	ここ           そのトークンをGoogleに見せて「本物か」を尋ねる
//	  ↓
//	Googleの返事   本物か + 0.0〜1.0 の点数
//
// ▼ ★v3 は「私はロボットではありません」を出さない
//
//	チェックを押させる代わりに、画面での動き方から点数を付ける。
//	1.0 に近いほど人間らしい。何点で弾くかはこちら側で決める
//	(.env の RECAPTCHA_MIN_SCORE / 既定は 0.5)。
//
// ▼ ★トークンは必ずサーバー側で確かめること
//
//	画面から送られてくる値は、いくらでも書き換えられる。
//	「トークンが入っていた」だけで通すと対策にならない。
//	Googleに問い合わせて初めて意味がある。
//
// ▼ ★設定が無いときは何もしない
//
//	.env の鍵が空なら、確かめずに「よし」と返す。
//	mailer と同じ考え方で、鍵を取る前でもフォームが動くようにしてある。
//
// =============================================================================
package recaptcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"case_gin/internal/config"
)

// Googleの確認窓口。
const verifyURL = "https://www.google.com/recaptcha/api/siteverify"

// 設定は起動時に1回だけ受け取って、ここに持っておく。
// mailer と同じ考え方(アプリ全体で1つを使い回す)。
var cfg *config.Config

// Setup = 設定を渡す。router.New() から呼ばれる。
func Setup(c *config.Config) {
	cfg = c
}

// ★相手はGoogleなので、こちらの都合で待ち続けない。
//
//	時間を切っておかないと、Google側が遅いときに
//	問い合わせの送信画面が固まったままになる。
var client = &http.Client{Timeout: 10 * time.Second}

// ErrFailed = 機械と判断した、という合図。
//
// ★呼び出し側でこれを見分けられるようにしてある。
//
//	通信できなかった(Googleが落ちている等)のと、
//	確かめた結果ダメだったのとでは、対応が変わるため。
var ErrFailed = errors.New("recaptcha: 検証に通りませんでした")

// verifyResponse = Googleからの返事。
//
// ▼ `json:"..."` について
//
//	Googleは error-codes のようにハイフン入りの名前を返してくる。
//	Goの項目名にハイフンは使えないので、この印で対応づける。
type verifyResponse struct {
	Success bool    `json:"success"`
	Score   float64 `json:"score"`
	Action  string  `json:"action"`

	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify = トークンを確かめる。通れば nil を返す。
//
//	token    … 画面が grecaptcha.execute() でもらった合言葉
//	remoteIP … アクセス元(判定の材料。空でもよい)
//	action   … 画面側で指定した行動名。★食い違ったら弾く
//
// ▼ ★action を確かめる理由
//
//	トークンはページのどこで取ったものでも形は同じ。
//	確かめないと、別の場所(例えば点数が甘い画面)で取ったトークンを
//	問い合わせに使い回されてしまう。
func Verify(token, remoteIP, action string) error {
	// 鍵が無いときは素通し(このファイル冒頭の★参照)。
	if cfg == nil || !cfg.RecaptchaEnabled() {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		// ★JavaScriptを切っている利用者もここに来る。
		//   通信する前に終わらせておく(空を送ってもGoogleは弾く)。
		return fmt.Errorf("%w: トークンが空です", ErrFailed)
	}

	form := url.Values{
		"secret":   {cfg.RecaptchaSecretKey},
		"response": {token},
	}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	res, err := client.PostForm(verifyURL, form)
	if err != nil {
		// ★ここは「機械だった」ではなく「確かめられなかった」。
		//   呼び出し側が区別できるよう、ErrFailed では包まない。
		return fmt.Errorf("recaptcha: Googleに問い合わせできません: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("recaptcha: Googleの返事が %d です", res.StatusCode)
	}

	var out verifyResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return fmt.Errorf("recaptcha: 返事を読み取れません: %w", err)
	}

	if !out.Success {
		// ★error-codes には理由が入っている。
		//   invalid-input-secret … RECAPTCHA_SECRET_KEY が違う
		//   timeout-or-duplicate … トークンが古い(2分で切れる)か使い回し
		//   ログに残さないと、鍵の入れ間違いに気づけない。
		log.Printf("[recaptcha] 検証に失敗しました: %v", out.ErrorCodes)
		return fmt.Errorf("%w: %v", ErrFailed, out.ErrorCodes)
	}

	if action != "" && out.Action != action {
		log.Printf("[recaptcha] 行動名が違います(期待: %s / 実際: %s)", action, out.Action)
		return fmt.Errorf("%w: 行動名が違います", ErrFailed)
	}

	if out.Score < cfg.RecaptchaMinScore {
		// ★境目の調整はこのログが頼りになる。
		//   普通の利用者が弾かれるようなら RECAPTCHA_MIN_SCORE を下げる。
		log.Printf("[recaptcha] 点数が足りません(%.2f < %.2f)", out.Score, cfg.RecaptchaMinScore)
		return fmt.Errorf("%w: 点数 %.2f", ErrFailed, out.Score)
	}

	return nil
}
