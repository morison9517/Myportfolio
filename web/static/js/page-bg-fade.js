/* page-bg-fade.js = 背景画像が切り替わるときのクロスフェード

   どの画像を出すかは background.css のメディアクエリが決める。
   その切り替えは一瞬で、background-image は CSS の transition では補間できない
   (対応していないブラウザがある)。そこで次の順に動かす。

     1. 画面サイズが変わるたびに、CSS が選んだ画像を確かめる
     2. 変わっていたら、それまでの画像を写した板を手前に重ねる
     3. 新しい画像の準備ができたら、手前の板を薄くして消す

   ★ResizeObserver は「配置が決まった後、画面を描く前」に呼ばれるので、
     新しい画像が一瞬見えてしまうことがない。

   ★新しい画像を待ってから薄くするのは、初めて表示する画像だと読み込みに時間がかかるため。
     待たないと、読み込みが終わるまでページの地の色が透けて見えてしまう。

   ★画像の条件はCSSだけに書き、ここには書かない。
     このファイルが読み込めなくても表示は崩れない(切り替えが一瞬になるだけ)。 */
(function () {
    "use strict";

    const board = document.querySelector(".page-bg");

    if (!board || !("ResizeObserver" in window) || !Element.prototype.animate) {
        return;
    }

    const DURATION_MS = 600;
    const EASING = "ease";

    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

    const currentImage = () => window.getComputedStyle(board).backgroundImage;

    let shownImage = currentImage();

    /* いま重ねている板。切り替え中でなければ null。 */
    let outgoing = null;

    /* 新しい画像を読み込み、表示できる状態になるまで待つ。
       ★読み込みに失敗しても止めない。止めると古い画像が残り続けてしまう。 */
    const whenReady = (image) => {
        const url = image.match(/url\("(.+)"\)/);

        if (!url) {
            return Promise.resolve();
        }

        const loader = new Image();
        loader.src = url[1];

        return loader.decode().catch(() => {});
    };

    const crossfade = (fromImage, toImage) => {
        /* ★切り替えの途中でまた切り替わったときは、見た目が飛ばないようにする。
             奥と手前の画像が入れ替わるので、いまの濃さの残りから始めれば同じ見た目になる。 */
        const startOpacity = outgoing
            ? 1 - Number(window.getComputedStyle(outgoing).opacity)
            : 1;

        if (outgoing) {
            outgoing.remove();
        }

        const layer = document.createElement("div");

        layer.className = "page-bg-outgoing";
        layer.style.backgroundImage = fromImage;
        layer.style.opacity = startOpacity;

        board.appendChild(layer);
        outgoing = layer;

        whenReady(toImage).then(() => {
            /* 待っている間に次の切り替えが始まっていたら、この板はもう使わない。 */
            if (outgoing !== layer) {
                return;
            }

            layer.animate(
                [{ opacity: startOpacity }, { opacity: 0 }],
                { duration: DURATION_MS * startOpacity, easing: EASING }
            ).onfinish = () => {
                layer.remove();

                if (outgoing === layer) {
                    outgoing = null;
                }
            };
        });
    };

    /* ★板は画面いっぱいに貼り付いているので、板の大きさ = 画面の大きさ。 */
    new ResizeObserver(() => {
        const nextImage = currentImage();

        if (nextImage === shownImage) {
            return;
        }

        const previousImage = shownImage;
        shownImage = nextImage;

        if (reducedMotion.matches) {
            return;
        }

        crossfade(previousImage, nextImage);
    }).observe(board);
})();
