/* ===========================================================================
   admin.js = 問い合わせ一覧の「削除」の確認ポップアップ

   ▼ 読み込み場所
     pages/admin.html の {{ define "scripts" }}。

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
