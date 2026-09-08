# 本番に出す手順

このファイルは**サーバーにアプリを載せるとき**に見る紙です。
日々の開発は [SETUP.md](SETUP.md) を見てください。

> ★**本番前に一度、必ず練習してください。**
> ここに書いてある手順を、本番当日に初めて打つのは危険です。
> 一度通しておけば、当日は同じことを繰り返すだけになります。

---

## 開発と本番の違い(先に頭に入れておくこと)

| | 開発 | 本番 |
| --- | --- | --- |
| 使うファイル | `compose.yml` | **`compose.prod.yml`** |
| 箱の中身 | Go本体 + air(1GB近い) | **実行ファイル1個(20MB程度)** |
| コードの反映 | 保存したら即 | **build し直したとき** |
| CSS・画像を配る人 | Gin | **Nginx** |
| DBのポート | PCから見える(3308) | **開けない** |
| エラー画面 | 詳しく出る | **出ない(内部情報が漏れるため)** |

**「保存しても本番に反映されない」のは正しい動きです。** 本番は実行ファイルに固めたもので動きます。

---

## 0. AWS EC2 にサーバーを立てる(初回だけ)

> このプロジェクトは **AWS EC2 の無料枠(1GB)** で動かす前提で書いています。
> 別のサーバーを使う場合、この章は飛ばして1章から読んでください。

### ★先に知っておく費用のこと

| | 無料か |
| --- | --- |
| インスタンスの稼働時間(t2.micro / t3.micro 750時間/月) | **無料枠あり** |
| EBS(ディスク)30GBまで | **無料枠あり** |
| **パブリック IPv4 アドレス** | **有料(1つ 月$3.6前後)** |
| 通信量(外向き) | 一定量まで無料、超えると従量 |

**「完全に0円」にはなりません。** 2024年2月から、IPv4アドレスは
インスタンスに付けていても課金対象です。

> ★**請求アラートを必ず設定してください。**
> AWSコンソールの「Billing」→「Budgets」で、月$5などの上限を決めて
> メール通知を設定します。無料枠切れに気づかず数万円、という事故が最も多い。
>
> ★2025年以降に作ったアカウントは、無料枠の内容が以前と変わっている
> 場合があります。自分のアカウントがどの条件かは請求ダッシュボードで確認を。

### ① インスタンスを作る

EC2 →「インスタンスを起動」で次のように選びます。

| 項目 | 選ぶもの | 理由 |
| --- | --- | --- |
| AMI | **Ubuntu Server 24.04 LTS** | `docker compose` が公式手順でそのまま入る |
| インスタンスタイプ | **t2.micro** または **t3.micro** | 無料枠の対象(どちらが対象かはリージョンによる) |
| キーペア | 新規作成して**必ずダウンロード** | 無くすとログインできなくなる |
| ストレージ | **20GB** gp3 | 既定の8GBだとDockerのイメージで埋まる |

### ② セキュリティグループ(通信の許可)

| ポート | 許可する相手 | 用途 |
| --- | --- | --- |
| 22 | **自分のIPのみ** | SSHログイン |
| 80 | すべて(0.0.0.0/0) | HTTP |
| 443 | すべて(0.0.0.0/0) | HTTPS |

> ★**3306(MySQL)は絶対に開けないこと。**
> 開けた瞬間から、世界中から総当たりでログインを試されます。

### ③ Elastic IP を割り当てる

そのままだと、インスタンスを停止して起動するたびにIPが変わり、
ドメインの設定が毎回ずれます。Elastic IP を取って固定します。

> ★使っていない Elastic IP は割高に課金されます。
> **インスタンスを消すときは Elastic IP も解放してください。**

### ④ ドメインを向ける

ドメインの管理画面(お名前.comなど)でDNSレコードを設定します。

| 種類 | ホスト名 | 値 |
| --- | --- | --- |
| A | `@`(またはmrrn.jp) | Elastic IP のアドレス |
| A | `www` | Elastic IP のアドレス |

反映には数分〜数時間かかります。確認は次のコマンドで。

```bash
nslookup mrrn.jp
```

