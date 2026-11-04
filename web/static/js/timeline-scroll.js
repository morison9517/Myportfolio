/* timeline-scroll.js = 年表(#experience)の「前へ」「次へ」ボタン

   スクロールそのものはブラウザの標準機能に任せていて、ここはボタンの係だけ。
   このファイルが読み込めなくても、ボタンが出ないだけで年表は読める。 */
(function () {
    "use strict";

    const scroller = document.querySelector(".timeline-scroller");
    const controls = document.querySelector(".timeline-controls");

    if (!scroller || !controls) {
        return;
    }

    const list = scroller.querySelector(".timeline");
    const items = Array.from(scroller.querySelectorAll(".timeline-item"));
    const prev = controls.querySelector("[data-timeline-prev]");
    const next = controls.querySelector("[data-timeline-next]");

    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

    /* ボタン1回で何項目先へ送るか。スマホは画面が狭いので少なくする。
       ★768px は他のCSS(reveal.css など)のスマホの境目とそろえてある。 */
    const narrowScreen = window.matchMedia("(max-width: 768px)");
    const jumpCount = () => (narrowScreen.matches ? 1 : 4);

    const maxScroll = () => scroller.scrollWidth - scroller.clientWidth;

    /* ★1pxの余裕を見る。画面の拡大率によっては端まで行っても小数ぶん届かない。 */
    const update = () => {
        const max = maxScroll();

        controls.hidden = max <= 1;

        /* ボタンを画面全体の中央に置くために、ページのスクロールバーの幅を渡す。 */
        controls.style.setProperty(
            "--scrollbar-width",
            `${window.innerWidth - document.documentElement.clientWidth}px`
        );

        prev.setAttribute("aria-disabled", String(scroller.scrollLeft <= 1));
        next.setAttribute("aria-disabled", String(scroller.scrollLeft >= max - 1));
    };

    /* 各項目を左端にそろえたときの scrollLeft。
       ★押すたびに測ること。項目の幅は画面幅と文字の大きさで変わる。 */
    const stops = () => {
        const origin = scroller.getBoundingClientRect().left - scroller.scrollLeft;
        const padding = parseFloat(getComputedStyle(list).paddingLeft) || 0;

        return items.map(
            (item) => item.getBoundingClientRect().left - origin - padding
        );
    };

    const move = (direction) => {
        const current = scroller.scrollLeft;
        const all = stops();

        /* 進む向きで一番近い項目。ここから jumpCount() - 1 項目先が行き先になる。 */
        const nearest = direction > 0
            ? all.findIndex((stop) => stop > current + 1)
            : all.findLastIndex((stop) => stop < current - 1);

        if (nearest === -1) {
            return;
        }

        /* ★端を超える場合は端の項目で止める。 */
        const index = Math.min(
            Math.max(nearest + direction * (jumpCount() - 1), 0),
            all.length - 1
        );

        scroller.scrollTo({
            left: all[index],
            behavior: reducedMotion.matches ? "auto" : "smooth",
        });
    };

    prev.addEventListener("click", () => move(-1));
    next.addEventListener("click", () => move(1));

    scroller.addEventListener("scroll", update, { passive: true });

    if ("ResizeObserver" in window) {
        new ResizeObserver(update).observe(scroller);
    } else {
        window.addEventListener("resize", update);
    }

    update();
})();
