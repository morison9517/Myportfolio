/* layout-motion.js = 画面幅が変わって並びが切り替わるときの動き

   @container による列数の切り替えは一瞬で、CSSの transition では補間できない
   (flex-direction や列数には途中の値が無いため)。そこで次の順に動かす。

     1. 各要素の、.container から見た位置を覚えておく
     2. 幅が変わったら測り直し、大きく動いた要素を見つける
     3. 元の位置から今の位置へ、transform で滑らせる

   ★ResizeObserver は「配置が決まった後、画面を描く前」に呼ばれるので、
     新しい配置が一瞬見えてしまうことがない。

   ★動かすのは transform だけ。配置そのものはCSSの結果のままなので、
     このファイルが読み込めなくても表示は崩れない(切り替えが一瞬になるだけ)。 */
(function () {
    "use strict";

    if (!("ResizeObserver" in window) || !Element.prototype.animate) {
        return;
    }

    const TARGETS = ".about-image, .about-text, .skill-card, .product-card, .skill-item";

    /* 横幅も一緒に補間する要素。文字が多い要素に使うと伸び縮みが目立つので絞っている。 */
    const SCALED = ".product-card";

    /* これより小さい移動は、幅に比例した普通の変化として動かさない。 */
    const JUMP_PX = 24;

    const DURATION_MS = 350;
    const EASING = "cubic-bezier(0.2, 0.8, 0.2, 1)";

    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

    const containers = new Set();
    const items = [];

    document.querySelectorAll(TARGETS).forEach((element) => {
        const container = element.closest(".container");

        if (container) {
            containers.add(container);
            items.push({ element, container, scaled: element.matches(SCALED) });
        }
    });

    if (items.length === 0) {
        return;
    }

    /* ★getBoundingClientRect ではなく offset 系で測る。
         あちらは動いている最中の transform を含むので、再生中に測ると値がずれる。 */
    const offsetFromPage = (element) => {
        let x = 0;
        let y = 0;

        for (let node = element; node; node = node.offsetParent) {
            x += node.offsetLeft;
            y += node.offsetTop;
        }

        return { x, y };
    };

    /* ★位置は要素の中心で持つ。transform の基準点が中心なので、拡大縮小しても計算が合う。 */
    const measure = ({ element, container }) => {
        const self = offsetFromPage(element);
        const base = offsetFromPage(container);

        return {
            x: self.x - base.x + element.offsetWidth / 2,
            y: self.y - base.y + element.offsetHeight / 2,
            width: element.offsetWidth,
            height: element.offsetHeight,
        };
    };

    const play = (item, from, to) => {
        const dx = from.x - to.x;
        const dy = from.y - to.y;

        if (Math.hypot(dx, dy) < JUMP_PX || to.width === 0 || to.height === 0) {
            return;
        }

        const scaleX = item.scaled ? from.width / to.width : 1;
        const scaleY = item.scaled ? from.height / to.height : 1;

        /* ★composite: "add" … 再生中にまた切り替わっても、動きを足し合わせて途切れさせない。 */
        item.element.animate(
            [
                { transform: `translate(${dx}px, ${dy}px) scale(${scaleX}, ${scaleY})` },
                { transform: "translate(0px, 0px) scale(1, 1)" },
            ],
            { duration: DURATION_MS, easing: EASING, composite: "add" }
        );
    };

    const lastRects = new Map();
    const lastWidths = new Map();

    const observer = new ResizeObserver((entries) => {
        /* ★幅が変わったときだけ動かす。高さだけの変化はフォントの読み込みなどで起きる。 */
        const widthChanged = entries.some((entry) => {
            const width = entry.contentRect.width;
            const previous = lastWidths.get(entry.target);

            lastWidths.set(entry.target, width);

            return previous !== undefined && previous !== width;
        });

        items.forEach((item) => {
            const next = measure(item);
            const previous = lastRects.get(item.element);

            if (widthChanged && previous && !reducedMotion.matches) {
                play(item, previous, next);
            }

            lastRects.set(item.element, next);
        });
    });

    containers.forEach((container) => observer.observe(container));
})();
