# Gin Template Demo

**Team AIB** のハッカソン用Gin(Go)開発テンプレート。
**`docker compose up` だけで、アプリとDBが揃った開発環境が立ち上がります。**

> リポジトリ名は `case_gin`、画面上の表示名は `Gin Template Demo` です。
> プロダクト名が決まったら、表示名は `internal/view/page.go` の
> `SiteName` の1か所を直してください(全ページのタイトルとヘッダーに反映されます)。

---

## いきなり動かす

```bash
# 1. 金庫を作る(初回だけ)
cp .env.example .env          # Windows: Copy-Item .env.example .env

# 2. 起動する
docker compose up --build
```

→ <http://localhost:8080> を開く

**DBに表を作るコマンドはありません。** 起動するたびに自動で用意されます。

うまくいかないときは **[docs/SETUP.md](docs/SETUP.md)** を見てください。困ったときの対処が全部書いてあります。

---

## 最初に出るデモページについて

起動して `/` を開くと「セットアップ完了 🎉」というデモページが出ます。
**これは消さなくて大丈夫です。** DjangoやRailsの初期画面と同じ仕組みで、条件を満たすと自動で出なくなります。

| | |
| --- | --- |
| 出る条件 | 開発モード **かつ** `internal/handlers/page.go` にまだ `"/"` が無いとき |
| 消える条件 | `page.go` の `r.GET("/", index)` のコメントを外す(それだけ) |
| 本番 | `APP_ENV=production` では最初から出ない。デモ用の表(`demo_todos`)もDBに作られない |
| あとで見たい | `/__demo` で開ける(開発モードのときだけ) |

```go
// internal/handlers/page.go のこの行のコメントを外した瞬間、デモは出なくなります
func RegisterPageRoutes(r *gin.Engine) {
	r.GET("/", index)
	r.GET("/health", health)
}
```

### ⚠ Gin版だけの注意点:URLの二重登録でアプリが落ちる

**Ginは同じURLを2回登録すると、起動した瞬間にプログラムが止まります。**

```
panic: handlers are already registered for path '/'
```

FlaskやDjangoは「先に登録したほうが勝つ」だけで済みますが、**Ginは落ちます。**
そのため、デモ側は「`/` が空いているかどうか」を必ず先に確認してから登録しています。

```go
// internal/demo/demo.go
if !hasRoute(r, http.MethodGet, "/") {
	r.GET("/", index)   // ← 空いているときだけ引き受ける
}
```

この `hasRoute()` の確認を消すと、**自分たちのトップページを作った日にアプリが起動しなくなります。**
`internal/demo/` を触るときは、ここだけ残してください。

> 逆に言うと、`page.go` に `r.GET("/", index)` を足すのは安全です。
> デモ側が自動で譲るので、二重登録にはなりません。

### デモの画面は1枚で完結しています

デモの画面(`internal/demo/page.html`)は、`base.html` も `style.css` も `main.js` も使いません。見た目も動きも全部そのファイルの中に入っています。

デモの役目は「セットアップが動いているか」を見せる**計器**です。計器が `base.html` に頼っていると、`base.html` を自分たちの見た目に作り替えた日に、この画面まで一緒に壊れます。しかも穴(block)の名前を変えただけだとエラーも出ず、真っ白になるだけなので「アプリが壊れた?」と勘違いします。1枚完結にしておけば、`base.html` を好きなだけ作り替えても、この計器だけは最後まで正しく動きます。

そのため `page.html` は `web/templates/pages/` ではなく `internal/demo/` に置いてあります(あそこに置くと共通の仕組みが自動で拾って `base.html` と合体させてしまうため)。読み込みは `go:embed` でアプリ本体に取り込んでいます。

ページの書き方の見本は `web/templates/pages/login.html` を見てください(`base.html` を継いだ本物のページです)。

### デモが本当に不要になったら(2か所)

1. `internal/demo/` フォルダを削除
2. `internal/router/router.go` の「デモ」の数行を削除

---

## 使っている技術

| 分類 | 技術 |
| --- | --- |
| フロント | 素のHTML / CSS / JavaScript(Goの標準テンプレート) |
| バック | Go 1.26 / Gin / GORM |
| DB | MySQL 8.4(確認は DBeaver、ポートは **3308**) |
| 環境 | Docker Compose / air(保存したら自動で作り直す道具) |
| 本番 | Nginx / AWS(**`compose.prod.yml` に構成済み。本番イメージは約48MB**) |

