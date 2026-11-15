// =============================================================================
// product.go = 作品(Products)の型と、探し方
//
// ★作品の中身はここには書かない。1作品 = 1つの Markdown ファイル。
//
//	web/content/products/Chorus.md
//	web/content/products/dether.md
//	...
//
//	トップページ・一覧ページ・詳細ページは、すべてそこから作られる。
//	作品を足す・直すときに触るのは .md だけで、Go と HTML は触らない。
//	書き方は web/content/products/README.md、見本は Chorus.md。
//
// ▼ ファイルの分担
//
//	product.go          … 作品の型と、探し方(このファイル)
//	product_load.go     … .md を読んで作品の一覧を作る
//	product_markdown.go … 本文の Markdown を HTML にする
//
// ★DBは使わない(inquiry.go と違い、後から自動で増えるものではないため)。
// =============================================================================
package models

import (
	"html/template"
	"net/url"
	"strings"
)

// Product = 作品1件分。.md の先頭の設定欄(--- で囲んだ部分)と本文から作る。
//
// ★yaml:"..." は設定欄での書き方。書き方の説明は web/content/products/README.md。
type Product struct {
	// Slug = URLに使う名前。/products/Chorus の "Chorus" の部分。
	//
	// ★設定欄には書かない。ファイル名(Chorus.md の "Chorus")がそのまま入る。
	// ★記号は使わないこと("HENRO+" の + は共有時に %2B に化ける)。
	// ★公開後はファイル名を変えないこと。変えると、貼られたリンクが404になる。
	Slug string `yaml:"-"`

	// Name = 画面に出す作品名。記号はそのまま書いてよい。
	Name string `yaml:"name"`

	// Date = 画面に出す年月(「2024年3月」)。
	//
	// ★設定欄には date: 2024-03 の形で書く。読み込むときにこの形へ直す。
	//   並び順(新しい順)もこの年月で決まる。
	Date string `yaml:"-"`

	// Event = 参加イベント名。個人で作ったものは「個人開発」と書く。
	Event string `yaml:"event"`

	// Award = 受賞。★空にすると、画面側が賞のバッジごと出さない。
	Award string `yaml:"award"`

	// Image = web/static/images/products/ の中のファイル名。
	// ★まだ用意できていない作品は "no-image.svg" にしておく。
	// ★詳細ページのいちばん上と、SNSに貼ったときの画像にも使う
	//   (SVGはSNSで表示されないので、そのときだけサイト共通の ogp.jpg になる)。
	Image string `yaml:"image"`

	// WideLogo = true にすると、画像を切らずに全体を見せ、後ろにシルク色の下地を敷く。
	//
	// ★横長のロゴ(背景が透明なもの)向け。既定(false)は枠いっぱいに広げて上下を切る。
	//   透明なロゴを既定のまま出すと、文字が大きくなりすぎるうえ、
	//   ダークモードでは濃い色の部分が背景に沈んで見えなくなる。
	WideLogo bool `yaml:"wide_logo"`

	// Outline = true にすると、画像のまわりに薄い枠線を付ける。
	//
	// ★画像の端がページの背景色に近く、境目が分かりにくいもの向け(Chorus など)。
	Outline bool `yaml:"outline"`

	// HideOnTop = true にすると、トップページの Products には出さない。
	// /products のほうには、この値に関係なく必ず出る。
	// ★既定(false)は「両方に出す」。skill.go の HideOnTop と同じ考え方。
	HideOnTop bool `yaml:"hide_on_top"`

	// ---------- ここから下は詳細ページ(/products/Slug)だけで使う ----------
	//
	// ★どれも空にすると、画面側がその見出しごと出さない。

	// Summary = 題名の下に出す、1〜2文の概要。
	// ★検索結果やSNSに出る説明文にも使う(120文字を超えたぶんは切る)。
	Summary string `yaml:"summary"`

	// Tech = 使用技術。Infra = インフラ。書き方は README.md。
	Tech  []Tech `yaml:"tech"`
	Infra []Tech `yaml:"infra"`

	// Members = 共同制作者。個人開発なら書かなくてよい。
	// ★名前とリンクを載せることは、本人に確認してから。
	Members []Member `yaml:"members"`

	// Links = リポジトリやイベントページへのリンク。
	Links []Link `yaml:"links"`

	// BodyHTML = 本文(設定欄より下の Markdown)を HTML にしたもの。
	//
	// ★template.HTML なので、画面側でエスケープされずにそのまま出る。
	//   入れてよいのは product_markdown.go が作ったものだけ。
	//   (Markdown に直接書いた HTML は、そこで取り除いている)
	BodyHTML template.HTML `yaml:"-"`

	// sortKey = 並べ替え用の年月("2024-03")。画面には出さない。
	sortKey string
}

// ---------------------------------------------------------------------------
// 使用技術・インフラ
// ---------------------------------------------------------------------------

