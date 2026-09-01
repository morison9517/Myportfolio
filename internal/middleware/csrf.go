// =============================================================================
// csrf.go = 成りすまし送信を防ぐ仕掛け
//
// ▼ 何を防いでいるのか
//
//	うちのサイトにログインしたまま、別のサイトの罠を踏んだとする。
//	罠のページに「うちのサイトへ送信するフォーム」が仕込まれていると、
//	ブラウザはログインのメモを一緒に送ってしまうので、
//	本人が押したことになって勝手に投稿・削除されてしまう。
//
// ▼ 対策
//
//	フォームに「使い捨ての整理券」を埋め込み、整理券が無い依頼は受け付けない。
//	整理券はうちのページを開かないと手に入らないので、罠のページからは作れない。
//
// ▼ 開発側でやること
//
//	フォーム …… <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
//	JavaScript … main.js の api.post() を使う(整理券が自動で付く)
//
// =============================================================================
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const keyCSRFToken = "csrf_token"

// CSRFToken = この人の整理券を返す(まだ無ければ発行する)。
//
// 画面を作るとき(view.Page)から自動で呼ばれるので、自分で呼ぶことはない。
func CSRFToken(c *gin.Context) string {
	s := sessions.Default(c)

	if token, ok := s.Get(keyCSRFToken).(string); ok && token != "" {
		return token
	}

	token := randomToken()
	s.Set(keyCSRFToken, token)
	_ = s.Save()
	return token
}

// CSRF = 整理券を確認する仕掛け。router.go で全ページに取り付ける。
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 「見るだけ」の依頼は何も変えないので確認しない。
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		expected := CSRFToken(c)

		// JavaScriptからはヘッダーで、HTMLのフォームからは隠し項目で送られてくる。
		// ★ヘッダーを先に見ている理由:
		//   PostForm を先に呼ぶと送信内容を読み進めてしまい、
		//   後続の処理(JSONの読み取り)が空になることがある。
		sent := c.GetHeader("X-CSRF-Token")
		if sent == "" {
			sent = c.PostForm("csrf_token")
		}

		// ★単純な == で比べない。
		//   == は違いが見つかった時点で終わるため、returnまでの時間差から
		//   正解を1文字ずつ推測される可能性がある。長さに関係なく同じ時間で比べる。
		if subtle.ConstantTimeCompare([]byte(expected), []byte(sent)) != 1 {
			rejectCSRF(c)
			return
		}

		c.Next()
	}
}

// rejectCSRF = 整理券が無い/違うときの返事。
func rejectCSRF(c *gin.Context) {
	// APIならJSONで返す。HTMLを返すとJavaScript側が読めずに混乱する。
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/__demo/api/") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "整理券(CSRFトークン)が正しくありません。ページを再読み込みしてください。",
		})
		return
	}

	c.String(http.StatusBadRequest,
		"整理券(CSRFトークン)が正しくありません。\n"+
			"フォームに次の1行が入っているか確認してください:\n"+
			`<input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">`)
	c.Abort()
}

// randomToken = 推測できないランダムな文字列を作る。
func randomToken() string {
	buf := make([]byte, 32)

	// crypto/rand = 予測できない乱数。
	// ★math/rand ではダメ。あちらは高速だが規則性があり、次の値を当てられる。
	if _, err := rand.Read(buf); err != nil {
		// ここが失敗するのはOS側の異常。安全側に倒して止める。
		panic("乱数を作れませんでした: " + err.Error())
	}

	return base64.RawURLEncoding.EncodeToString(buf)
}