### ⑤ ★スワップを作る(1GBのサーバーでは必須)

メモリが1GBしかないので、これが無いとビルド中やMySQLの起動時に
プロセスが強制終了されます。**症状は「原因不明で落ちる」なので、先に作ります。**

```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

確認します。`Swap:` が 2.0Gi になっていれば成功です。

```bash
free -h
```

> ★MySQL側の節約設定は `compose.prod.yml` に入れてあります
> (`performance-schema=OFF` など)。2GB以上のサーバーに移したら消して構いません。

---

## 1. 準備(初回だけ)

### ① サーバーに Docker を入れる

Ubuntu 24.04 の場合、まだ入っていないので入れます。

```bash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

`sudo` 無しで使えるようにします。**一度ログアウトして入り直すまで反映されません。**

```bash
sudo usermod -aG docker $USER
```

入り直したら確認します。

```bash
docker version
docker compose version
```

### ② コードを置く

```bash
git clone <リポジトリのURL>
cd case_gin
```

### ③ 金庫(`.env`)を作る

```bash
cp .env.example .env
```

**そして必ず中身を書き換えます。** 見本のままだと危険です。

| 項目 | 本番で入れる値 |
| --- | --- |
| `SECRET_KEY` | **長いランダムな文字列**(下のコマンドで作る) |
| `APP_ENV` | `production` |
| `DB_PASSWORD` / `DB_ROOT_PASSWORD` | **開発用と違うパスワード** |

割り印(SECRET_KEY)はこれで作れます。

```bash
python3 -c "import secrets; print(secrets.token_urlsafe(50))"
```

> ★`APP_ENV=development` のまま本番に出すと、起動ログが冗長になり、
> 内部の情報が出やすくなります。必ず `production` にしてください。
>
> ★`SECRET_KEY` を見本のままにすると、Ginは**起動を拒否します。**
> 割り印が初期値のままだと、ブラウザに預けたメモの中身を偽造されるためです。

---

## 2. 起動する

```bash
docker compose -f compose.prod.yml up -d --build
```

**`-f compose.prod.yml` を毎回付けます。** 忘れると開発用が起動します。

初回は数分かかります。終わったら状態を見ます。

```bash
docker compose -f compose.prod.yml ps
```

`db` `web` `nginx` の3つが `running` なら成功です。

**DBの表を作るコマンドは要りません。** アプリが起動時に自分で用意します。

### 開いて確認

ブラウザで `http://<サーバーのアドレス>/` を開きます。

> ★開いたURLが **404 になるときは、そのURLをまだ登録していない**だけです。
> 登録済みのURLは `internal/handlers/page.go` で確認できます。
> `/health` は必ず登録されているので、アプリ自体が動いているかの切り分けに使えます。

> DBの中身を直接見たいときは、下の「本番のDBをDBeaverで見たいとき」を参照。

---

## 3. 公開後にサイトを直したとき

### 流れ

1. **手元(自分のPC)で直す**
2. **コミットして push する**
3. **サーバーに入って、下の2行を打つ**

```bash
git pull
docker compose -f compose.prod.yml up -d --build
```

**この2行だけです。** DBの表づくりは起動時に自動で走ります。

> ★手元で直しただけでは反映されません。**push を忘れると `git pull` で何も落ちてきません。**

### ★メモリ1GBのサーバーでは、先に止めてから build する

build のとき、サーバーの上でGoのコンパイルが走ります。
動いているMySQLと同時だとメモリが足りず、スワップに逃げて非常に遅くなります。

**Go やHTMLを直したときは、先に止めてください。**

```bash
git pull
docker compose -f compose.prod.yml down
docker compose -f compose.prod.yml up -d --build
```

止まっている数分だけサイトが見られなくなりますが、
1GBの機械ではこちらのほうが速く終わります。

> ★CSS・画像だけの変更なら build 自体が不要なので、止める必要もありません。

古いイメージがディスクを埋めるので、たまに掃除します。

```bash
docker image prune -f
docker builder prune -f
```

---

### 直した場所によって、必要な操作が違う

