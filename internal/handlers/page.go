// =============================================================================
// page.go = 画面(HTML)を返す受付
//
//	ブラウザ「/ をください」 → ここ → 「index.html をどうぞ」
//
// ▼ ★ファイルを分けている理由
//
//	URLの登録を1つのファイルに全部書くと、増えたときに見通しが悪くなる。
//	そこで「種類ごとにファイルを分けて、そのファイルの中でURLを登録する」
//	形にしてある。
//
//	    page.go → 画面を返すURL      ← このファイル
//	    api.go  → JavaScript向けのURL
//
//	router.go は、それぞれの登録係を呼ぶだけ。普段は触らない。
//
// ▼ ★トップページはまだ空いています
//
//	下の見本のように r.GET("/", index) のコメントを外し、
//	web/templates/pages/index.html を置けばトップページが出ます。
//
// =============================================================================
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/models"
	"case_gin/internal/view"
)

// RegisterPageRoutes = このファイルが担当するURLを登録する。
//
// ★画面を1枚増やすときの手順
//  1. web/templates/pages/ にHTMLを1枚置く
//  2. ここに r.GET("/URL", 関数名) を1行足す
//  3. その関数を下に書く
func RegisterPageRoutes(r *gin.Engine) {
	r.GET("/", index)
	r.GET("/skills", skills)
	r.GET("/products", products)
	r.GET("/products/:slug", productDetail)
	r.GET("/privacy", privacy)
	r.GET("/terms", terms)
	r.GET("/health", health)
}

// index = トップページ。
//
//	view.Page(c, ...) が、ここに書いた情報 + 全ページ共通の情報(サイト名・整理券・
//	メッセージ)をまとめてくれる。ページ側は固有の情報だけ書けばよい。
//
//	"index.html" は web/templates/pages/index.html のこと。
//
//	★ここでは Title を渡していない。
//	  渡すと <title> が「ホーム | サイト名」になってしまうので、
//	  トップはサイト名だけにするため、あえて空にしている。
//	  下層ページでは "Title": "Products" のように渡す。
func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", view.Page(c, gin.H{
		// ★カテゴリの区切りを外した1本の並び。
		//   トップの一覧はアイコンを6列で並べるだけなので、区切りは要らない。
		//   ★全部ではない。ロゴ待ちのものなどは skill.go の HideOnTop で外している。
		"Skills": models.TopSkills(),
		// ★こちらも全部ではない。product.go の HideOnTop で外したものは出ない。
		"Products": models.TopProducts(),
		// 年表。中身は models/experience.go。
		"Experiences": models.Experiences,
	}))
}

// skills = 使える技術の一覧ページ(/skills)。
//
//	トップページの Skills はアイコンだけの一覧だが、こちらはカテゴリで区切り、
//	1つずつに一言コメントを付ける。中身の出どころは models/skill.go で共通。
func skills(c *gin.Context) {
	c.HTML(http.StatusOK, "skills.html", view.Page(c, gin.H{
		"Title":           "Skills",
		"Description":     "これまでに学んできた技術の一覧です。フロントエンド・バックエンド・フレームワーク・データベース・インフラの分野ごとにまとめています。",
		"SkillCategories": models.SkillCategories,
	}))
}

// products = 作品の一覧ページ(/products)。中身はトップページの Products と同じ。
//
//	カードを押した先は、下の productDetail(/products/Chorus など)。
func products(c *gin.Context) {
	c.HTML(http.StatusOK, "products.html", view.Page(c, gin.H{
		"Title":       "Products",
		"Description": "これまでに制作した作品の一覧です。ハッカソンやインターンで開発したものを掲載しています。",
		"Products":    models.AllProducts(),
	}))
}

// productDetail = 作品1件の詳細ページ(/products/Chorus など)。
//
//	:slug の部分に作品の Slug(web/content/products/ の .md のファイル名)が入る。
//
//	★.md が無い Slug は404。
//	★大文字・小文字だけ違うとき(/products/chorus)は、正しい形へ移す(301)。
//	  同じページが2つのURLで開けると、検索エンジンが別のページとして数えてしまう。
func productDetail(c *gin.Context) {
	slug := c.Param("slug")

	product, ok := models.FindProduct(slug)
	if !ok {
		ShowError(c, http.StatusNotFound)
		return
	}

	if product.Slug != slug {
		c.Redirect(http.StatusMovedPermanently, "/products/"+product.Slug)
		return
	}

	data := gin.H{
		"Title":   product.Name,
		"Product": product,
		// SNSに貼ったときの画像。空なら base.html が共通の ogp.jpg を使う。
		"OGImage": product.OGImage(),
	}

	// ★説明文は、空のときは渡さない。空文字を渡すと view.Page が
	//   「説明文あり」と見なし、サイト共通の説明文を入れてくれない。
	if description := product.Description(); description != "" {
		data["Description"] = description
	}

	c.HTML(http.StatusOK, "product.html", view.Page(c, data))
}

// privacy = プライバシーポリシー。フッター(base.html)から開く。
//
//	★下層ページは Title を渡す。<title> が「プライバシーポリシー | サイト名」になる。
func privacy(c *gin.Context) {
	c.HTML(http.StatusOK, "privacy.html", view.Page(c, gin.H{
		"Title":       "プライバシーポリシー",
		"Description": "本サイトにおける個人情報の取り扱いについて定めています。",
	}))
}

// terms = 利用規約。作りは privacy と同じ。
func terms(c *gin.Context) {
	c.HTML(http.StatusOK, "terms.html", view.Page(c, gin.H{
		"Title":       "利用規約",
		"Description": "本サイトの利用条件を定めています。",
	}))
}

// health = 動作確認用。
//
// gin.H{...} を返すと、Ginが自動でJSONに変換して返す。
//
// AWSやNginxが「アプリが生きているか」を定期的に確認しに来る先。
// 開発中も「画面が出ない…アプリ自体は動いてる?」の切り分けに使える。
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
