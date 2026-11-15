/* photo-viewer.js = 作品の詳細ページで、写真を押したら大きく見せる

   読み込んでいるのは product.html だけ。

   ▼ 仕組み

     写真は <a href="写真のURL"> で包んである。
     JavaScript が動かないときは、押せば写真そのものが開く(ただのリンク)。
     動くときはここでそれを止めて、<dialog id="photo-viewer"> に写真を入れて開く。

   ★<dialog> の showModal() が「後ろを押せなくする・Escで閉じる・
     キーボードの移動先を中に閉じ込める」までやってくれる(contact.js と同じ)。
     スクロールだけは止まらないので、<html> に modal-open を付けて止める(base.css)。 */

const photoViewer = document.getElementById("photo-viewer");

if (photoViewer) {
  setupPhotoViewer();
}

function setupPhotoViewer() {
  const image = photoViewer.querySelector(".photo-viewer-image");
  const caption = photoViewer.querySelector(".photo-viewer-caption");

  // 開く前にキーボードで選んでいた写真。閉じたらそこへ戻す。
  let opener = null;

  document.querySelectorAll(".product-photo-link").forEach((link) => {
    link.addEventListener("click", (event) => {
      /* ★Ctrl+クリックなど「新しいタブで開く」操作は止めない。
           写真だけを別のタブで見たい人の邪魔になる。 */
      if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) {
        return;
      }

      event.preventDefault();

      const thumbnail = link.querySelector("img");

      image.src = link.href;
      image.alt = thumbnail.alt;

      // ★説明文(Caption)が無い写真は、代わりに alt を見出しに出す。
      //   空のままだと、読み上げソフトがポップアップの名前を読めない。
      caption.textContent = link.dataset.caption || thumbnail.alt;

      opener = link;

      document.documentElement.classList.add("modal-open");
      photoViewer.showModal();
    });
  });

  photoViewer.addEventListener("close", () => {
    document.documentElement.classList.remove("modal-open");

    if (opener) {
      opener.focus();
      opener = null;
    }
  });

  photoViewer.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => photoViewer.close());
  });

  /* 写真の外側(薄暗い幕)を押したら閉じる。

     ★幕を押したときも event.target は <dialog> 自身になるので、
       target だけでは「中の余白」と「外の幕」を見分けられない。
       押した位置がポップアップの枠の外かどうかで判定する。 */
  photoViewer.addEventListener("click", (event) => {
    const rect = photoViewer.getBoundingClientRect();

    const inside =
      event.clientX >= rect.left &&
      event.clientX <= rect.right &&
      event.clientY >= rect.top &&
      event.clientY <= rect.bottom;

    if (!inside) {
      photoViewer.close();
    }
  });
}
