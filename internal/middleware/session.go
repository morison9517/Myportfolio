// =============================================================================
// middleware = 「全部の通り道に置いておく仕掛け」を集めた場所
//
// ▼ ミドルウェアとは
//
//	お店の入口に置く検温器のようなもの。
//	どのページに行くお客さんも必ずここを通るので、
//	「毎回やること」を1か所にまとめられる。
//
//	    ブラウザ → [セッション] → [ログイン確認] → [成りすまし対策] → 各ページ
//
//	これが無いと、全部のページの先頭に同じコードを書く羽目になる。
//
// -----------------------------------------------------------------------------
// このファイル(session.go)= ブラウザに小さなメモを持たせる仕組み
//
//	HTTPは「1回のやり取りが終わると相手を忘れる」仕組みなので、
//	そのままだとログインしても次のページで他人になってしまう。
//	そこでブラウザに「あなたは3番の人」というメモ(Cookie)を持たせる。
//
//	★メモの中身は SECRET_KEY で署名されている。
//	  中身を書き換えると署名が合わなくなるので、
//	  「自分は1番(管理者)です」と偽装できない。
//
// =============================================================================
package middleware

import (
	"encoding/gob"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	sessionName = "hackapp_session" // ブラウザに保存されるメモの名前
	keyUserID   = "user_id"         // メモの中の「誰か」の欄
)

// FlashMessage = 次の画面に一度だけ出すメッセージ。
//
//	Category … "success" / "error" / "warning"(見た目の色に対応)
//	Message  … 表示する文章
type FlashMessage struct {
	Category string
	Message  string
}

// ★この1行が無いとメッセージの保存時にエラーになる。
//
//	メモには文字や数字しか入れられないので、
//	「FlashMessage という形のものも入れます」と先に登録しておく必要がある。
//	init() は「このファイルが読み込まれたとき自動で1回だけ動く」関数。
func init() {
	gob.Register(FlashMessage{})
}

// Session = メモを扱えるようにする仕掛け。router.go で一番最初に取り付ける。
func Session(secretKey string, secure bool) gin.HandlerFunc {
	// 割り印(SECRET_KEY)を使ってメモに署名する係。
	store := cookie.NewStore([]byte(secretKey))

	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 60 * 60 * 24 * 7, // 7日間。過ぎたら自動でログアウト

		// JavaScriptからメモを読めなくする。
		// 万一ページに悪意のあるスクリプトを埋め込まれても、盗まれない。
		HttpOnly: true,

		// HTTPSのときだけメモを送る。開発中(http)は false でないと動かないので、
		// 本番かどうかで切り替えている。
		Secure: secure,

		// 他サイトからの遷移でメモを送らない(成りすましの入口をふさぐ)。
		SameSite: http.SameSiteLaxMode,
	})

	return sessions.Sessions(sessionName, store)
}

// Login = 「あなたは○番の人」とメモに書き込む。
func Login(c *gin.Context, userID uint) error {
	s := sessions.Default(c)
	s.Set(keyUserID, userID)

	// ★Save() を呼ばないとブラウザに届かない。
	//   「ログインしたのに次のページで未ログインに戻る」の原因はほぼこれ。
	return s.Save()
}

// Logout = メモを丸ごと捨てる。
func Logout(c *gin.Context) error {
	s := sessions.Default(c)

	// Delete ではなく Clear にしている理由:
	// 前の人の情報が1つでも残っていると事故のもとなので、全部消す。
	s.Clear()
	return s.Save()
}

// UserID = メモに書かれている番号を読む。ログインしていなければ 0。
func UserID(c *gin.Context) uint {
	s := sessions.Default(c)

	// ★.(uint) は「中身が uint 型なら取り出す」という書き方。
	//   メモには何でも入れられる形で保存されているので、取り出すときに型を指定する。
	if id, ok := s.Get(keyUserID).(uint); ok {
		return id
	}
	return 0
}

// -----------------------------------------------------------------------------
// flash = 次の画面に一度だけ出すメッセージ
//
//	「保存しました」「パスワードが違います」など。
//	表示する場所は base.html に1か所だけ書いてあるので、
//	各ページで自作する必要はない。
//
//	★メッセージを出す場所と出す画面が違うのがポイント。
//	  「登録処理 → トップページへ移動 → そこで表示」のように、
//	  移動先の画面に持ち越せる。
// -----------------------------------------------------------------------------

// Flash = メッセージを積んでおく。表示は次の画面で自動的に行われる。
//
//	middleware.Flash(c, "success", "保存しました")
func Flash(c *gin.Context, category, message string) {
	s := sessions.Default(c)
	s.AddFlash(FlashMessage{Category: category, Message: message})
	_ = s.Save()
}

// TakeFlashes = 溜まっているメッセージを取り出す(取り出したら消える)。
//
// 画面を作るとき(view.Page)から自動で呼ばれるので、自分で呼ぶことはない。
func TakeFlashes(c *gin.Context) []FlashMessage {
	s := sessions.Default(c)

	raw := s.Flashes()
	if len(raw) == 0 {
		return nil
	}

	// ★取り出しただけでは消えない。Save() して初めてメモから消える。
	//   これを忘れると同じメッセージが毎ページ出続ける。
	_ = s.Save()

	messages := make([]FlashMessage, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(FlashMessage); ok {
			messages = append(messages, m)
		}
	}
	return messages
}
