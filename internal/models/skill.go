// =============================================================================
// skill.go = 使える技術(Skills)の一覧
//
// ★スキル情報の唯一の置き場。トップページと /skills の両方がここを読む。
//
//	スキルを足す・直すときに触るのはこのファイルだけで、HTMLは触らない。
//
// ★DBは使わない(product.go と同じ理由。後から自動で増えるものではないため)。
//
// ▼ スキルを1つ足す手順
//
//  1. ロゴを web/static/images/skills/ に置く
//  2. 下の SkillCategories の、入れたいカテゴリの Skills に {...} を1つ足す
//
// トップページ(アイコンだけの一覧)にも自動で増える。
// ★トップの並び順は、このファイルの上から下の順そのまま(TopSkills)。
//
// ★ロゴがまだ無いときは Icon: "no-image.svg"(灰色の「?」)にしておく。
// そのままトップに出すと空き枠が並んで見えるので、
// ロゴを用意するまでは HideOnTop: true も一緒に付けること。
//
// ▼ ★トップページに出す数は6の倍数にすること
//
//	トップの一覧は 6列(スマホ3列 / 広い画面12列)で、余ると最後の行だけ
//	中央に寄って欠けて見える。
//	★ここは「一覧の件数」ではなく「トップに出す件数」。
//	  一覧を増やしてもトップの数を変えたくないときが HideOnTop の出番。
//	詳しくは web/static/css/skills.css と mediaqueries.css。
//
// =============================================================================
package models

// Skill = スキル1つ分。
type Skill struct {
	Name string

	Icon string

	Note string

	// HideOnTop = true にすると、トップページのアイコン一覧には出さない。
	// /skills のほうには、この値に関係なく必ず出る。
	// ★既定(false)は「両方に出す」。逆にすると付け忘れたときに
	//   「トップに出てこない」と悩むことになる。
	HideOnTop bool
}

// SkillCategory = スキルの分類1つ分。/skills のページで見出しごとに区切る単位。
type SkillCategory struct {
	// Name = 見出しに出す英語名。
	Name string

	// Label = 見出しの横に小さく添える日本語。
	Label string

	// Skills = そのカテゴリに入るスキル。★書いた順にそのまま画面に出る。
	Skills []Skill
}

