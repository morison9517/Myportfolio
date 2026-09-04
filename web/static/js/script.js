const themeSwitch =
    document.getElementById("themeSwitch");

themeSwitch.addEventListener("change", () => {
    document.body.classList.toggle("dark");
});

function toggleMenu() {
    const menu = document.querySelector(".menu-links");
    const icon = document.querySelector(".hamburger-icon");

    menu.classList.toggle("open");
    icon.classList.toggle("open");
}

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