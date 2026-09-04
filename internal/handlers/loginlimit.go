// =============================================================================
// loginlimit.go = ログインを何度も失敗した相手を、しばらく締め出す仕掛け
//
// ▼ なぜ必要か
//
//	パスワードは「試し放題」だと、時間さえかければいつか当たる(総当たり)。
//	1回ずつは正しい手順なので、パスワードの中身だけでは防げない。
//	「短い時間に何度も失敗している」という回数で止めるのがこの仕組み。
//
// ▼ 決めごと
//
//	5回続けて失敗すると、そのアクセス元は15分間ログインできない。
//	1回でも成功すれば、それまでの失敗は帳消しになる。
//	最後の失敗から15分経つと、失敗の記録そのものが消える。
//
// ▼ ★「アクセス元(IP)ごと」に数えている理由
//
//	「管理者アカウントごと」に数えると、他人がわざと5回間違えるだけで
//	自分が15分間締め出される(いやがらせが成立してしまう)。
//	アクセス元ごとなら、締め出されるのは間違えた側だけで済む。
//
// ▼ ★記録はアプリの中にしか無い
//
//	DBには保存していないので、アプリを再起動すると失敗回数は消える。
//	サーバー1台で動かしているうちはこれで十分。
//	★複数台に増やしたときは、台ごとに別々に数えることになるので、
//	  そのときはDBやRedisなど共通の置き場に移すこと。
//
// =============================================================================
package handlers

import (
	"sync"
	"time"
)

const (
	// 何回間違えたら締め出すか。
	maxLoginAttempts = 5

	// 締め出す長さ。
	loginLockDuration = 15 * time.Minute

	// 失敗の記録を覚えておく長さ。
	// ★これが無いと、何日も前にたまたま1回間違えた記録が残り続け、
	//   久しぶりに来て4回間違えただけで締め出されてしまう。
	loginAttemptWindow = 15 * time.Minute
)

// loginAttempt = あるアクセス元の失敗の記録。
type loginAttempt struct {
	count       int       // 続けて失敗した回数
	lastFailure time.Time // 最後に失敗した時刻
	lockedUntil time.Time // この時刻までログインできない
}

// ★mutex(排他ロック)が要る理由
//
//	同じ瞬間に複数のアクセスが来ると、同じ地図(map)を同時に書き換えることになる。
//	Goのmapは同時書き込みを禁止していて、起きるとアプリごと落ちる。
//	mu.Lock() 〜 Unlock() の間は1人ずつしか通れないようにして防ぐ。
var (
	loginAttemptsMu sync.Mutex
	loginAttempts   = make(map[string]*loginAttempt)
)

// loginLockRemaining = このアクセス元があと何分締め出されているか。
//
// 0 が返れば、締め出されていない(ログインしてよい)。
func loginLockRemaining(ip string) time.Duration {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	attempt, ok := loginAttempts[ip]
	if !ok {
		return 0
	}

	// ★まだ締め出されていない(失敗を数えている途中)。
	//
	//   ここを飛ばして下の time.Until に進めてはいけない。
	//   締め出し前の lockedUntil は「ゼロ値(西暦1年)」なので、
	//   必ず「もう過ぎている」と判定され、数えている途中の記録まで
	//   消してしまう。そうなると回数が5に届かず、締め出しが一度も働かない。
	if attempt.lockedUntil.IsZero() {
		return 0
	}

	remaining := time.Until(attempt.lockedUntil)
	if remaining <= 0 {
		// 締め出しが明けている。記録を消して、また5回から数え直す。
		delete(loginAttempts, ip)
		return 0
	}

	return remaining
}

// recordLoginFailure = 失敗を1回数える。5回目で締め出す。
func recordLoginFailure(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	sweepOldAttempts()

	attempt, ok := loginAttempts[ip]
	if !ok {
		attempt = &loginAttempt{}
		loginAttempts[ip] = attempt
	}

	attempt.count++
	attempt.lastFailure = time.Now()

	if attempt.count >= maxLoginAttempts {
		attempt.lockedUntil = time.Now().Add(loginLockDuration)
	}
}

// clearLoginFailures = 失敗の記録を消す。ログインに成功したときに呼ぶ。
func clearLoginFailures(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	delete(loginAttempts, ip)
}

// remainingLoginAttempts = あと何回間違えたら締め出されるか。
//
// 画面に「あと◯回」と出すために使う。
func remainingLoginAttempts(ip string) int {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	attempt, ok := loginAttempts[ip]
	if !ok {
		return maxLoginAttempts
	}

	remaining := maxLoginAttempts - attempt.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// sweepOldAttempts = 古くなった記録を捨てる。
//
// ★呼び出し側で mu.Lock() 済みであること(この関数自身はロックしない)。
//
//	ロックの中からさらにロックを取ろうとすると、自分の番を自分で待つ形になり
//	そこで永久に止まる。
//
// ★捨てないと、いたずらでIPを変えながら試された分だけ記録が増え続け、
//
//	アプリの使うメモリが際限なく膨らむ。
func sweepOldAttempts() {
	now := time.Now()

	for ip, attempt := range loginAttempts {
		// 締め出し中のものは、明けるまで残す。
		if now.Before(attempt.lockedUntil) {
			continue
		}

		if now.Sub(attempt.lastFailure) > loginAttemptWindow {
			delete(loginAttempts, ip)
		}
	}
}
