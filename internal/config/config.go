// =============================================================================
// config.go = 設定を1か所に集める場所
//
//	.env(金庫) ──読む──> config.go ──渡す──> アプリの各所
//
//	コード内に直接パスワードやDB住所を書くと、変更時に全ファイルを探し回るうえ、
//	GitHubに秘密を上げてしまう。
//
// ▼ package config = このフォルダは config という1つのまとまり、という表札
//
//	他のファイルから import "case_gin/internal/config" と書いて呼び出す。
//	Goでは「1フォルダ = 1パッケージ」で、フォルダ名とパッケージ名を揃えるのが決まり。
//
// ▼ ★Goの大事なルール:大文字で始まる名前だけが外から使える
//
//	Config  → 他のパッケージから使える(公開)
//	loadEnv → このフォルダの中だけ(非公開)
//	「なぜか他のファイルから見えない」はほぼこれが原因。
//
// =============================================================================
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config = アプリ全体の設定をまとめた入れ物。
//
// ▼ struct(構造体)= 項目に名前を付けた入れ物
//
//	「設定」という1つの塊にしておくと、関数に渡すときも引数1個で済む。
type Config struct {
	Env       string // development / production
	Port      string // アプリが待ち受ける番号
	SecretKey string // ブラウザに預けるメモ(Cookie)に署名するときの割り印

	// ▼ ★本番でいちばんハマる設定
	//
	//	true にすると「HTTPSのときだけメモ(Cookie)を持ち歩く」という意味になる。
	//	本番は必ずHTTPSにするので true が正解。
	//
	//	★ただし、まだHTTPSにしていない状態(http:// のまま)で true にすると、
	//	  メモがブラウザに保存されず、フラッシュメッセージなどが消えます。
	//	  エラーも出ないので原因がまず分かりません。
	//
	//	HTTPのまま動かすときだけ .env に
	//	    SECURE_COOKIES=false
	//	と書いて一時的に切る。★HTTPSにしたら必ず true に戻すこと。
	SecureCookies bool

	// ▼ Nginxを前に置くときに必要な設定
	//
	//	Nginxを通すと、Ginから見た相手はNginxになってしまう。
	//	本当のアクセス元は封筒の隅(ヘッダー)に書いてあるので、
	//	「そのヘッダーを信じてよい相手」をここで指定する。
	//
	//	郵便に例えると、転送されてきた手紙の差出人が
	//	「転送してくれた人」になってしまう状態を直す設定です。
	//
	//	★空にすると誰も信用しない(開発中はこちら)。
	//	  信用する相手を書かないまま外に晒すと、利用者がヘッダーを
	//	  詐称してアクセス元を偽れてしまうので、必ず限定すること。
	TrustedProxies []string

	// 利用者が上げたファイル(プロフィールアイコンなど)の保存先。
	//
	//	★web/static と分ける理由
	//	  static … 自分たちが用意したファイル。Gitに入れる。
	//	  media  … 利用者が後から上げたファイル。Gitに入れない。
	//	  static は箱を作り直せば元通りだが、media は消したら戻らない。
	//	  だから本番では media だけを箱の外の保管庫に置く。
	UploadDir string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// ▼ 管理ページ(問い合わせの一覧)に入るための合言葉
	//
	//	利用者の新規登録は無いので、DBのusers表ではなくここで持つ。
	//	★.env はGitHubに上がらないので、DBのパスワードと同じ扱い。
	//	  それでも本番では必ず長いものに変えること。
	AdminEmail    string
	AdminPassword string

	// ▼ 問い合わせフォームのメール送信(SMTP)
	//
	//	独自ドメインのメール(例 contact@mrrn.jp)から送る想定。
	//	契約しているメールサービスの「メールソフトの設定」に載っている
	//	送信サーバー名・ポート・IDとパスワードをそのまま入れる。
	//
	//	★SMTPHost が空のときはメールを送らない(保存だけする)。
	//	  開発中にメールを飛ばしたくないときは空にしておけばよい。
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string

	// MailFrom … 送信元アドレス(例 contact@mrrn.jp)
	// MailFromName … 受信箱に表示される名前(例 mrrn.jp)
	MailFrom     string
	MailFromName string

	// ContactNotifyTo … 新しい問い合わせを知らせる宛先(自分の受信箱)
	ContactNotifyTo string
}

