/* reveal.js = スクロールで画面に入ってきた要素に is-visible を付ける係。
   トップページ専用。どう出すか(向き・待ち時間)は reveal.css 側で決めている。

   動かすかどうかの判断は index.html の <head> が済ませていて、
   結果が <html> の reveal という印になっている。 */

(function () {
    const html = document.documentElement;

    if (!html.classList.contains("reveal")) {
        return;
    }

    /* 見張る仕掛けが無いブラウザでは隠したままになるので、印を外して全部出す。 */
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

    /* ---------- カードを順番に出すセクション ----------
       Skills と Products は作りが同じなので、1つの表で扱う。

       step … カード1枚あたり何秒ずらすか
       more … 一覧の下にある「Read More」

       ★何番目かをHTMLに書かないこと。カードを1枚足すたびに書き直しになる。 */
    const cardGroups = [
        {
            cards: "#skills .skill-card",
            step: 0.06,
            grid: "#skills .skills-grid",
            more: "#skills .skills-more",
        },
        {
            cards: "#products .product-card",
            step: 0.1,
            grid: "#products .products-grid",
            more: "#products .products-more",
        },
    ];

    cardGroups.forEach((group) => {
        const cards = document.querySelectorAll(group.cards);

        cards.forEach((card, index) => {
            card.style.setProperty(
                "--reveal-step",
                `${(index * group.step).toFixed(2)}s`
            );
        });

        /* Read More は自分が画面に入るのを待たず、最後のカードの0.3秒後に続けて出す。 */
        const more = document.querySelector(group.more);

        if (more && cards.length > 0) {
            const lastStep = (cards.length - 1) * group.step;

            more.style.setProperty(
                "--reveal-step",
                `${(lastStep + 0.3).toFixed(2)}s`
            );
        }
    });

    /* ---------- 年表 ----------
       線がマークに届く瞬間にマークを出したいので、
       「項目の高さのうちマークが何割の位置か」を測って渡す。CSSでは計算できない。
       ★出す直前に測ること。先に測ると画面幅が変わったときに古くなる。 */
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

    /* 1件目は待たずに出し、2件目からは前の項目の revealLead 後に出す。
       まとめて入ってきても必ず上から順になる。 */
    const queueTimelineItem = (item) => {
        setMarkProgress(item);

        const now = performance.now();
        const start = Math.max(now, timelineNext);

        item.style.setProperty(
            "--line-delay",
            `${((start - now) / 1000).toFixed(2)}s`
        );

        timelineNext = start + revealLead;
    };

    /* ★ここに書くのは「画面に入ったかを見る単位」。
         隠す対象そのものの一覧は reveal.css にある。 */
    const watched = [
        "#about h2",
        "#about .about-text",

        "#skills h2",
        "#skills .skills-grid",

        "#products h2",
        "#products .products-grid",

        "#experience h2",

        "#contact h2",
        "#contact .contact-form",
    ];

    /* 一覧が出たときに、一緒に出す Read More の対応表。
       ★一覧が無いセクションでは合図の出しようがないので、そのときだけ単独で見張る。 */
    const moreByGrid = new Map();

    cardGroups.forEach((group) => {
        const grid = document.querySelector(group.grid);
        const more = document.querySelector(group.more);

        if (!more) {
            return;
        }

        if (grid) {
            moreByGrid.set(grid, more);
        } else {
            watched.push(group.more);
        }
    });

    const onEnter = (entries, self) => {
        const entering = entries
            .filter((entry) => entry.isIntersecting)
            .map((entry) => entry.target);

        /* 合図が届く順番は決まっていないので、年表だけDOM順に並べ直す。 */
        entering
            .filter((target) => timelineItems.includes(target))
            .sort((a, b) => timelineItems.indexOf(a) - timelineItems.indexOf(b))
            .forEach(queueTimelineItem);

        entering.forEach((target) => {
            target.classList.add("is-visible");

            const more = moreByGrid.get(target);

            if (more) {
                more.classList.add("is-visible");
            }

            self.unobserve(target);
        });
    };

    /* 下(マイナス) … 数字を大きくするほど画面の上のほうで出はじめる。
       上(9999px)   … 画面より上は「入っている」扱い。
                      途中で再読み込みしたとき、上へ戻ると何も無い状態になるのを防ぐ。 */
    const makeObserver = (bottom) =>
        new IntersectionObserver(onEnter, {
            rootMargin: `9999px 0px ${bottom} 0px`,
        });

    const observer = makeObserver("-25%");

    /* 年表は項目の途中から文字が出るので、判定線をその分だけ上げる。 */
    const timelineObserver = makeObserver("-45%");

    document.querySelectorAll(watched.join(", ")).forEach((target) => {
        observer.observe(target);
    });

    timelineItems.forEach((item) => {
        timelineObserver.observe(item);
    });

    /* ★最後に止めること。途中で失敗した場合はタイマーが生きているほうがよい。 */
    clearTimeout(window.revealFallbackTimer);
})();