| 直した場所 | 中身 | 必要な操作 |
| --- | --- | --- |
| `web/static/` | CSS・JS・画像 | **`git pull` だけ**(build 不要) |
| `web/templates/` | HTML | `git pull` + **build し直し** |
| `internal/` `cmd/` | Goのプログラム | `git pull` + **build し直し** |
| `.env` | 設定・パスワード | サーバー上で直接編集 + `up -d`(build 不要) |

**CSSと画像だけが特別扱いです。**
`compose.prod.yml` で Nginx がサーバー上のフォルダを直接見ているため
(`./web/static:/var/www/static:ro`)、`git pull` した瞬間に新しいものが配られます。

一方 HTML は Dockerfile の `COPY web ./web` で箱の中に焼き込まれるので、
build し直さないと古いままです。

> ★迷ったら `up -d --build` を打って構いません。
> 必要のないときでも、無駄になるだけで害はありません。

---

### よく直す場所の早見表

| やりたいこと | 直すファイル | build |
| --- | --- | --- |
| 作品(Products)を追加・修正 | `internal/models/product.go` | 要 |
| 経歴・自己紹介・スキル | `web/templates/pages/index.html` | 要 |
| プライバシーポリシー・利用規約 | `web/templates/pages/privacy.html` `terms.html` | 要 |
| タブに出るサイト名 | `internal/view/page.go` の `SiteName` | 要 |
| 色・余白・アニメーションの速さ | `web/static/css/` の各ファイル | 不要 |
| 作品やスキルの画像を差し替え | `web/static/images/` | 不要 |

---

### ★CSSを直したときの落とし穴(キャッシュ)

`docker/nginx/prod.conf` に `expires 30d` があるため、
**一度サイトを見た人のブラウザは、最大30日間 古いCSSを使い続けます。**

自分で確認するときは `Ctrl + F5`(スーパーリロード)で最新が見られますが、
**すでに見た人には届きません。**

確実に配りたいときは、`web/templates/layouts/base.html` の読み込みに
バージョンを付けて、直すたびに数字を上げます。

```html
<link rel="stylesheet" href="/static/css/base.css?v=2">
```

別のファイル扱いになるので、全員に新しいものが届きます。
(この場合はHTMLの変更なので build が要ります)

---

### 反映されたか確かめる

```bash
docker compose -f compose.prod.yml ps            # db / web / nginx が running か
docker compose -f compose.prod.yml logs -f web   # 起動時のエラーが出ていないか
```

ブラウザで開くときは、**必ず `Ctrl + F5`** で開いてください。
普通の再読み込みだと、手元のブラウザが古いCSSを使い、
「直したのに変わらない」と勘違いします。

---

### DBの列を増やしたとき

> ★GORMの `AutoMigrate` は、列を**増やす**ことはできますが、
> 型を変えたり列を消したりはしません。
>
> 本番でどうしても変える必要が出たら、DBeaverで直接ALTERするか、
> 一度データを捨てて作り直すことになります。
> **本番に出す前にモデルを固めておくのが安全です。**

---

## 4. HTTPSにする

練習の段階では `http://` のままで構いません。**本番では必ずHTTPSにします。**

> ★HTTPSにしたら、あわせて **HTTP/2 を有効にしてください。**
> `docker/nginx/prod.conf` の 443 のブロックに `http2 on;` を1行足すだけです。
> CSSを複数ファイルに分けているため、まとめて取得できるHTTP/2の効果が大きく出ます。
> ブラウザはHTTPSのときしかHTTP/2を使わないので、それまでは書いても効きません。

やり方は2つあります。

このプロジェクトは **Let's Encrypt(無料)** を使う前提で、
`docker/nginx/prod.conf` と `compose.prod.yml` の設定を用意してあります。

### ★順番が重要

**証明書が無い状態で Nginx を起動すると、設定を読めずに落ちます。**
必ず「証明書を取る → Nginx を起動」の順で進めてください。

### ① 最新のコードを取る(まだ Nginx は再起動しない)

```bash
cd ~/Myportfolio
git pull
```

