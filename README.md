# Myportfolio

個人のポートフォリオサイトのソースコードです。

**公開中のサイト:** <https://mrrn.jp/>

Go(Gin)で書いたサーバーがHTMLを組み立てて返す、シンプルな構成にしています。
フロントエンドのフレームワークは使わず、素のHTML / CSS / JavaScriptで作っています。

---

## ページ構成

| URL | 内容 |
| --- | --- |
| `/` | トップページ(Welcome / About / Skills / Products / Experience / Contact) |
| `/skills` | 使える技術の一覧(分野ごとに一言コメント付き) |
| `/products` | 制作した作品の一覧 |
| `/privacy` `/terms` | プライバシーポリシー・利用規約 |
| `/admin` | 問い合わせの管理ページ(ログインが必要) |
| `/health` | 動作確認用(`{"status":"ok"}` を返す) |

## 主な機能

- **問い合わせフォーム** — 内容をDBに保存し、管理ページから確認・返信できます。
  SMTPを設定すると、自分宛の通知メールと返信メールを送ります。
- **スパム対策** — reCAPTCHA v3 でフォーム送信を判定します。
- **セキュリティ** — CSRFトークン、セッション管理、管理ページのログイン試行回数の制限。
- **アクセス解析** — Google Analytics 4(本番の `.env` に測定IDを書いたときだけ有効)。
- **演出** — スクロールに合わせて要素を表示するアニメーション。

## 使っている技術

| 分野 | 技術 |
| --- | --- |
| バックエンド | Go 1.26 / Gin / GORM |
| フロントエンド | HTML(Goテンプレート) / CSS / JavaScript |
| データベース | MySQL 8.4 |
| インフラ | Docker Compose / Nginx / Let's Encrypt / AWS EC2 |
| 開発 | air(保存時の自動ビルド) |

## ディレクトリ構成

```
.
├── cmd/server/          起動の入口(main.go)
├── internal/
│   ├── config/          .env の読み込み
│   ├── database/        DB接続とテーブルの自動作成
│   ├── handlers/        URLごとの処理(ページ・問い合わせ・管理画面)
│   ├── mailer/          メール送信
│   ├── middleware/      セッション・CSRF・ログイン確認
│   ├── models/          作品・スキル・問い合わせのデータ
│   ├── recaptcha/       reCAPTCHA の検証
│   ├── router/          URLと処理の組み立て
│   └── view/            テンプレートの描画
├── web/
│   ├── templates/       HTML(layouts / pages)
│   └── static/          CSS・JS・画像
├── docker/nginx/        本番用の Nginx 設定
├── docs/                環境構築・デプロイの手順書
├── compose.yml          開発用
└── compose.prod.yml     本番用
```

## 手元で動かす

Docker Desktop が入っていれば、Goを入れなくても動きます。

```bash
git clone <リポジトリのURL>
cd Myportfolio
cp .env.example .env        # Windows は Copy-Item .env.example .env
docker compose up --build
```

<http://localhost:8080> を開いてください。
DBのテーブルは起動時に自動で作られます。

詳しい手順と困ったときの対処は [docs/SETUP.md](docs/SETUP.md) にあります。

## 本番環境

AWS EC2(メモリ1GB)の上で、Nginx → Gin → MySQL の3つのコンテナを動かしています。

```bash
docker compose -f compose.prod.yml up -d --build
```

サーバーの用意、HTTPS化、更新の手順は [docs/DEPLOY.md](docs/DEPLOY.md) にまとめています。

## ライセンス

ソースコードは [MIT License](LICENSE) です。

ただし、**画像(写真・技術ロゴ)とサイトに掲載している文章・プロフィール・作品情報は対象外**です。
詳しくは [LICENSE](LICENSE) の「適用範囲」を見てください。
