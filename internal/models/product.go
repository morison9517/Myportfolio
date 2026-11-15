// =============================================================================
// product.go = 作品(Products)の一覧
//
// ★作品情報の唯一の置き場。トップページ・一覧ページ・詳細ページがすべてここを読む。
//
//	作品を足す・直すときに触るのはこのファイルだけで、HTMLは触らない。
//
// ★DBは使わない(inquiry.go と違い、後から自動で増えるものではないため)。
//
// ▼ 作品を1件足す手順
//
//  1. カード用の画像を web/static/images/products/ に置く
//  2. 下の Products に {...} を1つ足す(並べた順にそのまま画面に出る)
//  3. 詳細ページ(/products/Slug)の中身を Summary 以降に書く。書き方は Chorus が見本
//
// ▼ 詳細ページに写真を載せる手順
//
//  1. web/static/images/products/<Slug>/ というフォルダを作り、そこに置く
//     (例:web/static/images/products/Chorus/home.png)
//     ★横幅は1600px程度まで縮めてから置くこと。スマホの写真をそのまま置くと
//     1枚で数MBになり、ページが重くなる。
//  2. Body の中に Photo("Chorus/home.png", "説明文", "alt") を書く
//
// ▼ 使用技術・インフラのアイコン
//
//	Tool("Go") のように名前だけ書くと、skill.go にある同じ名前のスキルの
//	アイコンを使う(大文字・小文字は区別しない)。
//	skill.go に無いものは Tool("Spotify", "tech/spotify.png") のように画像も書く。
//	画像の位置は web/static/images/ から見た場所。置き場所の目安:
//
//	    使用技術 … web/static/images/tech/
//	    インフラ … web/static/images/infra/aws/s3.svg のように事業者ごとのフォルダ
//
//	★どちらも見つからないときは「?」の画像(no-image.svg)が出る。
//
// =============================================================================
package models

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// Product = 作品1件分。
type Product struct {
	// Slug = URLに使う名前。/products/Chorus の "Chorus" の部分。
	//
	// ★記号は使わないこと("HENRO+" の + は共有時に %2B に化ける)。
	//   画面に出るのは Name のほうなので、見た目は変わらない。
	// ★公開後は変えないこと。変えると、貼られたリンクが404になる。
	// ★大文字・小文字の違いは吸収する(/products/chorus → /products/Chorus へ移す)。
	Slug string

	// Name = 画面に出す作品名。記号はそのまま書いてよい。
	Name string

	// Date = 年月。「2024年3月」のように書く。
	Date string

	// Event = 参加イベント名。個人で作ったものは「個人開発」と書く。
	Event string

	// Award = 受賞。★空にすると、画面側が賞のバッジごと出さない。
	Award string

	// Image = web/static/images/products/ の中のファイル名。
	// ★まだ用意できていない作品は "no-image.svg" にしておく。
	// ★詳細ページのいちばん上と、SNSに貼ったときの画像にも使う
	//   (SVGはSNSで表示されないので、そのときだけサイト共通の ogp.jpg になる)。
	Image string

	// WideLogo = true にすると、画像を切らずに全体を見せ、後ろにシルク色の下地を敷く。
	//
	// ★横長のロゴ(背景が透明なもの)向け。既定(false)は枠いっぱいに広げて上下を切る。
	//   透明なロゴを既定のまま出すと、文字が大きくなりすぎるうえ、
	//   ダークモードでは濃い色の部分が背景に沈んで見えなくなる。
	WideLogo bool

	// Outline = true にすると、画像のまわりに薄い枠線を付ける。
	//
	// ★画像の端がページの背景色に近く、境目が分かりにくいもの向け(Chorus など)。
	//   既定(false)は枠線なし。
	Outline bool

	// HideOnTop = true にすると、トップページの Products には出さない。
	// /products のほうには、この値に関係なく必ず出る。
	// ★既定(false)は「両方に出す」。skill.go の HideOnTop と同じ考え方。
	HideOnTop bool

	// ---------- ここから下は詳細ページ(/products/Slug)だけで使う ----------
	//
	// ★どれも空にすると、画面側がその見出しごと出さない。

	// Summary = 題名の下に出す、1〜2文の概要。
	// ★検索結果やSNSに出る説明文にも使う(120文字を超えたぶんは切る)。
	Summary string

	// Body = 詳細内容。Heading / Subheading / Text / Photo を書いた順に並べる。
	//
	// ★HTMLタグは書けない(そのまま文字として出る)。段落を分けたいときは Text を分ける。
	Body []Block

	// Tech = 使用技術。Tool(...) で書く(書き方はファイル冒頭)。
	Tech []Tech

	// Infra = インフラ。書き方は Tech と同じ。
	Infra []Tech

	// Members = 共同制作者。個人開発なら空のままでよい。
	// ★名前とリンクを載せることは、本人に確認してから。
	Members []Member

	// Links = リポジトリやイベントページへのリンク。
	Links []Link
}