> `git pull` しただけでは Nginx は古い設定のまま動き続けます。
> 再起動するまで反映されないので、この時点では問題ありません。

### ② Nginx を止めて、証明書を取る

80番を certbot に明け渡す必要があるので、一時的に止めます。

```bash
docker compose -f compose.prod.yml stop nginx

docker run --rm -p 80:80   -v case_gin_prod_certbot_conf:/etc/letsencrypt   certbot/certbot certonly --standalone   -d mrrn.jp -d www.mrrn.jp   --email 自分のメールアドレス --agree-tos --no-eff-email
```

`Successfully received certificate` と出れば成功です。

> ★`--email` は証明書の期限切れを知らせるために使われます。実際に受け取れる
> アドレスを入れてください。
>
> ★**失敗を繰り返さないこと。** Let's Encrypt には
> 「同じドメインで1週間に5回まで」といった回数制限があります。
> DNSが未反映のまま試すと、その回数を無駄に消費します。
> 先に `nslookup mrrn.jp` でIPが返ることを確認してください。

### ③ Nginx を起動する

```bash
docker compose -f compose.prod.yml up -d nginx
docker compose -f compose.prod.yml logs nginx | tail -20
```

エラーが出ていなければ、`https://mrrn.jp/` が開きます。
`http://` で開いても、自動でHTTPSに切り替わります。

### ④ ★`.env` を直して web を作り直す

```
SECURE_COOKIES=true
```

HTTPで動かすために `false` にしていたはずなので、`true` に戻します。

```bash
nano .env
docker compose -f compose.prod.yml up -d web
```

> ★`restart` ではなく `up -d` を使うこと。
> `restart` では `.env` を読み直しません。

### ⑤ 自動更新を仕掛ける

証明書は90日で切れます。放置するとサイトが開けなくなるので、
月に一度、更新を試すようにしておきます。

```bash
crontab -e
```

開いたファイルの末尾に、次の1行を足します。

```
0 4 1 * * cd ~/Myportfolio && docker run --rm -v case_gin_prod_certbot_conf:/etc/letsencrypt -v case_gin_prod_certbot_www:/var/www/certbot certbot/certbot renew --webroot -w /var/www/certbot --quiet && docker compose -f compose.prod.yml exec nginx nginx -s reload
```

毎月1日の午前4時に更新を試し、成功したら Nginx に設定を読み直させます。
期限が近くないときは certbot が何もせずに終わるので、毎月動かして構いません。

> ★更新は `--standalone` ではなく `--webroot` を使っています。
> Nginx を止めずに済むため、更新のたびにサイトが落ちません。
> `prod.conf` の 80番に置いた `/.well-known/acme-challenge/` がこのためのものです。

### ⑥ 動作を確認したら HSTS を有効にする

`docker/nginx/prod.conf` の443ブロックにある次の行のコメントを外します。

```
add_header Strict-Transport-Security "max-age=31536000" always;
```

「今後このサイトはHTTPSでしか開くな」とブラウザに覚えさせる指定です。

> ★**一度覚えさせると、期間中(1年)は取り消せません。**
> 証明書が切れるとサイトが完全に開けなくなります。
> HTTPSが数日安定して動くことを確認してから有効にしてください。

### 参考: AWSのロードバランサーに任せる方法

証明書の更新が自動になりますが、**動かしているだけで月20ドル前後かかります。**
VPC・サブネット2つ・ターゲットグループの設定も必要です。
サーバー1台の構成では、上のLet's Encryptのほうが安く簡単です。

### ★HTTPSにしたら必ず戻すこと

`.env` の1行です。

```
SECURE_COOKIES=true
```

HTTPで動かすために `false` にしていた場合、**戻し忘れるとメモ(Cookie)が
HTTPSでない経路でも送られてしまいます。**

---

## 5. よく使うコマンド

| やりたいこと | コマンド |
| --- | --- |
| 起動 | `docker compose -f compose.prod.yml up -d --build` |
| 直したものを反映(→ 3章) | `git pull` してから同じコマンド |
| 停止 | `docker compose -f compose.prod.yml down` |
| 状態を見る | `docker compose -f compose.prod.yml ps` |
| ログを見る | `docker compose -f compose.prod.yml logs -f web` |
| Nginxのログ | `docker compose -f compose.prod.yml logs -f nginx` |
| 箱の中に入る | `docker compose -f compose.prod.yml exec web bash` |

