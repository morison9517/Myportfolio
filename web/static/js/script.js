const themeSwitch =
    document.getElementById("themeSwitch");

/* ポップアップ(<dialog>)が開いている間も、テーマのトグルを隠さない。

   <dialog> はブラウザの最前面の層に出るので、z-index では越えられない。
   開いている間だけ、トグルもその層(popover)へ上げる。

   ★<dialog> の中へ移すのは、外に置くと見えていても操作できないため。
     開いた <dialog> の外側はブラウザが押せなくする。
   ★最前面の層に出た要素は、<dialog> の位置や透明度の影響を受けず、
     画面に対して同じ位置に出る。 */
if ("showPopover" in HTMLElement.prototype) {
    const toggle = themeSwitch.closest(".theme-switch");
    const homeParent = toggle.parentNode;
    const homeNext = toggle.nextSibling;

    document.querySelectorAll("dialog").forEach((dialog) => {
        new MutationObserver(() => {
            if (dialog.open) {
                dialog.append(toggle);
                toggle.popover = "manual";
                toggle.showPopover();
            } else if (dialog.contains(toggle)) {
                toggle.hidePopover();
                toggle.removeAttribute("popover");
                homeParent.insertBefore(toggle, homeNext);
            }
        }).observe(dialog, { attributes: true, attributeFilter: ["open"] });
    });
}

themeSwitch.addEventListener("change", () => {
    document.body.classList.toggle("dark");
});

/* Go側(middleware.Flash)が描いたお知らせも、少し経ったら消す。

   ★JavaScriptから出したもの(main.js の showMessage)は自分で消えるが、
     こちらはHTMLとして届くので、何もしないと画面に貼り付いたまま残る。
     見た目は base.css の .flash-list。 */
document.querySelectorAll(".flash-list .flash").forEach((item) => {
    setTimeout(() => item.remove(), 8000);
});

/* フッターの高さを測って、CSSから使えるようにする。

   ▼ 何のためか

     #contact を「フッターと合わせて画面ちょうど1枚分」にしたい。
     そのためには画面の高さからフッターの高さを引く必要があるが、
     CSSには「あの要素の高さ」を知る手段が無いので、ここで測って渡す。
     使っているのは contact.css の min-height。

   ▼ ★決め打ちにしない理由

     フッターはクレジット文の折り返し方で高さが変わる。
     画面幅を変えるだけで数十pxずれるので、実測でないと合わない。

   ★ResizeObserver = 「この要素の大きさが変わったら教えて」という仕掛け。
     画面の回転、ウィンドウの幅変更、フォントの読み込み完了など、
     高さが変わるきっかけを一通り拾ってくれる。
     window の resize だけを見ていると取りこぼす。 */
const siteFooter = document.querySelector("footer");

if (siteFooter) {
    const updateFooterHeight = () => {
        /* ★offsetHeight ではなく getBoundingClientRect を使う。
             offsetHeight は整数に丸めた値しか返さないので、
             実際が 250.6px でも 251px と答えてしまう。 */
        const height = siteFooter.getBoundingClientRect().height;

        /* ★実際より2px小さく伝える。

             ぴったりに合わせると、端数の丸め方の違いで
             #contact がわずかに足りず、1つ前のセクション(#experience)の
             背景画像が画面の上に細く覗いてしまう。

             少なめに伝えると #contact がその分だけ高くなる。
             はみ出す側は「フッターが画面の下にわずかに隠れる」だけなので、
             見た目の問題にならない。隙間ができない側に倒しておく。 */
        document.documentElement.style.setProperty(
            "--footer-height",
            `${Math.floor(height) - 2}px`
        );
    };

    updateFooterHeight();

    /* ★#contact の高さはフッターの高さから決まるが、
         フッターの高さは #contact に影響されない。
         一方通行なので、測り直しが延々と続くことはない。 */
    new ResizeObserver(updateFooterHeight).observe(siteFooter);
}

const menuToggle = document.getElementById("menuToggle");
const menuPanel = document.getElementById("site-menu");

function toggleMenu() {
    const isOpen = menuPanel.classList.toggle("open");

    menuToggle.classList.toggle("open", isOpen);

    /* 開閉の状態と、次に押すと何が起きるかを読み上げに伝える。 */
    menuToggle.setAttribute("aria-expanded", String(isOpen));
    menuToggle.setAttribute("aria-label", isOpen ? "メニューを閉じる" : "メニューを開く");
}

/* Esc で閉じて、開いたボタンにフォーカスを戻す。 */
document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && menuPanel.classList.contains("open")) {
        toggleMenu();
        menuToggle.focus();
    }
});

/* Tab でメニューの外へ出たら閉じる。開いたままだと、次の部品がメニューの裏に隠れる。
   ★relatedTarget が空のとき(マウスで余白を押したときなど)は閉じない。 */
document.getElementById("hamburger-nav").addEventListener("focusout", (event) => {
    const next = event.relatedTarget;

    if (next && !event.currentTarget.contains(next) && menuPanel.classList.contains("open")) {
        toggleMenu();
    }
});

function applyTheme() {
    const savedTheme = localStorage.getItem("theme");

    if (savedTheme === "dark") {
        document.body.classList.add("dark");
        themeSwitch.checked = true;
    } else {
        document.body.classList.remove("dark");
        themeSwitch.checked = false;
    }
}

/* ページを開いたときのテーマ当てはフェードさせない。

   切り替えスイッチを押したときはフェードしてほしいが、
   ここは「もともとダークだった人に元の見た目を戻す」だけなので、
   フェードすると毎回「白→黒」が見えてしまう。

   ★印(theme-loading)を付けている間だけ transition が止まる(base.css)。
   ★offsetHeight を読んでいるのは、ここで一度スタイルを確定させるため。
     これが無いと「darkを付ける」と「印を外す」がまとめて処理され、
     結局フェードが起きてしまう。 */
function applyThemeWithoutFade() {
    document.body.classList.add("theme-loading");

    applyTheme();

    document.body.offsetHeight;

    document.body.classList.remove("theme-loading");
}

applyThemeWithoutFade();

window.addEventListener("pageshow", applyThemeWithoutFade);

themeSwitch.addEventListener("change", () => {
    const isDark = themeSwitch.checked;

    document.body.classList.toggle("dark", isDark);

    localStorage.setItem(
        "theme",
        isDark ? "dark" : "light"
    );
});