// ---------------------------------------------------------------------------
// 詳細内容(Body)の部品
// ---------------------------------------------------------------------------

// Block = 詳細内容の1かたまり。直接 {...} で書かず、下の4つの関数で作る。
type Block struct {
	// Kind = "heading" / "subheading" / "text" / "photo" のどれか。
	// 画面(product.html)はこれを見て出し分ける。
	Kind string

	// Text = 見出しや段落の文章。
	Text string

	// Image = 写真のファイル名。web/static/images/products/ から見た場所。
	Image string

	// Caption = 写真の下に出す説明文。空なら出さない。
	Caption string

	// Alt = 写真の代わりに読み上げる文章。
	Alt string
}

// Heading = 大きい見出し(例:「搭載した機能」「ハッカソン」)。
func Heading(text string) Block {
	return Block{Kind: "heading", Text: text}
}

// Subheading = Heading の中を区切る小さい見出し(例:「Home」「Search」)。
func Subheading(text string) Block {
	return Block{Kind: "subheading", Text: text}
}

// Text = 段落1つ。
func Text(text string) Block {
	return Block{Kind: "text", Text: text}
}

// Photo = 写真1枚。クリックすると拡大して見られる。
//
//	image   … web/static/images/products/ から見た場所("Chorus/home.png")
//	caption … 写真の下に出す説明文。要らなければ ""
//	alt     … 写真が見えない人向けの説明。★空にしないこと
//	          (caption と同じ文でもよいが、写っているものを書くとより親切)
func Photo(image, caption, alt string) Block {
	return Block{Kind: "photo", Image: image, Caption: caption, Alt: alt}
}

// ---------------------------------------------------------------------------
// 使用技術・インフラ
// ---------------------------------------------------------------------------

// Tech = 使用技術(またはインフラ)1つ分。直接 {...} で書かず、Tool(...) で作る。
type Tech struct {
	Name string

	// Icon = web/static/images/ から見た画像の場所("skills/go.png" など)。
	Icon string
}

// noImageIcon = アイコンが見つからないときに出す「?」の画像。
const noImageIcon = "skills/no-image.svg"

// Tool = 使用技術・インフラを1つ作る。
//
//	Tool("Go")                          … skill.go の「Go」のアイコンを使う
//	Tool("Spotify", "tech/spotify.png") … 画像を指定する
//
// ★アイコンは白い角丸の上に載るので、ダークモード用の画像は要らない。
func Tool(name string, icon ...string) Tech {
	if len(icon) > 0 && icon[0] != "" {
		return Tech{Name: name, Icon: icon[0]}
	}

	return Tech{Name: name, Icon: skillIcon(name)}
}

// skillIcon = skill.go から同じ名前のスキルを探し、そのアイコンの場所を返す。
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
	Name string
	URL  string
}

