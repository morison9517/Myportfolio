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
        const result = await api.post("/api/contact", {
          name: document.getElementById("contact-name").value,
          email: document.getElementById("contact-email").value,
          body: document.getElementById("contact-body").value,
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