> ★`down -v` は**絶対に打たないでください。**
> `-v` はDBと画像の保管庫ごと消す指定です。利用者のデータが全部消えます。

---

## 6. 本番のDBをDBeaverで見たいとき

本番では**DBのポートを開けていません。** 開けると世界中からログインを試されます。

代わりに、SSHのトンネルを通して見ます。DBeaverの接続設定で
「SSH」タブを開き、サーバーへのSSH情報を入れてください。
そのうえで、ホストは `127.0.0.1`、ポートは `3306` にします。

> トンネル = 自分のPCとサーバーの間に専用の通路を1本引くイメージです。
> 通路の中を通るので、外からは見えません。

---

## 7. 困ったとき

### 画面が真っ白 / デザインが崩れている

Nginxが `web/static/` を見つけられていない可能性があります。
`compose.prod.yml` の nginx に、この行があるか確認してください。

```yaml
- ./web/static:/var/www/static:ro
```

> ★Django版と違い、Ginには「CSSを1か所に集めるコマンド」がありません。
> ソースのフォルダをそのままNginxに見せる作りになっています。
> **サーバー上に `git clone` したソースが必要**なのはこのためです。

### メッセージが表示されない / セッションが保持されない

**`SECURE_COOKIES` が原因です。**

`http://` でアクセスしているのに `true` になっていると、
ブラウザにメモ(Cookie)が保存されません。エラーも出ないので、まずここを疑ってください。

`https://` にするまでは `.env` に `SECURE_COOKIES=false` を入れてください。

### ボタンを押すと 400(整理券が正しくありません)

フォームに整理券が入っていません。HTMLにこの1行があるか確認してください。

```html
<input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />
```

JavaScriptから送っている場合は、`main.js` の `api` を使えば自動で付きます。

### アクセス元のIPが全部同じに見える

`.env` の `TRUSTED_PROXIES` が空です。Nginxを通すと、Ginから見た相手は
Nginxになります。本当のアクセス元はヘッダーに書いてありますが、
**無条件に信じると詐称できてしまう**ので、既定では無視する作りです。

`compose.prod.yml` で `TRUSTED_PROXIES: 172.16.0.0/12` を渡しています。
Dockerの内部ネットワークの範囲です。

### 502 Bad Gateway と出る

Nginxは動いているが、Gin が返事をしていません。
web のログを見てください。だいたい起動時のエラーです。

```bash
docker compose -f compose.prod.yml logs web
```

### プロフィールアイコンが表示されない

`media` の保管庫がNginxから見えていない可能性があります。
`compose.prod.yml` の nginx に `media_files:/var/www/media:ro` があるか確認してください。

**なお、開発用と本番用で画像の保管庫は別です。** 開発中に入れた画像は本番にはありません。

### 起動しない / 起動してすぐ落ちる

`volumes:` に `- .:/app` を書いていないか確認してください。
**本番でこれを書くと、固めた実行ファイルが隠れて起動しません。**
本番でいちばん多い事故です。

`SECRET_KEY` を見本のままにしている場合も、Ginは起動を拒否します。
ログに理由が出ているので確認してください。

---

## 8. 本番に出す前のチェックリスト

デプロイの練習のときに、この順で確認してください。

- [ ] `.env` の `APP_ENV` が `production`
- [ ] `.env` の `SECRET_KEY` を見本から変えた(★変えないと起動しません)
- [ ] `.env` の `DB_PASSWORD` を開発用から変えた
- [ ] `/` が開ける
- [ ] `/health` が `{"status":"ok"}` を返す
- [ ] **CSSが当たっている**(見た目が崩れていない)
- [ ] `docker compose -f compose.prod.yml restart` して、データが残っている

**最後の1つが特に大事です。** 再起動でデータが消えるなら、保管庫の設定が間違っています。
