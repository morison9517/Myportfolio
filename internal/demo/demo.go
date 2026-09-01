// =============================================================================
// internal/demo/ = 動作確認用のデモページ一式(自分たちのプロダクトには含まれません)
//
// ▼ ★このフォルダの決まり
//
//  1. 開発モード(APP_ENV=development)のときだけ取り付けられる。
//     本番では取り付けないので、消し忘れても本番には出ない。
//     デモ用の表(todos)も、本番のDBには作られない。
//
//  2. トップページ("/")は、handlers/page.go にまだ "/" が無いときだけ引き受ける。
//     自分たちのトップページを書いた瞬間に、デモは自動で出なくなる。
//     → 開発を始めるときに「デモを消す作業」は不要。
//
//  3. いつでも見たいときは /__demo で開ける(開発モードのときだけ)。
//
//  4. ★デモの画面(page.html)は1枚で完結している。
//     base.html も style.css も main.js も使わない。
//
//     デモの役目は「セットアップが動いているか」を見せる計器なので、
//     共通のファイルに頼らせていない。
//     base.html を自分たちの見た目に作り替えても、この画面は壊れない。
//     逆に base.html を継がせていると、穴(block)の名前を変えた瞬間に
//     エラーも出ずに真っ白になり、「アプリが壊れた?」と勘違いする。
//
//     ★DjangoやRailsの「セットアップ完了ページ」と同じ仕組みです。
//     あちらも、自分でトップページを作ると自動で出なくなります。
//     あちらも1枚完結で、他のファイルに一切頼っていません。
//
// ▼ ★ページの書き方の見本にはしないこと
//
//	デモは1枚完結なので、普通のページとは書き方が違います。
//	見本にするなら web/templates/pages/login.html です
//	(base.html を継いだ本物のページ)。
//
// ▼ ★Ginだけの注意点(ここが他の2つと違う)
//
//	Ginは同じURLを2回登録すると、起動した瞬間にプログラムが止まる。
//
//	    panic: handlers are already registered for path '/'
//
//	FlaskやDjangoは「先に登録したほうが勝つ」だけで済むが、Ginは落ちる。
//	そこで下の hasRoute() で「"/" が空いているか」を必ず先に確認している。
//	この確認を消すと、自分たちのトップページを作った日にアプリが起動しなくなる。
//
// ▼ 本当に不要になったら(2か所)
//
//	・このフォルダ internal/demo/ を削除
//	・internal/router/router.go の「デモ」の数行を削除
//
// =============================================================================
package demo

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/database"
	"case_gin/internal/middleware"
)

// ★//go:embed = このHTMLをアプリ本体の中に取り込む指示。
//
//	ふつうテンプレートは web/templates/ から読むが、そこに置くと
//	共通の仕組み(internal/view/render.go)が自動で拾って
//	base.html と合体させてしまう。1枚完結にしたいので、
//	デモのHTMLはこのフォルダに置いて、自分で読み込んでいる。
//
//	おかげでデモの削除は「フォルダを消す + router.go の数行を消す」だけで済む。
//
//go:embed page.html
var pageHTML string

// デモ画面のテンプレート。起動時に1回だけ組み立てる。
var pageTmpl *template.Template

// 設定は Register() で受け取って覚えておく(画面に「ログイン機能:有効」を出すため)。
var appConfig *config.Config

// Register = デモを取り付ける。router.go から、開発モードのときだけ呼ばれる。
func Register(r *gin.Engine, cfg *config.Config) error {
	appConfig = cfg

	// デモ用の表を用意する。
	// ★ここでやっているので、本番のDBには todos 表が作られない。
	if err := database.DB.AutoMigrate(&Todo{}); err != nil {
		return err
	}

	// page.html を読み込む。書き間違いがあればここで気づける。
	tmpl, err := template.New("demo").Parse(pageHTML)
	if err != nil {
		return err
	}
	pageTmpl = tmpl

	// /__demo は常に開けるようにしておく(自分たちのトップページを作った後も
	// 「DBに繋がっているか」をここで確認できる)。
	r.GET("/__demo", index)
	registerAPIRoutes(r)

	// ★トップページが空いているときだけ、デモが "/" を引き受ける。
	//   handlers/page.go に "/" を書いた後は、この条件が外れてデモは出なくなる。
	//   (確認せずに登録すると、URLの二重登録でGinが起動時に落ちる)
	if !hasRoute(r, http.MethodGet, "/") {
		r.GET("/", index)
	}

	return nil
}

// hasRoute = そのURLが既に登録されているか調べる。
//
// r.Routes() は「今この受付に登録されている全URLの一覧」。
// ★Ginは同じURLを2回登録すると落ちるので、登録の前に必ずここを通す。
func hasRoute(r *gin.Engine, method, path string) bool {
	for _, info := range r.Routes() {
		if info.Method == method && info.Path == path {
			return true
		}
	}
	return false
}

// index = デモページ。"/" と "/__demo" の両方から使われる。
//
// ★普通のページと違い、c.HTML(...) を使っていない。
//
//	c.HTML は base.html と合体させる共通の仕組みを通るので、
//	1枚完結のこのページには使えない。ここでは自分で組み立てて返している。
//	自分たちのページを作るときは c.HTML を使うこと(見本は handlers/page.go)。
func index(c *gin.Context) {
	dbStatus, dbMessage := checkDB()

	data := map[string]any{
		"Title":       "セットアップ確認",
		"DBStatus":    dbStatus,
		"DBMessage":   dbMessage,
		"AuthEnabled": appConfig.AuthEnabled,
		"CSRFToken":   middleware.CSRFToken(c),

		// ★真偽値だけを渡している(利用者オブジェクトそのものは渡さない)。
		//   画面に名前を出すと .Username を読むことになり、チームが
		//   Userの項目名を変えた日にデモが壊れる。
		//   「ログイン中か」だけならフレームワーク側の概念なので絶対に壊れない。
		"LoggedIn": middleware.CurrentUser(c) != nil,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)

	if err := pageTmpl.Execute(c.Writer, data); err != nil {
		// ここまで来たらHTMLを書き始めた後なので、画面には途中まで出ている。
		// 原因はターミナルのログで確認する。
		_ = c.Error(err)
	}
}

// checkDB = DBに繋がるか実際に試す。
//
// 「返事してください」という最小の確認を送り、返事が来るかで判定している。
//
// エラーで落とさず画面は出す理由:
// DBが起動しきっていないだけでこの画面が真っ白になると原因が分かりにくい。
// 画面は出しつつ「DBだけ未接続」と伝えたほうが切り分けが速い。
func checkDB() (string, string) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return "ng", truncate(err.Error())
	}

	if err := sqlDB.Ping(); err != nil {
		return "ng", truncate(err.Error())
	}

	return "ok", ""
}

// truncate = エラー文が数百文字になることがあるので、先頭だけにする。
func truncate(s string) string {
	const limit = 120
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
