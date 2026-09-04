/* ===========================================================================
   admin.js = 管理ページの確認ポップアップ(削除 / 返信の送信)

   ▼ 読み込み場所
     pages/admin.html と pages/admin_reply.html の {{ define "scripts" }}。

   ★2つの画面で同じファイルを使い回している。
     それぞれの処理は「その画面に部品があるか」を確かめてから動くので、
     片方の画面にしか無い部品でも問題ない。

   ▼ 流れ

     「削除」を押す
       → 送信を止めて、確認のポップアップを開く
       → 「削除する」   … ポップアップを閉じ、本当に送信する
       → 「キャンセル」「×」「Escキー」 … 閉じるだけ。何も起きない

   ▼ ★ポップアップは1つだけ用意して使い回している

     問い合わせの件数ぶんポップアップを作ると、同じHTMLが何十個も並ぶ。
     「今どれを消そうとしているか」だけを覚えておけば1つで足りる。
   =========================================================================== */

const deleteDialog = document.getElementById("delete-confirm");
const deleteForms = document.querySelectorAll(".inquiry-delete");

if (deleteDialog && deleteForms.length > 0) {
  setupDeleteConfirm();
}

const replyDialog = document.getElementById("reply-confirm");
const replyForm = document.querySelector(".reply-form");

if (replyDialog && replyForm) {
  setupReplyConfirm();
}

/* ===========================================================================
   返信の送信の確認

   ▼ なぜ必要か
     押した瞬間にメールが飛び、取り消せない。
     指が触れただけで送ってしまう事故を、一段挟んで防ぐ。

   ▼ ★未入力チェックはブラウザに任せている
     送信ボタンは type="submit" のままなので、
     返信欄が空なら submit イベント自体が起きず、ここには来ない。
   =========================================================================== */
function setupReplyConfirm() {
  const executeButton = document.getElementById("reply-execute");

  // 送信中かどうか。連打で同じ返信が何通も飛ぶのを防ぐ。
  let sending = false;

  /* ★<html> に modal-open を付けているのはスクロールを止めるため。
       <dialog> は後ろを押せなくはしてくれるが、スクロールは止まらない。
       見た目は base.css と contact.css(.modal)。 */
  replyDialog.addEventListener("close", () => {
    document.documentElement.classList.remove("modal-open");
  });

  // × と キャンセル。閉じるだけで何も起きない。
  replyDialog.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => replyDialog.close());
  });

  replyForm.addEventListener("submit", (event) => {
    // ブラウザの「そのまま送信する」動きを止める。
    event.preventDefault();

    if (sending) {
      return;
    }

    document.documentElement.classList.add("modal-open");
    replyDialog.showModal();
  });

  executeButton.addEventListener("click", () => {
    if (sending) {
      return;
    }
    sending = true;

    replyDialog.close();

    /* ★form.submit() を使う(requestSubmit ではない)。
         submit() は submit イベントを起こさずに送信するので、
         上の「送信を止めてポップアップを出す」処理に戻ってこない。
         requestSubmit() だとイベントが起きて、永久にポップアップが出続ける。 */
    replyForm.submit();
  });
}

function setupDeleteConfirm() {
  const message = document.getElementById("delete-confirm-message");
  const executeButton = document.getElementById("delete-execute");

  // 今どの問い合わせを消そうとしているか。
  let targetForm = null;

  /* ★<html> に modal-open を付けているのはスクロールを止めるため。
       <dialog> は後ろを押せなくはしてくれるが、スクロールは止まらない。
       見た目は base.css と contact.css(.modal)。 */
  deleteDialog.addEventListener("close", () => {
    document.documentElement.classList.remove("modal-open");

    // 閉じたら覚えていた相手を忘れる。
    // ★これが無いと、キャンセルした相手が残ったままになる。
    targetForm = null;
  });

  // × と キャンセル。閉じるだけで何も起きない。
  deleteDialog.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => deleteDialog.close());
  });

  deleteForms.forEach((form) => {
    form.addEventListener("submit", (event) => {
      // ブラウザの「そのまま送信する」動きを止める。
      event.preventDefault();

      targetForm = form;

      // 誰の問い合わせを消すのかを出す(HTML側の data-name)。
      message.textContent = `${form.dataset.name} 様のお問い合わせを削除します。`;

      document.documentElement.classList.add("modal-open");
      deleteDialog.showModal();
    });
  });

  executeButton.addEventListener("click", () => {
    if (!targetForm) {
      return;
    }

    /* ★close() より先に控えを取る。
         close() で上の close イベントが動き、targetForm が空になるため。 */
    const form = targetForm;

    deleteDialog.close();

    /* ★form.submit() を使う(requestSubmit ではない)。
         submit() は submit イベントを起こさずに送信するので、
         上の「送信を止めてポップアップを出す」処理に戻ってこない。
         requestSubmit() だとイベントが起きて、永久にポップアップが出続ける。 */
    form.submit();
  });
}
