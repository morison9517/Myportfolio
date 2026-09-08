// =============================================================================
// product.go = 作品(Products)の一覧
//
// ★作品情報の唯一の置き場。トップページと一覧ページが両方ここを読む。
//
//	作品を足す・直すときに触るのはこのファイルだけで、HTMLは触らない。
//
// ★DBは使わない(inquiry.go と違い、後から自動で増えるものではないため)。
//
// ▼ 作品を1件足す手順
//
//  1. 画像を web/static/images/products/ に置く
//  2. 下の Products に {...} を1つ足す(並べた順にそのまま画面に出る)
//
// =============================================================================
package models

// Product = 作品1件分。
type Product struct {
	// Slug = URLに使う名前。/products/Chorus の "Chorus" の部分。
	//
	// ★詳細ページはまだ無いので、今は404になる。リンク先だけ先にこの形にしてある。
	// ★記号は使わないこと("HENRO+" の + は共有時に %2B に化ける)。
	//   画面に出るのは Name のほうなので、見た目は変わらない。
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
	Image string

	// Body = 詳細ページに出す説明。1つが1段落になる。
	//
	// ★詳細ページを作るまでは、どこにも表示されない。
	// ★HTMLタグは書けない(そのまま文字として出る)。改行は行を分けて書く。
	Body []string
}

// Products = 作品の一覧。★上に書いたものが画面でも上に出る。
var Products = []Product{
	{
		Slug:  "Kotas-Portfolio",
		Name:  "Kota's Portfolio",
		Date:  "2026年9月",
		Event: "個人開発",
		Image: "smile.png",
		Body: []string{
			"説明はまだ書いていません。",
		},
	},
	{
		Slug:  "Minnano-Hiroba",
		Name:  "みんなの広場",
		Date:  "2025年8月",
		Event: "ニフティ株式会社 プロダクト開発5daysインターン",
		Image: "no-image.svg",
		Body: []string{
			"説明はまだ書いていません。",
		},
	},
	{
		Slug:  "HENRO-plus",
		Name:  "HENRO+",
		Date:  "2024年8月",
		Event: "TwoGate Dev Camp 2024 Summer",
		Image: "HENRO+.svg",
		Body: []string{
			"説明はまだ書いていません。",
		},
	},
	{
		Slug:  "Chorus",
		Name:  "Chorus",
		Date:  "2024年3月",
		Event: "Kloudハッカソン#4",
		Award: "デザイン賞",
		Image: "Chorus-dark.png",
		Body: []string{
			"説明はまだ書いていません。",
		},
	},
}