// Load = .env を読んで Config を組み立てて返す。起動時に1回だけ呼ぶ。
//
// ▼ *Config の * について
//
//	Config だと「設定の中身をまるごとコピーして渡す」、
//	*Config だと「設定の置き場所(住所)を渡す」という意味になる。
//	住所を渡せば全員が同じ1つの設定を見るので、コピーのズレが起きない。
func Load() *Config {
	// .env が無くてもエラーにしない。
	// 本番(AWS)では .env ファイルではなく、サーバー側の環境変数を直接使うため。
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env が見つかりませんでした(環境変数を直接使います)")
	}

	cfg := &Config{
		Env:       env("APP_ENV", "development"),
		Port:      env("PORT", "8080"),
		SecretKey: env("SECRET_KEY", "dev-secret-key-change-me"),

		SecureCookies:  envBool("SECURE_COOKIES", true),
		TrustedProxies: envList("TRUSTED_PROXIES"),
		UploadDir:      env("UPLOAD_DIR", "media"),

		DBHost:     env("DB_HOST", "db"),
		DBPort:     env("DB_PORT", "3306"),
		DBName:     env("DB_NAME", "hack_app"),
		DBUser:     env("DB_USER", "hack_user"),
		DBPassword: env("DB_PASSWORD", "hack_password"),

		AdminEmail:    env("ADMIN_EMAIL", ""),
		AdminPassword: env("ADMIN_PASSWORD", ""),

		SMTPHost:     env("SMTP_HOST", ""),
		SMTPPort:     env("SMTP_PORT", "587"),
		SMTPUser:     env("SMTP_USER", ""),
		SMTPPassword: env("SMTP_PASSWORD", ""),

		MailFrom:     env("MAIL_FROM", ""),
		MailFromName: env("MAIL_FROM_NAME", ""),

		ContactNotifyTo: env("CONTACT_NOTIFY_TO", ""),
	}

	// ★本番で割り印が初期値のままだと、メモの中身を偽造される。
	//   起動時に気づけるよう、警告を出す。
	if cfg.IsProduction() && cfg.SecretKey == "dev-secret-key-change-me" {
		log.Println("[config] ⚠ 本番なのに SECRET_KEY が初期値のままです")
	}

	// ★合言葉が空のままだと管理ページに誰も入れない(空の入力も弾く作りなので、
	//   誰でも入れてしまうことはない)。気づけるように知らせておく。
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		log.Println("[config] ADMIN_EMAIL / ADMIN_PASSWORD が未設定です(管理ページに入れません)")
	}

	// ★メールの設定が無いときは、問い合わせをDBに保存するだけになる。
	//   「送ったのに届かない」と悩まないよう、起動時に知らせておく。
	if !cfg.MailEnabled() {
		log.Println("[config] SMTP_HOST / MAIL_FROM が未設定です(問い合わせは保存のみ・メールは送りません)")
	}

	return cfg
}

// MailEnabled = メールを送れる設定がそろっているか。
//
// 送信サーバーと送信元アドレスの両方が無いと送れないので、
// どちらか欠けていたら「送らない」と判断する。
func (c *Config) MailEnabled() bool {
	return c.SMTPHost != "" && c.MailFrom != ""
}

// IsProduction = 本番モードかどうか。
//
// ▼ func (c *Config) ... = Config に付いている関数(メソッド)
//
//	cfg.IsProduction() と書けるようになる。
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// DSN = DBへの接続文字列を組み立てる。読み方:
//
//	ユーザー:パスワード@tcp(住所:ポート)/DB名?オプション
//
// parseTime=true … DBの日時を Go の時刻として受け取る(無いと文字列のまま届く)
// charset=utf8mb4 … 絵文字も扱える指定(utf8 だと絵文字でエラーになる)
// loc=Asia%2FTokyo … 日本時間で扱う(無いと9時間ずれる)
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FTokyo",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// env = 環境変数を読む。無ければ2つ目の値を使う。
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// envBool = .env の "true"/"false" という文字を、Goの true/false に変換する。
//
// .env に書けるのは文字だけなので、"false" をそのまま使うと
// 「中身のある文字 = true」と判定されてしまう。その事故を防ぐ。
// envList = カンマ区切りの環境変数を一覧にする。
//
// 空なら nil を返す。nil は「誰も信用しない」の意味になる。
//
//	TRUSTED_PROXIES=172.18.0.0/16   → ["172.18.0.0/16"]
//	TRUSTED_PROXIES=(未設定)         → nil
func envList(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	var out []string
	for _, part := range strings.Split(value, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
