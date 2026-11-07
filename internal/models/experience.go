// =============================================================================
// experience.go = 経歴(Experience)の年表
//
// ★経歴の唯一の置き場。トップページの Experience がここを読む。
//
//	経歴を足す・直すときに触るのはこのファイルだけで、HTMLは触らない。
//
// ★DBは使わない(product.go と同じく、後から自動で増えるものではないため)。
//
// ▼ 経歴を1件足す手順
//
//  1. 下の Experiences に {...} を1つ足す(並べた順にそのまま左から並ぶ)
//
//     ★左が古く、右が新しい並び。新しい経歴は一番下に足す。
//     ★奇数番目は線の上の段、偶数番目は下の段に自動で並ぶ(experience.css)。
//
// =============================================================================
package models

// Experience = 年表の1件分。
type Experience struct {
	// Date = 年の欄に出す文字。「2024/3」「2025/4~」のように書く。
	Date string

	// Title = 見出し。1つが1行になる(行の間に改行が入る)。
	//
	// ★空にすると見出しごと出さない(Body だけの項目になる)。
	Title []string

	// Body = 説明。1つが1行になる(行の間に改行が入る)。
	//
	// ★product.go の Body と違い「1つ = 1段落」ではなく「1つ = 1行」。
	//   全部が1つの段落にまとまり、行の間に <br> が入る。
	// ★HTMLタグは書けない(そのまま文字として出る)。改行は行を分けて書く。
	// ★空にすると説明ごと出さない。
	Body []string

	// Gap = true にすると、この項目の線に破断マークが入る。
	// 間が長く空くところ(「2005~」のような期間)に使う。
	Gap bool

	// Links = 説明の下に出す外部リンク。無ければ書かなくてよい。
	Links []ExperienceLink
}

// ExperienceLink = 年表の項目に付ける外部リンク1つ分。
type ExperienceLink struct {
	// Label = 画面に出す文字。
	Label string

	// URL = リンク先。
	URL string
}

// Experiences = 年表の一覧。★上に書いたものが画面では左に出る。
var Experiences = []Experience{
	{
		Date:  "2005/9",
		Title: []string{"香川県にて生まれる"},
	},
	{
		Date: "2005~",
		Body: []string{"幼少・小・中学時代"},
		Gap:  true,
	},
	{
		Date:  "2021/4",
		Title: []string{"香川高等専門学校へ入学"},
		Body: []string{
			"高専に入ってIT、通信の業界で働くことを志し、",
			"香川高専　通信ネットワーク工学科へ入学。",
		},
	},
	{
		Date:  "2021~",
		Title: []string{"ITではなく電気の学科だった"},
		Body: []string{
			"自分のやりたいこととは離れたことを専門としており",
			"日々の勉強に徐々についていけなくなる。",
		},
		Gap: true,
	},
	{
		Date:  "2024/3",
		Title: []string{"Kloudハッカソン#4へ参加　デザイン賞を受賞"},
		Body: []string{
			"2週間2人でDjangoフルスタックを担当。",
			"音楽でつながるSNS『Chorus』を開発し",
			"デザイン賞を受賞。",
		},
	},
	{
		Date:  "2024/8",
		Title: []string{"TwoGate Dev Camp 2024 Summerへ参加"},
		Body: []string{
			"3日間2人でDjangoフルスタックを担当。",
			"お遍路を快適にしてくれるLine bot『HENRO+』を開発",
			"大きな経験となった。",
		},
	},
	{
		Date:  "2025/3",
		Title: []string{"香川高等専門学校を中退"},
		Body: []string{
			"2度のハッカソンを通じて自分の進みたい道はITの方であると",
			"再認識。意を決して高専を中退する。",
		},
	},
	{
		Date: "2025/4~",
		Title: []string{
			"穴吹コンピュータカレッジへ入学/在学中",
			"(現在は穴吹ITビジネスカレッジ)",
		},
		Body: []string{
			"専門学校へ再入学し、現在授業や独学で様々な技術を学んでいる。",
		},
	},
	{
		Date:  "2025/8",
		Title: []string{"ニフティ株式会社が開催するインターンシップに参加"},
		Body: []string{
			"ニフティ株式会社にて、プロダクト開発5daysインターンに参加。",
			"Flaskを用いたチーム開発に臨む。",
		},
		Links: []ExperienceLink{
			{Label: "ニフティ株式会社", URL: "https://www.nifty.co.jp/"},
		},
	},
	{
		Date: "2026/10",
		Title: []string{"JPHacks2026 HackDayに参加　Sponsor Award受賞"},
		Body: []string{
			"日本最大級のハッカソンJPHacksに参加(大阪会場)。",
			"6人体制で日常の「音」を集めるアプリ『dether』を開発。",
			"各スポンサー企業からもらえるSponsor Awardをソフトバンク株式会社から受賞した。",
		},
		Links: []ExperienceLink{
			{Label: "JPHacks", URL: "https://jphacks.com/"},
		},
	},
}
