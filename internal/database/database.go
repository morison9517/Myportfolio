// =============================================================================
// database.go = 共用の道具置き場(DBに繋ぐ係)
//
// ▼ なぜ1か所にまとめるのか
//
//	DBへの回線は「アプリ全体でたった1本」を全員で使い回すのが正しい。
//	ファイルごとに勝手に繋ぐと、回線が増えすぎてDBが悲鳴を上げるうえ、
//	「保存したのに他から見えない」という不具合の原因にもなる。
//
// ▼ 使い方
//
//	起動時に1回 database.Connect(cfg) を呼べば、あとはどのファイルからでも
//	database.DB を使ってDBを操作できる。
//
// =============================================================================
package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"case_gin/internal/config"
	"case_gin/internal/models"
)

// DB = みんなで使い回すDBの窓口。
//
// ★ここでは箱を用意しただけで、中身は空(nil)。
//
//	Connect() を呼んだ瞬間に中身が入る。
var DB *gorm.DB

// Connect = DBに接続して、上の DB に入れる。
func Connect(cfg *config.Config) error {
	// 開発中はGORMが実行したSQLを全部ログに出す。
	// 「思ったデータが取れない」ときに、実際に投げられたSQLが見えると原因が一瞬で分かる。
	// 本番はうるさい&情報が漏れるので、エラーだけにする。
	logLevel := gormlogger.Info
	if cfg.IsProduction() {
		logLevel = gormlogger.Error
	}

	// ★ここで待つ処理を入れている理由
	//   MySQLは起動命令から実際に使えるまで10〜30秒かかる。
	//   compose.yml 側でも待つようにしてあるが、それでも間に合わないことがある。
	//   繋がるまで数回やり直すことで「起動直後だけ落ちる」を防ぐ。
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: gormlogger.Default.LogMode(logLevel),
		})
		if err == nil {
			log.Println("[db] 接続できました")
			return nil
		}

		log.Printf("[db] 接続待ち... (%d回目)", attempt)
		time.Sleep(2 * time.Second)
	}

	return err
}

// Migrate = models/ に書いた設計図どおりにDBへ表を作る。
//
// 起動のたびに自動で実行される。
//   - 表が無ければ作る
//   - 項目(列)が増えていれば足す
//   - 既にあるデータは消えない
//
// ★注意:列の「型」を変えた場合(80文字→200文字など)は自動で追従しないことがある。
//
//	そのときは作り直す(docs/SETUP.md の「DBを作り直す」を参照)。
//
// ★新しいモデルを作ったら、下のリストに1行足すこと。忘れると表が作られない。
//
// ※ デモ用の表(demo_todos)はここには書かない。
//
//	開発モードのときだけ internal/demo/demo.go が自分で用意するので、
//	本番のDBには作られない。
func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
	)
}

// Reset = 表を全部消してから作り直す。★中のデータも全部消える★
//
// 列の型を変えたときなど、形が合わなくなったときの復旧用。
// 実行方法は docs/SETUP.md にある。
func Reset() error {
	// ★消す順番が大事
	//   表どうしが紐付いている場合、参照している側(子)から先に消す。
	//   逆にすると「まだ使われている」と怒られて消せない。
	//
	//   "demo_todos" はデモ用の表。デモは本番のどの表とも紐付いていないので
	//   順番は関係ない。無ければ何も起きないので、そのままでよい。
	//   internal/demo/ を削除したら、この "demo_todos" も消してよい。
	if err := DB.Migrator().DropTable("demo_todos", &models.User{}); err != nil {
		return err
	}
	return Migrate()
}