// SkillCategories = スキルの一覧(カテゴリごと)。
//
// ★上に書いたカテゴリが画面でも上に出る。
var SkillCategories = []SkillCategory{
	{
		Name: "Frontend",
		Label: "フロントエンド",
		Skills: []Skill{
			{
				Name: "HTML5",
				Icon: "html5.png",
				Note: "授業の中から個人開発、ハッカソンまで幅広くお世話になっています。構造を意識した開発を行っています。",
			},
			{
				Name: "CSS3",
				Icon: "css3.png",
				Note: "HTMLと共に開発に使用しています。フレームワークは使用していません。",
			},
			{
				Name: "Java Script",
				Icon: "js.png",
				Note: "ページのアニメーション、APIのfetchによく使用しています。",
				HideOnTop: true,
			},
		},
	},
	{
		Name: "Backend",
		Label: "バックエンド",
		Skills: []Skill{
			{
				Name: "Go",
				Icon: "go.png",
				Note: "2025年から独学で学んでいる言語です。このサイトのサーバーも Go で書いています。",
			},
			{
				Name: "Python",
				Icon: "python.png",
				Note: "ハッカソンのWebアプリ開発で使用した経験があります。最初に覚えた言語です。",
			},
			{
				Name: "Java",
				Icon: "java.png",
				Note: "主に授業で学習しています。オブジェクト指向の考え方は、ほかの言語を読むときの土台になっています。",
			},
			{
				Name: "PHP",
				Icon: "php.png",
				Note: "学校の授業で学習。フレームワークを使わずに、簡単な Web アプリを組み立てました。",
			},
		},
	},
	{
		Name: "Framework",
		Label: "フレームワーク",
		Skills: []Skill{
			{
				Name: "Gin",
				Icon: "gin.png",
				Note: "このサイトのバックエンド。ルーティング・テンプレート・ミドルウェアを使っています。",
			},
			{
				Name: "Django",
				Icon: "django.png",
				Note: "Kloudハッカソン#4等で開発に使用しました。ORM とテンプレートの使い方を学びました。",
			},
			{
				Name: "Flask",
				Icon: "flask.png",
				Note: "2026年8月のニフティ株式会社でのインターンシップにて使用しました。",
			},
		},
	},
	{
		Name: "Database",
		Label: "データベース",
		Skills: []Skill{
			{
				Name: "SQLite",
				Icon: "sqlite.png",
				Note: "Kloudハッカソン#4のDjangoデフォルトにて使用しました。",
				HideOnTop: true,
			},
			{
				Name: "MySQL",
				Icon: "mysql.png",
				Note: "学校の授業でのテーブル設計などで使用。本ポートフォリオでも使用しています。",
			},
			{
				Name: "PostgreSQL",
				Icon: "postgre.png",
				Note: "TwoGate Dev Camp 2024 Summerにて使用しました。",
				HideOnTop: true,
			},
		},
	},
	{
		Name: "Infrastructure",
		Label: "インフラ",
		Skills: []Skill{
			{
				Name: "AWS",
				Icon: "aws.png",
				Note: "現在はAWSを主に使用しています。本サイトも EC2 上に Docker で公開しています。",
			},
			{
				Name: "Azure",
				Icon: "azure.png",
				Note: "学生向けクレジットが付与されるので、使ったことがあるインフラストラクチャです。Kloudハッカソン#4にて使用しました。",
				HideOnTop: true,
			},
			{
				Name: "Nginx",
				Icon: "nginx.png",
				Note: "本番でリバースプロキシと HTTPS の終端を担当。画像や CSS の配信もこちらからやります。",
				HideOnTop: true,
			},
		},
	},
	{
		Name: "Git",
		Label: "バージョン管理",
		Skills: []Skill{
			{
				Name: "Git",
				Icon: "git.png",
				Note: "作業ごとにブランチを切って進めています。コミットは目的ひとつ分で区切るよう意識中です。",
				HideOnTop: true,
			},
			{
				Name: "GitHub",
				Icon: "github.png",
				Note: "個人開発のコードの置き場。作ったものを公開する場所としても使っています。",
			},
		},
	},
	{
		Name: "OS",
		Label: "オペレーティングシステム",
		Skills: []Skill{
			{
				Name: "Ubuntu",
				Icon: "ubuntu.png",
				Note: "本サイトを公開しているサーバーのOSとして使用しています。コマンド操作での環境構築もこのタイミングで改めて学びました。",
				HideOnTop: true,
			},
		},
	},
	{
		Name: "Tools",
		Label: "ツール",
		Skills: []Skill{
			{
				// ★HideOnTop はロゴ待ちではなく、トップを12個に保つため。
				Name: "VS Code",
				Icon: "vscode.png",
				Note: "普段のエディタ。拡張機能と整形の設定はリポジトリに入れて、どの環境でも同じにしています。",
				HideOnTop: true,
			},
			{
				// ★HideOnTop はロゴ待ちではなく、トップを12個に保つため。
				Name: "Docker",
				Icon: "docker.png",
				Note: "開発も本番も同じ構成をコンテナで動かしています。Compose で DB と Web をまとめて起動。",
				HideOnTop: true,
			},
		},
	},
}

// TopSkills = トップページのアイコン一覧に出すぶんだけを、
// カテゴリの区切りを外して1本の並びにして返す。
//
//	HideOnTop: true のものは飛ばす。
//	カテゴリの順・その中の順のまま並ぶので、トップもカテゴリごとにまとまって見える。
//
// ★一覧を別に持たず毎回組み立てているのは、SkillCategories を直したときに
// こちらを直し忘れて食い違うのを防ぐため。
// 一覧は十数件なので、作り直しても負担にならない。
func TopSkills() []Skill {
	skills := []Skill{}

	for _, category := range SkillCategories {
		for _, skill := range category.Skills {
			if skill.HideOnTop {
				continue
			}

			skills = append(skills, skill)
		}
	}

	return skills
}