// Link = 外部ページへのリンク1つ分。
type Link struct {
	Label string
	URL   string
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

// Description = 検索結果やSNSに出す説明文。
//
// ★Summary があればそれ、無ければ Body の最初の段落を使う。どちらも無ければ空
// (そのときはサイト共通の説明文になる)。
func (p Product) Description() string {
	text := p.Summary

	if text == "" {
		for _, block := range p.Body {
			if block.Kind == "text" {
				text = block.Text
				break
			}
		}
	}

	if utf8.RuneCountInString(text) > descriptionLimit {
		text = string([]rune(text)[:descriptionLimit-1]) + "…"
	}

	return text
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

// FindProduct = Slug から作品を1件探す。大文字・小文字は区別しない。
//
// ★見つかった作品の Slug が、渡したものと文字まで同じとは限らない
// ("chorus" で探すと "Chorus" が返る)。正しい形へ移すのはハンドラ側の仕事。
func FindProduct(slug string) (Product, bool) {
	for _, product := range Products {
		if strings.EqualFold(product.Slug, slug) {
			return product, true
		}
	}

	return Product{}, false
}

// TopProducts = トップページに出すぶんだけを返す。HideOnTop: true のものは飛ばす。
// ★並び順は Products のまま。
func TopProducts() []Product {
	products := []Product{}

	for _, product := range Products {
		if product.HideOnTop {
			continue
		}

		products = append(products, product)
	}

	return products
}

// Products = 作品の一覧。★上に書いたものが画面でも上に出る。
var Products = []Product{
	{
		Slug:     "dether",
		Name:     "dether",
		Date:     "2026年10月",
		Event:    "JPHacks2026 HackDay",
		Award:    "Sponsor Award",
		Image:    "dether_logo_light.png",
		WideLogo: true,
		Body: []Block{
			Text("説明はまだ書いていません。"),
		},
		HideOnTop: false,
	},
	{
		Slug:  "Kotas-Portfolio",
		Name:  "Kota's Portfolio",
		Date:  "2026年9月",
		Event: "個人開発",
		Image: "smile.jpg",
		Body: []Block{
			Text("説明はまだ書いていません。"),
		},
		HideOnTop: false,
	},
	{
		Slug:  "Minnano-Hiroba",
		Name:  "みんなの広場",
		Date:  "2025年8月",
		Event: "ニフティ株式会社 プロダクト開発5daysインターン",
		Image: "no-image.svg",
		Body: []Block{
			Text("説明はまだ書いていません。"),
		},
		HideOnTop: false,
	},
	{
		Slug:  "HENRO-plus",
		Name:  "HENRO+",
		Date:  "2024年8月",
		Event: "TwoGate Dev Camp 2024 Summer",
		Image: "HENRO+.svg",
		Body: []Block{
			Text("説明はまだ書いていません。"),
		},
		HideOnTop: true,
	},
	{
		Slug:    "Chorus",
		Name:    "Chorus",
		Date:    "2024年3月",
		Event:   "Kloudハッカソン#4",
		Award:   "デザイン賞",
		Image:   "Chorus-dark.png",
		Outline: true,
		// ★(仮)の付いた文章と写真は、表示を確かめるための見本。差し替えること。
		//   写真は今はカード用の画像を流用している。
		//   本物は web/static/images/products/Chorus/ に置き、"Chorus/home.png" のように書く。
		Summary: "(仮)ここに1〜2文の概要が入ります。",
		Body: []Block{
			Heading("(仮)搭載した機能"),
			Subheading("(仮)Home"),
			Text("(仮)ここに機能の説明が入ります。段落を分けたいときは Text を分けて書きます。"),
			Photo("Chorus-dark.png", "(仮)ホーム画面", "(仮)Chorus のホーム画面のスクリーンショット"),
			Heading("(仮)ハッカソン"),
			Text("(仮)ここにイベントの様子などが入ります。"),
		},
		// ★使用技術・インフラ・リンクは somando.jp/products/chorus の掲載内容に合わせた見本。
		//   Spotify / gunicorn / インフラ3つは画像がまだ無いので「?」になる。
		Tech: []Tech{
			Tool("Django"),
			Tool("SQLite"),
			Tool("Spotify", "tech/spotify.png"),
			Tool("Nginx"),
			Tool("gunicorn", "tech/gunicorn.png"),
		},
		Infra: []Tech{
			Tool("Virtual Machine", "infra/azure/virtual-machine.svg"),
			Tool("S3", "infra/aws/s3.svg"),
			Tool("Route 53", "infra/aws/route53.svg"),
		},
		Members: []Member{
			{Name: "(仮)共同制作者の名前", URL: "https://github.com/"},
		},
		Links: []Link{
			{Label: "GitHub", URL: "https://github.com/Ersteller-team/Chorus"},
			{Label: "Speaker Deck", URL: "https://speakerdeck.com/somando/chorus-kloudhatukason-number-4"},
			{Label: "Kloudハッカソン#4 イベントレポート", URL: "https://kloud.community/Report/kloud-hackathon-4"},
		},
		HideOnTop: false,
	},
}