---

## フォルダの地図

「アプリ = 1軒のお店」だと思って読んでください。

```
case_gin/
├── cmd/server/main.go      お店の開店作業(起動の入口)
│
├── internal/               お店の裏側(★他のプロジェクトからは入れない場所)
│   ├── config/             設定を1か所に集める
│   ├── database/           DBに繋ぐ共用の道具
│   ├── models/             データの形(DBの表)を決める
│   ├── middleware/         入口に置く仕掛け(セッション・ログイン確認・成りすまし対策)
│   ├── handlers/           受付。URL → 処理
│   │   ├── page.go           画面(HTML)を返す
│   │   ├── auth.go           ログイン・新規登録
│   │   └── api.go            JavaScript向けにデータだけ返す
│   ├── demo/               動作確認用のデモ(開発モード限定・触らない)
│   │                       画面(page.html)も1枚完結でここに入っている
│   ├── router/             受付をまとめて組み立てる(土台)
│   └── view/               型紙(base.html)を使う仕組み(土台)
│
├── web/                    お客さんの目に入るもの
│   ├── templates/
│   │   ├── layouts/base.html   全ページ共通の型紙
│   │   ├── pages/              各ページの中身(ここに置くだけで使える)
│   │   └── partials/           型紙が長くなったら部品を切り出す置き場
│   └── static/                 CSS / JS / 画像
│
├── media/                  利用者が上げたファイル(★中身はGitHubに上げない)
│
├── docs/                   チームで見る手順書(SETUP / Gohelp / DEPLOY)
├── tools/                  開発中だけ使う小道具スクリプト
│
├── compose.yml             アプリとDBをまとめて動かす段取り表(開発用)
├── compose.prod.yml        本番用の段取り表(★開発中は使わない)
├── docker/nginx/           本番でCSSと画像を配るNginxの設定
├── Dockerfile              箱を組み立てるレシピ
├── .air.toml               保存したら自動で作り直す設定
├── go.mod / go.sum         買い物リストとレシート(全員が同じ部品を使うための記録)
│
├── .vscode/                チーム共通のエディタ設定
│
├── LICENSE                 使ってよい条件(MIT)
├── .env                    金庫(★GitHubに上げない)
└── .env.example            金庫の中身の見本(こちらは上げる)
```

> **`internal` という名前について**
> Goでは `internal` という名前のフォルダは「このプロジェクトの中からしか使えない」という
> 決まりになっています。お店のバックヤードのようなもので、外から勝手に触られません。
> 特別な設定は不要で、名前を付けるだけでそうなります。

---

## 担当ごとに触る場所

**基本的に他の人と同じファイルを触らないように分けてあります。** これでコンフリクト(変更の取り合い)がほぼ起きません。

| 担当 | 触る場所 |
| --- | --- |
| 見た目 | `web/templates/` `web/static/css/` |
| 画面の動き | `web/static/js/` |
| データの形 | `internal/models/` |
| URLと処理(画面) | `internal/handlers/page.go` |
| URLと処理(データ) | `internal/handlers/api.go` |
| ログイン | `internal/handlers/auth.go` |

`main.go` `router/` `view/` `config/` `database/` `middleware/` `compose.yml` `Dockerfile` は**土台**です。
触る必要が出たら、**先にチームに共有してから**変更してください(全員に影響します)。

---

## ページを1枚増やす手順

1. **`web/templates/pages/` にHTMLを1枚置く**(`login.html` をコピーするのが早い)

   ```html
   {{ define "content" }}
   <h1>マイページ</h1>
   {{ end }}
   ```

   > ヘッダーとフッターは書きません。`base.html` から自動的に付きます。

2. **`internal/handlers/page.go` に2つ足す**

   ```go
   func RegisterPageRoutes(r *gin.Engine) {
       r.GET("/health", health)
       r.GET("/mypage", myPage)   // ← 1行足す
   }

   func myPage(c *gin.Context) {  // ← 関数を書く
       c.HTML(http.StatusOK, "mypage.html", view.Page(c, gin.H{
           "Title": "マイページ",
       }))
   }
   ```

3. 保存して数秒待つ → <http://localhost:8080/mypage>

**登録作業はこれだけです。** テンプレートの一覧に書き足す必要はありません(自動で見つけます)。

