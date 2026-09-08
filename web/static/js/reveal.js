/* ===========================================================================
   reveal.js = スクロールで画面に入ってきた要素を出す係(トップページ専用)
   ===========================================================================

   ▼ このファイルがやること

     1. 出す対象を見張る(画面に入ったかどうか)
     2. 入ったものに is-visible を付ける

     「どう出すか」(透明から・下から・左から など)は
     すべて reveal.css 側で決めている。ここは合図を出すだけ。

   ▼ ★動かす/動かさないの判断はここではしない

     判断しているのは index.html の <head> のスクリプト。
     結果が <html> の reveal という印になっているので、
     ここでは印の有無を見るだけでよい。

       印がある … / または /# で開いた(スクロールで順に出す)
       印がない … /#about などで直接来た、視差効果を減らす設定

   ▼ ★見張る先が「1枚ずつ」ではないものがある

     Skills と Products はカード1枚ずつではなく、一覧(グリッド)を見ている。
     1枚ずつ見ると、下の行が画面に入るのは上の行よりずっと後になるので、
     「1番目から順番に」の流れが途切れてしまう。
     一覧を見て、中のカードは待ち時間をずらして出す。
   =========================================================================== */

(function () {
    const html = document.documentElement;

    /* 印が無ければ何もしない。要素は隠れていないので、このまま全部見えている。 */
    if (!html.classList.contains("reveal")) {
        return;
    }

    /* ★古いブラウザ向けの保険。
         見張る仕掛けが無いと is-visible を付けられず、
         隠したままになってしまうので、印を外して全部出す。 */
    if (!("IntersectionObserver" in window)) {
        html.classList.remove("reveal");
        return;
    }

    const revealLead = (() => {
        const raw = getComputedStyle(html).getPropertyValue("--reveal-lead").trim();
        const value = parseFloat(raw);

        if (!value) {
            return 200;
        }

        return raw.endsWith("ms") ? value : value * 1000;
    })();

    /* -----------------------------------------------------------------------
       1. 順番に出すもの(1番目から少しずつ遅らせる)
       -----------------------------------------------------------------------
       ★HTMLではなくここでずらす量(--reveal-step)を入れている理由
         カードは後から増えたり減ったりする。
         「何番目か」をHTMLに書くと、1枚足すたびに以降を全部書き直すことになる。
         ここで数えれば、HTMLは何もしなくてよい。

       step … 1枚あたり何秒ずらすか。大きくするとゆっくり順に出る。 */
    const productCardGroup = { selector: "#products .product-card", step: 0.1 };

    const staggerGroups = [
        { selector: "#skills .skill-card", step: 0.06 },
        productCardGroup,
    ];

    staggerGroups.forEach((group) => {
        document.querySelectorAll(group.selector).forEach((card, index) => {
            card.style.setProperty(
                "--reveal-step",
                `${(index * group.step).toFixed(2)}s`
            );
        });
    });

    /* Products の Read More は、一覧が画面に入るのを別に待たず、
       最後の作品の 0.3 秒後に続けて出す。 */
    const productsGrid = document.querySelector("#products .products-grid");
    const productsMore = document.querySelector("#products .products-more");
    const productCards = document.querySelectorAll(productCardGroup.selector);

    if (productsMore && productCards.length > 0) {
        const lastStep = (productCards.length - 1) * productCardGroup.step;

        productsMore.style.setProperty(
            "--reveal-step",
            `${(lastStep + 0.3).toFixed(2)}s`
        );
    }

    /* -----------------------------------------------------------------------
       2. 年表のマーク(丸・破断マーク)の位置を測る
       -----------------------------------------------------------------------
       縦線は項目の上端から下端まで引かれ、マークはその途中にある。
       線がマークに届く瞬間にマークを出したいので、
       「項目の高さのうち、マークが何割の位置か」を測って渡す。

       ★CSSだけでは計算できない。
         マークの位置 --mark-y は「項目の高さ」に対する割合で決まっているが、
         CSSにはそれを待ち時間(秒)に変換する手段が無い。

       ★下余白は padding で取っている(experience.css)。
         マークはその余白を除いた中央にあるので、
         高さから余白を引いて半分にすると、マークまでの距離になる。

       ★出す直前に測ること。
         先にまとめて測ると、途中で画面幅が変わったときに値が古くなる。 */
    const setMarkProgress = (item) => {
        const body = item.querySelector(".timeline-body");

        if (!body) {
            return;
        }

        const height = body.getBoundingClientRect().height;

        if (height <= 0) {
            return;
        }

        const gapBottom = parseFloat(getComputedStyle(body).paddingBottom) || 0;

        item.style.setProperty(
            "--mark-progress",
            ((height - gapBottom) / 2 / height).toFixed(3)
        );
    };

    const timelineItems = Array.from(
        document.querySelectorAll("#experience .timeline-item")
    );

    let timelineNext = 0;

    const queueTimelineItem = (item) => {
        setMarkProgress(item);

        const now = performance.now();

        if (timelineNext === 0) {
            timelineNext = now + revealLead * 2;
        }

        const start = Math.max(now, timelineNext);

        item.style.setProperty(
            "--line-delay",
            `${((start - now) / 1000).toFixed(2)}s`
        );

        timelineNext = start + revealLead;
    };

    /* -----------------------------------------------------------------------
       3. 見張る要素
       -----------------------------------------------------------------------
       ★ここに書くのは「画面に入ったかを見る単位」。
         隠す対象そのものの一覧は reveal.css にある。
         Skills / Products は一覧を、年表は項目(<li>)を見ている。 */
    const watched = [
        "#about h2",
        "#about .about-text",

        "#skills h2",
        "#skills .skills-grid",
        "#skills .skills-more",

        "#products h2",
        "#products .products-grid",

        "#experience h2",

        "#contact h2",
        "#contact .contact-form",
    ];

    const onEnter = (entries, self) => {
        const entering = entries
            .filter((entry) => entry.isIntersecting)
            .map((entry) => entry.target);

        /* ★年表だけは、まとめて入ってきたときに上から順に並べ直してから
             待ち時間を決める。合図が届く順番は決まっていないので、
             これをしないと下の項目が先に出ることがある。 */
        entering
            .filter((target) => timelineItems.includes(target))
            .sort((a, b) => timelineItems.indexOf(a) - timelineItems.indexOf(b))
            .forEach(queueTimelineItem);

        entering.forEach((target) => {
            target.classList.add("is-visible");

            if (target === productsGrid && productsMore) {
                productsMore.classList.add("is-visible");
            }

            /* 一度出したら見張りを外す。
               ★戻したくないので、外すのが正しい。
                 付けたままだと、上へスクロールして画面から出たときに
                 また合図が来る(is-visible は付いたままなので実害は無いが、
                 見張り続けるぶん無駄になる)。 */
            self.unobserve(target);
        });
    };

    /* ▼ ★上下で意味が違う

         下(マイナス)… 画面の下端からその割合ぶんは「まだ入っていない」ことにする。
                       数字を大きくするほど、画面の上のほうで出はじめる。

         上(9999px) … 画面より上にあるものは「入っている」ことにする。
                       ページの途中で再読み込みすると、
                       そこより上の要素は二度と画面に入らず、
                       上へ戻ったときに何も無い状態になってしまう。
                       見えない場所で先に出しておけば、それを防げる。 */
    const makeObserver = (bottom) =>
        new IntersectionObserver(onEnter, {
            rootMargin: `9999px 0px ${bottom} 0px`,
        });

    const observer = makeObserver("-25%");

    /* 年表は項目の途中(丸の高さ)から文字が出るので、他と同じ判定だと
       出はじめが画面の下に寄りすぎる。判定線をその分だけ上げる。 */
    const timelineObserver = makeObserver("-45%");

    /* 一覧が無ければ合図の出しようがないので、そのときだけ単独で見張る。 */
    if (!productsGrid) {
        watched.push("#products .products-more");
    }

    document.querySelectorAll(watched.join(", ")).forEach((target) => {
        observer.observe(target);
    });

    timelineItems.forEach((item) => {
        timelineObserver.observe(item);
    });

    /* -----------------------------------------------------------------------
       4. 保険のタイマーを止める
       -----------------------------------------------------------------------
       index.html の <head> が「5秒経っても reveal.js が動かなければ
       印を外す」タイマーを仕掛けている。ここまで来たら不要。

       ★止めるのは最後にすること。
         上のどこかで失敗して止まった場合は、タイマーが生きているほうがよい
         (5秒後に印が外れて、隠れていたものが全部出る)。 */
    clearTimeout(window.revealFallbackTimer);
})();
