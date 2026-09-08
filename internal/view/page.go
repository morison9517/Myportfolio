// =============================================================================
// page.go = 画面に渡す情報を組み立てる場所
//
// ▼ ★Ginのハマりどころ
//
//	Goのテンプレートは「渡したものしか見えない」。
//	整理券も、メッセージも、渡さなければ画面に出せない。
//	かといって全ページのハンドラに毎回同じ数行を書くのは事故のもと。
//
//	そこで「毎回必要なもの」をここで自動的に詰める。
//	ハンドラ側はそのページ固有の情報だけ書けばよくなる:
//
//	    c.HTML(200, "index.html", view.Page(c, gin.H{
//	        "Title": "ホーム",
//	    }))
//
//	これで画面からは {{ .Title }} も {{ .CSRFToken }} も両方使える。
//
// =============================================================================
package view

import (
	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/middleware"
)

// 設定は起動時に1回だけ受け取って、ここに持っておく。
//
// ★Page() は毎回のアクセスで呼ばれるが、引数に *gin.Context しか無い。
//
//	全ハンドラの呼び出しに設定を足して回るより、
//	mailer と同じように起動時に1つ預かるほうが変更が小さい。
var cfg *config.Config

// defaultDescription = 説明文を渡さなかったページで使われる文章。
//
// ★120文字前後に収めること。長いと検索結果で途中から切られる。
const defaultDescription = "Webエンジニアを目指す専門学生 Kota Morito のポートフォリオです。" +
	"Go言語を中心としたバックエンド開発の作品と経歴を掲載しています。"

// siteURL = 「https://ドメイン」までを組み立てる。
//
// ★Nginx を通すと通信は中で http に戻るため、r.TLS だけでは判定できない。
//
//	本物の入口が https だったかは X-Forwarded-Proto に入っている
//	(Nginx が付けている。設定は docker/nginx/prod.conf)。
func siteURL(c *gin.Context) string {
	scheme := "http"

	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return scheme + "://" + c.Request.Host
}

// Setup = テンプレートを準備する。router.go から呼ばれる。
func Setup(c *config.Config, templateDir string) (*Renderer, error) {
	cfg = c

	// 開発中だけ、HTMLを保存するたびに読み込み直す設定にする。
	return NewRenderer(templateDir, !cfg.IsProduction())
}

// Page = 全ページ共通の情報に、そのページ固有の情報を足して返す。
//
// 画面から使える共通の値:
//
//	{{ .SiteName }}          … サイトの表示名
//	{{ .Description }}       … 検索結果やSNSに出る説明文
//	{{ .SiteURL }}           … https://example.com(末尾スラッシュ無し)
//	{{ .CanonicalURL }}      … 今開いているページの正式なURL
//	{{ .CSRFToken }}         … フォームに入れる整理券
//	{{ .Flashes }}           … 「保存しました」などのメッセージ
//	{{ .RecaptchaSiteKey }}  … スパム対策の鍵(未設定なら空)
func Page(c *gin.Context, data gin.H) gin.H {
	if data == nil {
		data = gin.H{}
	}

	// ★サイト名はここ1か所(全ページのタイトルとヘッダーに反映される)。
	data["SiteName"] = "Kota's Portfolio"

	// 検索結果やSNSに出る説明文。
	// ★ページごとに変えたいときは、ハンドラ側で "Description" を渡す。
	if _, ok := data["Description"]; !ok {
		data["Description"] = defaultDescription
	}

	// SNSに貼ったときの画像やURLは、相対パスでは伝わらない。
	// ★ドメインを設定に持たず、アクセスされたURLから組み立てている。
	//   こうしておくと、ドメインを変えても直す場所が無い。
	base := siteURL(c)

	data["SiteURL"] = base
	data["CanonicalURL"] = base + c.Request.URL.Path

	data["CSRFToken"] = middleware.CSRFToken(c)
	data["Flashes"] = middleware.TakeFlashes(c)

	// ★渡すのはサイトキー(画面に埋める鍵)だけ。
	//   シークレットキーは画面に出したら意味が無くなるので絶対に渡さない。
	//   未設定なら空文字が入り、画面側は {{ if }} で読み込みを飛ばす。
	if cfg != nil {
		data["RecaptchaSiteKey"] = cfg.RecaptchaSiteKey
	}

	return data
}
