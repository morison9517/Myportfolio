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

applyTheme();

window.addEventListener("pageshow", applyTheme);

themeSwitch.addEventListener("change", () => {
    const isDark = themeSwitch.checked;

    document.body.classList.toggle("dark", isDark);

    localStorage.setItem(
        "theme",
        isDark ? "dark" : "light"
    );
});