// Tech = 使用技術(またはインフラ)1つ分。
type Tech struct {
	Name string

	// Icon = web/static/images/ から見た画像の場所("skills/go.png" など)。
	Icon string
}

// noImageIcon = アイコンが見つからないときに出す「?」の画像。
const noImageIcon = "skills/no-image.svg"

// UnmarshalYAML = 設定欄の tech: / infra: の1行を Tech にする。
//
// 2通りの書き方を受け付ける。
//
//   - Django                                  … skill.go の同じ名前のアイコンを使う
//   - { name: S3, icon: infra/aws/s3.svg }   … 画像を指定する
//
// ★アイコンは白い角丸の上に載るので、ダークモード用の画像は要らない。
func (t *Tech) UnmarshalYAML(unmarshal func(any) error) error {
	var name string
	if err := unmarshal(&name); err == nil {
		*t = Tech{Name: name, Icon: skillIcon(name)}
		return nil
	}

	var entry struct {
		Name string `yaml:"name"`
		Icon string `yaml:"icon"`
	}
	if err := unmarshal(&entry); err != nil {
		return err
	}

	*t = Tech{Name: entry.Name, Icon: entry.Icon}
	if t.Icon == "" {
		t.Icon = skillIcon(entry.Name)
	}

	return nil
}

// skillIcon = skill.go から同じ名前のスキルを探し、そのアイコンの場所を返す。
// 大文字・小文字は区別しない。
//
// ★skill.go の Icon が no-image.svg のもの(ロゴ待ち)は、そのまま「?」になる。
func skillIcon(name string) string {
	for _, category := range SkillCategories {
		for _, skill := range category.Skills {
			if strings.EqualFold(skill.Name, name) {
				return "skills/" + skill.Icon
			}
		}
	}

	return noImageIcon
}

// ---------------------------------------------------------------------------
// 共同制作者・リンク
// ---------------------------------------------------------------------------

// Member = 共同制作者1人分。URL は空でもよい(そのときは名前だけ出す)。
type Member struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Link = 外部ページへのリンク1つ分。
type Link struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

// IconClass = 名前の前に付けるアイコン(Font Awesome のクラス名)。
func (m Member) IconClass() string {
	return linkIcon(m.URL)
}

// IconClass = 名前の前に付けるアイコン(Font Awesome のクラス名)。
func (l Link) IconClass() string {
	return linkIcon(l.URL)
}

// linkIcons = リンク先のドメインごとのアイコン。★無いドメインは地球のアイコンになる。
var linkIcons = map[string]string{
	"github.com":      "fa-brands fa-github",
	"speakerdeck.com": "fa-brands fa-speaker-deck",
	"x.com":           "fa-brands fa-x-twitter",
	"twitter.com":     "fa-brands fa-x-twitter",
	"youtube.com":     "fa-brands fa-youtube",
}

func linkIcon(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")

		if icon, ok := linkIcons[host]; ok {
			return icon
		}
	}

	return "fa-solid fa-globe"
}

// ---------------------------------------------------------------------------
// 詳細ページ用の値
// ---------------------------------------------------------------------------

// descriptionLimit = 検索結果に出す説明文の上限(文字数)。view/page.go の注意書きと同じ。
const descriptionLimit = 120

// Description = 検索結果やSNSに出す説明文(Summary)。
// ★空なら空を返す。そのときはサイト共通の説明文になる。
func (p Product) Description() string {
	text := []rune(p.Summary)

	if len(text) > descriptionLimit {
		return string(text[:descriptionLimit-1]) + "…"
	}

	return string(text)
}

// OGImage = SNSに貼ったときの画像の場所(/static/... の形)。
//
// ★SVGとまだ画像が無いものは空を返す。そのときはサイト共通の ogp.jpg になる。
//
//	X や Facebook は SVG を表示しない。
func (p Product) OGImage() string {
	if p.Image == "" || p.Image == "no-image.svg" || strings.HasSuffix(strings.ToLower(p.Image), ".svg") {
		return ""
	}

	return "/static/images/products/" + p.Image
}

// ---------------------------------------------------------------------------
// 探し方
// ---------------------------------------------------------------------------

// FindProduct = Slug から作品を1件探す。大文字・小文字は区別しない。
//
// ★見つかった作品の Slug が、渡したものと文字まで同じとは限らない
// ("chorus" で探すと "Chorus" が返る)。正しい形へ移すのはハンドラ側の仕事。
func FindProduct(slug string) (Product, bool) {
	for _, product := range AllProducts() {
		if strings.EqualFold(product.Slug, slug) {
			return product, true
		}
	}

	return Product{}, false
}

// TopProducts = トップページに出すぶんだけを返す。HideOnTop: true のものは飛ばす。
// ★並び順は AllProducts と同じ(新しい順)。
func TopProducts() []Product {
	products := []Product{}

	for _, product := range AllProducts() {
		if product.HideOnTop {
			continue
		}

		products = append(products, product)
	}

	return products
}
