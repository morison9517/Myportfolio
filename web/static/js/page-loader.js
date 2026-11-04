/* page-loader.js = 読み込み画面を閉じる合図(見た目は page-loader.css)

   次の3つがそろったら閉じる。

     ページの読み込み完了 … 背景画像なども含む(window の load)
     フォントの読み込み完了 … 待たないと、Welcome の文字が表示後に別の形へ入れ替わる
     最低表示時間           … 波が WAVE_ROUNDS 回流れるまでは、どれだけ速くても閉じない

   ★回線が遅くても待たせ続けないよう、MAX_MS で必ず閉じる。
     波は動き続ける表現なので、WCAG 2.2.2 の目安(5秒)にも合わせてある。 */
(function () {
    "use strict";

    /* 待つ波の回数。1.5 なら「1回流れ切って、次の波が半分進むまで」。 */
    const WAVE_ROUNDS = 1.2;
    const MIN_MS = 1000;
    const MAX_MS = 5000;

    const html = document.documentElement;
    const loader = document.getElementById("page-loader");

    if (!loader) {
        return;
    }

    if (!html.classList.contains("page-loading")) {
        loader.remove();
        return;
    }

    /* このスクリプトが動いたので、<head> の保険はもう要らない。 */
    clearTimeout(window.pageLoaderFallbackTimer);

    /* 読み込み中は、隠れている部品へキーボードで移動できないようにする。 */
    const hidden = document.querySelectorAll(".skip-link, .site-header, main, footer");

    hidden.forEach((element) => {
        element.inert = true;
    });

    const wait = (ms) => new Promise((resolve) => setTimeout(resolve, Math.max(0, ms)));

    const pageLoaded = new Promise((resolve) => {
        if (document.readyState === "complete") {
            resolve();
        } else {
            window.addEventListener("load", resolve, { once: true });
        }
    });

    /* wf-loading は base.html の Adobe Fonts のスクリプトが、終わったときに外す。 */
    const webFontsLoaded = new Promise((resolve) => {
        if (!html.classList.contains("wf-loading")) {
            resolve();
            return;
        }

        const observer = new MutationObserver(() => {
            if (!html.classList.contains("wf-loading")) {
                observer.disconnect();
                resolve();
            }
        });

        observer.observe(html, { attributes: true, attributeFilter: ["class"] });
    });

    /* ★時間は「実際に画面へ描かれてから」数える。

         requestAnimationFrame は画面を描くときに呼ばれ、ブラウザが裏で先読みしている間や
         別のタブを見ている間は呼ばれない。これを待たないと、アドレスバーから開いたときに
         映る前に処理が終わり、映った瞬間に読み込み画面が消えたように見える。 */
    const nextFrame = () => new Promise((resolve) => requestAnimationFrame(resolve));

    /* 波が指定の回数ぶん流れるまでの時間。

         終わった波の数 × 1周の時間 + 今の波の進み具合 × 波が右端まで流れ切る時間
         波が右端まで流れ切る時間 = 最後の文字の開始の遅れ + 最後の文字が動いている時間

       ★「最後の文字が全周を終えるまで」にしないこと。1周には波の後の休みも含まれ、
         さらに最後の文字は遅れて始まるので、先頭の文字が次の波に入ってしまう。
       ★数字を直接書かず、CSS の値を読む。波の速さを変えても自動で追従する。
       ★動きを減らす設定では波が無いので 0 を返し、MIN_MS だけ待つ。 */
    const waveDuration = () => {
        const last = loader.querySelector(".page-loader__wave span:last-child");
        const animation = last ? last.getAnimations()[0] : null;

        if (!animation) {
            return 0;
        }

        const timing = animation.effect.getComputedTiming();
        const keyframes = animation.effect.getKeyframes();

        /* 最後から2番目のキーフレームが、文字が動き終えて休みに入る位置(CSS の 36%)。 */
        const movingRatio = keyframes[keyframes.length - 2].computedOffset;

        const finishedWaves = Math.ceil(WAVE_ROUNDS) - 1;
        const currentWave = WAVE_ROUNDS - finishedWaves;
        const sweep = timing.delay + timing.duration * movingRatio;

        return finishedWaves * timing.duration + currentWave * sweep;
    };

    nextFrame().then(() => {
        const ready = Promise.all([
            pageLoaded,
            webFontsLoaded,
            document.fonts ? document.fonts.ready : Promise.resolve(),
            wait(Math.max(MIN_MS, waveDuration())),
        ]);

        return Promise.race([ready, wait(MAX_MS)]);
    }).then(nextFrame).then(() => {
        hidden.forEach((element) => {
            element.inert = false;
        });

        loader.classList.add("is-leaving");
        html.classList.remove("page-loading");

        /* ★transitionend は動かなかったときに来ないので、時間でも必ず片付ける。
             描画が遅い環境で動きの途中に消さないよう、動く時間より十分長くしてある。 */
        const remove = () => loader.remove();

        loader.addEventListener("transitionend", (event) => {
            if (event.target === loader && event.propertyName === "transform") {
                remove();
            }
        });

        setTimeout(remove, 4000);
    });
})();