ログインしている人だけに見せたいときは、1行挟むだけです。

```go
r.GET("/mypage", middleware.RequireLogin(), myPage)
```

---

## base.html(型紙)の仕組み

**今回いちばん新しい考え方なので、ここだけ先に押さえてください。**

ヘッダーとフッターを全ページにコピペすると、直すときに全ファイルを回ることになります。
そこで**共通部分を1枚の「型紙」にまとめ、各ページは真ん中の中身だけを書く**形にしています。

```
base.html(型紙)                    pages/index.html(中身)
┌──────────────────┐
│ ヘッダー          │
├──────────────────┤
│                  │  ←──  {{ define "content" }}
│  ここが穴         │            <h1>ホーム</h1>
│                  │        {{ end }}
├──────────────────┤
│ フッター          │
└──────────────────┘
```

- 型紙側:`{{ block "content" . }}{{ end }}` と書いた場所が**穴**になる
- ページ側:`{{ define "content" }} ～ {{ end }}` に書いた中身が**その穴に入る**
- 名前(`content`)が両者で一致していることが条件

穴は3つ用意してあります。

| 穴の名前 | 用途 |
| --- | --- |
| `content` | ページの本体(必須) |
| `head_extra` | そのページだけで使うCSSを足したいとき |
| `scripts` | そのページだけで使うJSを足したいとき |

**ヘッダーを直したいときは `base.html` を1枚直すだけ**で、全ページに反映されます。

> ★つまずきやすい点が2つあります。詳しくは `web/templates/layouts/base.html` の
> 冒頭のコメントに書いてあるので、最初に一度読んでください。
>
> 1. `{{ .Title }}` のように書いても、**渡していない情報は空になる**(エラーも出ない)
> 2. `{{ range }}` の中では**ドットの意味が変わる**(全体の情報は `{{ $.SiteName }}`)

---

## よく使うコマンド

| やりたいこと | コマンド |
| --- | --- |
| 起動する | `docker compose up` |
| 裏で起動する | `docker compose up -d` |
| 止める | `docker compose down` |
| エラーを見る | `docker compose logs -f web` |
| DBを作り直す(データは消える) | `docker compose exec web go run ./cmd/server -reset-db` |
| 箱の中に入る | `docker compose exec web bash` |
| ライブラリを追加する | `docker compose exec web go get <ライブラリ>` |
| 書き方を自動で整える | `docker compose exec web go fmt ./...` |
| 怪しい書き方を調べる | `docker compose exec web go vet ./...` |

---

## 3つのテンプレートの対応表

同じ構成・同じ画面で作ってあるので、1つ分かれば他も読めます。

| やること | Flask版 | Gin版 | Django版 |
| --- | --- | --- | --- |
| 起動の入口 | `src/web/app.py` | `cmd/server/main.go` | `manage.py` + `config/` |
| 設定 | `config.py` | `internal/config/` | `config/settings.py` |
| データの形 | `models.py` | `internal/models/` | `main/models.py` |
| 画面を返す | `routes.py` | `handlers/page.go` | `main/views.py` |
| ログイン | `auth/routes.py` | `handlers/auth.go` | `accounts/views.py` |
| 型紙 | `base.html`(Jinja) | `base.html`(Go) | `base.html`(Django) |
| 表を作る | `flask init-db` | 起動時に自動 | 起動時に自動 |
| 表の形を変える | 作り直し(データ消滅) | 列の追加のみ可 | **データを保ったまま変更可** |
| 管理画面 | 無い | 無い | **`/admin/`** |
| アプリのポート | 5000 | 8080 | 8000 |
| DBのポート | 3307 | 3308 | 3309 |
| 本番イメージ | 438MB | **48MB** | 913MB |

> ポートをずらしてあるので、**3つ同時に起動しても衝突しません。**

---

## ドキュメント

- **[docs/SETUP.md](docs/SETUP.md)** — 環境構築、日々の操作、DBeaverでの接続、困ったときの対処
- **[docs/Gohelp.md](docs/Gohelp.md)** — Goの書き方(コーディング経験者向けの早わかり)
- **[docs/DEPLOY.md](docs/DEPLOY.md)** — 本番に出す手順

---

## ライセンス

MIT License([LICENSE](LICENSE))

自由に使って、改造して、公開して構いません(商用も可)。
条件は「著作権表示を残すこと」だけです。
