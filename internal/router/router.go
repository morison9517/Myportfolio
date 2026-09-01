// =============================================================================
// router.go = URLの受付をまとめて組み立てる場所
//
// ★このファイルは土台です。普段は触りません。
//
//	URLを増やしたいときに触るのは handlers/ の中のファイルです。
//	ここを触るのは「新しい担当ファイルを増やしたとき」だけ(下に2行足す)。
//
// ▼ 上から順に「通り道」を作っていくイメージ
//
//	ブラウザ
//	  ↓
//	[静的ファイル] CSS・JS・画像はここで直接返す(下の仕掛けを通らない)
//	  ↓
//	[セッション]   ブラウザのメモを読めるようにする
//	  ↓
//	[成りすまし対策] 送信に整理券が付いているか確認する
//	  ↓
//	[各ページ]     handlers/ の中の関数
//
// =============================================================================
package router

import (
	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/handlers"
	"case_gin/internal/middleware"
	"case_gin/internal/view"
)

// ファイルの置き場所。プロジェクトの入口(compose.ymlのある場所)から見た位置。
const (
	templateDir = "web/templates"
	staticDir   = "web/static"
)

// New = アプリの受付を組み立てて返す。
func New(cfg *config.Config) (*gin.Engine, error) {
	// 本番では起動ログを静かにし、デバッグ用の警告も出さないモードにする。
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.Default() = 「アクセスの記録」と「異常時に落ちない仕掛け」が
	// 最初から付いた受付。
	// 異常時に落ちない仕掛けがあるので、1か所でエラーが出てもアプリ全体は止まらない。
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// アップロードの受け皿サイズ。
	// ★本当の上限は本番のNginx側でも設定する(でないと巨大ファイルで詰まる)。
	r.MaxMultipartMemory = 16 << 20 // 16MB

	// ★「アクセス元のIPをどこまで信用するか」の設定。
	//
	//   Nginxを前に置くと、Ginから見た相手はNginxになってしまう。
	//   本当のアクセス元はヘッダーに書いてあるが、それを無条件で信じると
	//   利用者が自分でヘッダーを詐称してIPを偽れてしまう。
	//   そこで「このアドレスから来たヘッダーなら信じてよい」を指定する。
	//
	//   開発中は空(誰も信用しない)。本番は .env の TRUSTED_PROXIES に
	//   Nginxのアドレス範囲を書く(手順は docs/DEPLOY.md)。
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}

	// --- 画面の型紙(base.html)を使えるようにする ---
	renderer, err := view.Setup(cfg, templateDir)
	if err != nil {
		return nil, err
	}
	r.HTMLRender = renderer

	// --- CSS / JS / 画像 ---
	// ★ここを Use より先に書いている理由
	//   Ginは「Useより後に登録したURL」にだけ仕掛けを適用する。
	//   先に書くことで、画像1枚読むたびに共通の仕掛けが走るのを避けられる。
	r.Static("/static", staticDir)

	// --- サイト直下の /favicon.ico ---
	// ★HTMLを返さないURL(/health のようなJSON)や、<link rel="icon"> を読む前の
	//   ブラウザは、最後の手段としてサイト直下の /favicon.ico を取りに来る。
	//   ここで応答しておくと、どのURLでもタブにアイコンが出る。
	r.StaticFile("/favicon.ico", staticDir+"/tabicon.png")

	// --- 利用者が上げたファイル(プロフィールアイコンなど) ---
	// ★開発モードのときだけ、Ginが自分で画像を配る。
	//   本番ではNginxが配るので登録しない(compose.prod.yml 参照)。
	if !cfg.IsProduction() {
		r.Static("/media", cfg.UploadDir)
	}

	// --- 全ページ共通の仕掛け(順番に意味がある) ---
	// ★2つ目の引数が「HTTPSのときだけメモ(Cookie)を送る」の指定。
	//   本番でも練習中はHTTPなので、.env の SECURE_COOKIES で切り替えられる。
	r.Use(middleware.Session(cfg.SecretKey, cfg.IsProduction() && cfg.SecureCookies))
	r.Use(middleware.CSRF())

	// --- 担当ごとの受付を取り付ける ---
	// ★新しい担当ファイル(例:handlers/room.go)を作ったら、ここに1行足す。
	handlers.RegisterPageRoutes(r)
	handlers.RegisterAPIRoutes(r)

	return r, nil
}
