/* ===========================================================================
   contact.js = 問い合わせフォーム(#contact)の確認ポップアップと送信

   ▼ 読み込み場所
     pages/index.html の {{ define "scripts" }}。
     main.js より後に読み込まれるので、api / showMessage / withBusy が使える。

   ▼ 流れ

     「確認」を押す
       → 未入力があればブラウザがここで止める(submitイベントすら起きない)
       → 注意事項のポップアップを開く
       → 「送信」  … ポップアップを閉じて送信する
                     終わったら画面上部にお知らせを出し、入力欄を空にする
       → 「キャンセル」「×」「Escキー」 … 閉じるだけ。入力欄はそのまま

   ▼ ★ページを切り替えずに送っている理由

     普通のフォーム送信はページが切り替わるので、
     1枚もののこのサイトでは画面がいちばん上に戻ってしまう。
     入力欄の位置に留まったまま結果を出したいので、
     送信を止めて(preventDefault)、その場で送っている。
   =========================================================================== */

const contactForm = document.getElementById("contact-form");
const confirmDialog = document.getElementById("contact-confirm");

/* --- スパム対策(reCAPTCHA v3)------------------------------------------------

   ★v3 はチェックボックスを出さない。
     送信の直前に Google から合言葉(トークン)をもらい、
     問い合わせと一緒に送る。本物かどうかはGin側が確かめる。

   ★鍵が無いときは空文字を返す。
     .env に RECAPTCHA_SITE_KEY を入れる前でもフォームは動く
     (Gin側も鍵が無ければ検証しない)。

   ★トークンは2分で切れる。
     だから確認ポップアップを開いた時ではなく、
     「送信」を押した直後に取る。先に取っておくと、
     利用者が確認画面で悩んでいる間に切れてしまう。 */
async function recaptchaToken() {
  const siteKey = window.RECAPTCHA_SITE_KEY;

  if (!siteKey || typeof grecaptcha === "undefined") {
    return "";
  }

  /* grecaptcha.ready は「準備できたら呼ぶ」形なので、
     await で待てるように Promise に包み直している。 */
  return new Promise((resolve) => {
    grecaptcha.ready(async () => {
      try {
        /* action は Gin側の recaptchaAction と同じ文字にすること。
           食い違うと全部弾かれる。 */
        resolve(await grecaptcha.execute(siteKey, { action: "contact" }));
      } catch (error) {
        /* ★取れなくても送信自体は止めない。
             通信が不安定なだけで問い合わせを諦めさせるのは行き過ぎ。
             空で送れば、Gin側が「確かめられなかった」として扱う。 */
        console.warn("reCAPTCHA のトークンを取得できませんでした", error);
        resolve("");
      }
    });
  });
}

/* --- バッジをフッターの最上部へ移す -----------------------------------------

   ★なぜJavaScriptが要るのか
     Googleのスクリプトは、バッジを <body> の直下に勝手に作る。
     どこに作るかを指定する方法が無いので、
     出来上がったものを後からこちらの箱へ移している。

   ★いつ現れるか分からないので MutationObserver で見張る
     バッジはページ読み込みの後、遅れて作られる。
     その瞬間が読めないため、<body> に要素が足された合図を受け取って、
     現れていたら移す、という形にしている。

   ★失敗しても送信は止まらない
     バッジは表示の話で、合言葉(トークン)の取得とは別。
     移せなくても既定の右下に出るだけで、フォームは普通に動く。

   ★もしバッジの表示が崩れたら
     要素を移すとき、中のiframeは読み込み直しになる。
     ここが原因で不具合が出るようなら、移すのをやめて
     「バッジを消して、代わりに注記を置く」形に替えるのが確実
     (Googleが認めている方法。文面はGoogleの案内にある)。 */
function placeRecaptchaBadge() {
  const slot = document.querySelector(".footer-recaptcha");

  if (!slot || !window.RECAPTCHA_SITE_KEY) {
    return;
  }

  // 見つかったら移す。移したかどうかを返す。
  const move = () => {
    const badge = document.querySelector(".grecaptcha-badge");

    if (!badge || slot.contains(badge)) {
      return false;
    }

    slot.appendChild(badge);
    return true;
  };

  // もう出来ていれば、見張らずに済ませる。
  if (move()) {
    return;
  }

  const observer = new MutationObserver(() => {
    if (move()) {
      observer.disconnect();
    }
  });

  observer.observe(document.body, { childList: true });

  /* ★見張りっぱなしにしない。
       鍵が間違っているなどでバッジが最後まで作られないと、
       ページを開いている間ずっと監視が動き続けてしまう。 */
  setTimeout(() => observer.disconnect(), 15000);
}

placeRecaptchaBadge();

/* ★このファイルはトップページでしか読み込まないが、
     将来ほかのページで読み込んでも落ちないように確かめておく。
     フォームが無い画面で addEventListener すると、そこで止まる。 */
if (contactForm && confirmDialog) {
  setupContactForm();
}

function setupContactForm() {
  const confirmButton = contactForm.querySelector("button[type=submit]");
  const sendButton = document.getElementById("contact-send");

  // 送信中かどうか。連打で同じ問い合わせが何件も届くのを防ぐ。
  let sending = false;

  /* --- ポップアップの開け閉め --------------------------------------------
     ★<html> に modal-open を付けているのはスクロールを止めるため。
       <dialog> は後ろを押せなくはしてくれるが、スクロールは止まらない。
       見た目は base.css。 */

  confirmDialog.addEventListener("close", () => {
    document.documentElement.classList.remove("modal-open");
  });

  /* data-close が付いているボタン(× と キャンセル)。
     ★何も起こさずに閉じるだけ。入力欄はそのまま残る。 */
  confirmDialog.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => confirmDialog.close());
  });

  /* --- 「確認」を押したとき ----------------------------------------------
     ★submit イベントは、required が全部埋まっているときしか起きない。
       未入力チェックをブラウザに任せられるので、ここでは何もしなくてよい。 */
  contactForm.addEventListener("submit", (event) => {
    // ブラウザの「送信してページを切り替える」動きを止める。
    event.preventDefault();

    if (sending) {
      return;
    }

    document.documentElement.classList.add("modal-open");
    confirmDialog.showModal();
  });

  /* --- ポップアップの「送信」を押したとき -------------------------------- */
  sendButton.addEventListener("click", async () => {
    if (sending) {
      return;
    }
    sending = true;

    // 先にポップアップを閉じる。
    confirmDialog.close();

    // withBusy = 送信中はボタンを押せなくする(二重送信を防ぐ)。
    await withBusy(confirmButton, async () => {
      try {
        // ★整理券(CSRFトークン)は api が自動で付ける。
        //   こちらの token はスパム対策(reCAPTCHA)用で別物。
        const result = await api.post("/api/contact", {
          name: document.getElementById("contact-name").value,
          email: document.getElementById("contact-email").value,
          body: document.getElementById("contact-body").value,
          token: await recaptchaToken(),
        });

        // 結果は画面上部のお知らせで伝える(見た目は base.css の .flash)。
        showMessage(result.message, "success");

        // 送れたら入力欄を空に戻す。
        contactForm.reset();
      } catch (error) {
        /* ★失敗したときは入力欄を消さない。書き直しになってしまう。
             文言は Gin側が {"error": "..."} で返したものがそのまま入っている。 */
        showMessage(error.message, "error");
      }
    });

    sending = false;
  });
}
