// =============================================================================
// product.go = 作品(Products)の一覧
//
// ★ここが作品情報の唯一の置き場です。
//
//	トップページ・一覧ページ・詳細ページの3画面が、すべてこの1つを読みます。
//	作品を足す・直すときに触るのはこのファイルだけで、HTMLは触りません。
//
// ▼ inquiry.go との違い
//
//	inquiry.go … DBの表。利用者が送ってきたものを保存する
//	product.go … ここに書いた内容がそのまま画面に出る。DBは使わない
//
//	作品はこちらが書くもので、後から自動で増えたりしない。
//	そのためDBに入れず、プログラムの中に直接書いている。
//
// ▼ ★作品を1件足すときの手順
//
//  1. 画像を web/static/images/products/ に置く
//
//  2. 下の Products に {...} を1つ足す(並べた順にそのまま画面に出る)
//
//     これだけで3画面すべてに反映され、詳細ページのURLも自動で増える。
//
// =============================================================================
package models

// Product = 作品1件分。
type Product struct {
	// Slug = URLに使う名前。/products/Chorus の "Chorus" の部分。
	//
	// ★詳細ページはまだ作っていないので、今は404になる。
	//   カードのリンク先だけ先にこの形にしてある。
	//
	// ★記号は使わないこと。
	//   "HENRO+" の + はURLでは意味を持つ文字で、
	//   共有したときに %2B に化けることがある。
	//   そのため Slug では "HENRO-plus" のように置き換えている。
	//   画面に出るのは下の Name のほうなので、見た目は変わらない。
	Slug string

	// Name = 画面に出す作品名。こちらは記号をそのまま書いてよい。
	Name string

	// Date = 年月。「2024年3月」のように書く。
	Date string

	// Event = 参加イベント名。個人で作ったものは「個人開発」と書く。
	Event string

	// Award = 受賞。無い作品は空にしておく。
	// ★空にすると、画面側が賞のバッジごと出さない。
	Award string

	// Image = web/static/images/products/ の中のファイル名。
	// ★まだ用意できていない作品は "no-image.svg" にしておく。
	Image string

	// Body = 詳細ページに出す説明。1つが1段落になる。
	//
	// ★1行にまとめず、段落ごとに区切って書く。
	//   HTMLの<br>や<p>は書けない(そのまま文字として出てしまう)。
	//   画面側が1つずつ<p>で囲むので、ここでは文章だけを書けばよい。
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
