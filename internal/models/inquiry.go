// =============================================================================
// models = データの形(DBの表)を決める場所
//
//	1つの struct = DBの1つの表。1行 = 1件のデータ。
//	    type Inquiry struct  →  inquiries 表
//
//	この設計図があるおかげで、SQLを書かずに
//	    database.DB.Find(&inquiries)     → 全件取得
//	    database.DB.Create(&inquiry)     → 1件追加
//	と書ける。
//
// ▼ `gorm:"..."` `json:"..."` の部分(タグ)について
//
//	項目名の右に付いている付箋のようなもの。
//	    gorm: DBの表を作るときの指示(文字数、空を許すか、など)
//	    json: JavaScriptに渡すときの名前
//
// ▼ ★大文字始まりでないと動かない
//
//	項目名が小文字始まりだと、GORMもJSONも画面のHTMLもその項目を読めない。
//	「保存されない」「画面に出ない」の原因のほとんどがこれ。
//
// =============================================================================
package models

import "time"

// Inquiry = 問い合わせ1件分。
//
// 画面(#contact)から送られてきた内容をそのまま残しておく置き場。
// メールが届かなかったときでも、管理ページ(/admin)で中身を確認できる。
type Inquiry struct {
	// primaryKey = この番号で1行を必ず特定できる印。自動で1,2,3…と振られる。
	ID uint `gorm:"primaryKey" json:"id"`

	// size:100 … MySQLでは文字の最大長を決める必要がある
	// not null … 空を許さない
	Name string `gorm:"size:100;not null" json:"name"`

	// 返信先。index を付けているのは「同じ人からの問い合わせ」を
	// 探すときに速くするため。
	Email string `gorm:"size:255;not null;index" json:"email"`

	// ★本文は size ではなく type:text にする。
	//   size(VARCHAR)には長さの上限があり、長文を入れると途中で切れる。
	Body string `gorm:"type:text;not null" json:"body"`

	// ▼ 後から「誰が・いつ送ったのか」を追えるようにしておく。
	//   迷惑メールが大量に来たときに、送信元を絞り込むのに使う。
	//   ★IPv6も入るので45文字。
	RemoteIP string `gorm:"size:45" json:"-"`

	// CreatedAt という名前にしておくと、GORMが保存時に自動で現在時刻を入れる。
	// アプリ側で入れ忘れる事故がなくなる(お決まりの名前)。
	CreatedAt time.Time `json:"created_at"`

	// RepliedAt = 返信を送った日時。まだ返していなければ空。
	//
	// ★*time.Time の * は「値が無い状態を持てる」という意味。
	//   time.Time のままだと、未返信でも「西暦1年1月1日」という
	//   もっともらしい日時が入ってしまい、返信済みと区別できない。
	//
	// ★画面では「空かどうか」だけを見て、未返信/返信済みを出し分けている
	//   (admin.html の {{ if .RepliedAt }})。
	RepliedAt *time.Time `json:"replied_at,omitempty"`

	// ReplyBody = 送った返信の本文。まだ返していなければ空。
	//
	// ★管理ページで「何と答えたか」を後から読み返すために残している。
	//   メールは送ったら手元に残らないので、ここに控えを取っておく。
	//
	// ★size ではなく type:text にする(本文と同じ理由)。
	//   size(VARCHAR)には長さの上限があり、長文を入れると途中で切れる。
	ReplyBody string `gorm:"type:text" json:"reply_body,omitempty"`
}

// TableName = この設計図が対応する表の名前。
//
// 指定しなくてもGORMが自動で "inquiries" と決めてくれるが、
// 名前が勝手に決まると探しにくいので、明示しておく。
func (Inquiry) TableName() string {
	return "inquiries"
}
