// =====================================================================
// Yemekhane · api.js
// Bütün sayfaların ortak yardımcıları. Arayüz veriyi yalnızca buradaki
// istek() fonksiyonuyla /api/ üzerinden alır; iş kuralı burada değil,
// API'de ve veritabanındadır.
// =====================================================================

// istek("/menu?...") → JSON. Hata durumunda API'nin {"hata": "..."} mesajı
// Error olarak fırlatılır; sayfalar bu mesajı olduğu gibi gösterir.
export async function istek(yol, { method = "GET", govde } = {}) {
    const secenek = { method, headers: { Accept: "application/json" } };
    if (govde !== undefined) {
        secenek.headers["Content-Type"] = "application/json";
        secenek.body = JSON.stringify(govde);
    }

    let yanit;
    try {
        yanit = await fetch("/api" + yol, secenek);
    } catch {
        throw new Error("Sunucuya ulaşılamadı. İnternet bağlantınızı kontrol edin.");
    }

    // Oturum yoksa ya da süresi dolduysa giriş sayfasına git
    if (yanit.status === 401 && !location.pathname.endsWith("login.html")) {
        location.href = "login.html";
        throw new Error("Lütfen giriş yapın.");
    }

    const veri = await yanit.json().catch(() => null);
    if (!yanit.ok) {
        const hata = new Error(veri?.hata ?? "Beklenmeyen bir hata oluştu (HTTP " + yanit.status + ").");
        hata.durum = yanit.status; // ör. 404 → "menü yok" ayrı gösterilebilsin
        throw hata;
    }
    return veri;
}

// ---------- Biçimlendirme ----------

const paraBicimi = new Intl.NumberFormat("tr-TR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// tl(42.5) → "42,50 TL"
export const tl = (sayi) => paraBicimi.format(sayi) + " TL";

// gunMetni("2026-11-16") → "Pazartesi 16 Kasım"
export function gunMetni(tarih) {
    const [y, a, g] = tarih.split("-").map(Number);
    return new Date(y, a - 1, g).toLocaleDateString("tr-TR", { weekday: "long", day: "numeric", month: "long" });
}

// saatMetni("2026-11-14T07:40:01+03:00") → "14.11.2026 07:40"
export function saatMetni(zaman) {
    return new Date(zaman).toLocaleString("tr-TR", {
        day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit",
    });
}

// bugun() → "2026-11-24" (tarayıcının yerel saatine göre)
export function bugun() {
    const t = new Date();
    return [t.getFullYear(), String(t.getMonth() + 1).padStart(2, "0"), String(t.getDate()).padStart(2, "0")].join("-");
}

// ---------- DOM ----------

// el("li", { class: "x" }, "metin", baskaEleman) → <li class="x">metin…</li>
// Metinler textContent olarak eklenir; kullanıcı yorumu gibi veriler HTML
// olarak yorumlanmaz (XSS'e karşı).
export function el(etiket, ozellikler = {}, ...cocuklar) {
    const e = document.createElement(etiket);
    for (const [ad, deger] of Object.entries(ozellikler)) {
        if (deger === false || deger == null) continue;
        if (ad.startsWith("on")) e.addEventListener(ad.slice(2), deger);
        else if (deger === true) e.setAttribute(ad, "");
        else e.setAttribute(ad, deger);
    }
    for (const c of cocuklar.flat()) {
        if (c == null || c === false) continue;
        e.append(c instanceof Node ? c : document.createTextNode(String(c)));
    }
    return e;
}

// mesaj(p, "Bakiye yetersiz.")            → kırmızı kutu
// mesaj(p, "Rezervasyon yapıldı.", "basari") → yeşil kutu
// mesaj(p, null)                          → kutuyu gizler
export function mesaj(kutu, metin, tur = "hata") {
    if (!metin) {
        kutu.hidden = true;
        return;
    }
    kutu.textContent = metin;
    kutu.className = "bildirim bildirim-" + tur;
    kutu.setAttribute("role", tur === "hata" ? "alert" : "status");
    kutu.hidden = false;
}

// <select> seçeneklerini doldurur: secenekler([{id, ad}], seçili id)
export function secenekleriDoldur(select, liste, secili) {
    select.replaceChildren(...liste.map((x) => el("option", { value: x.id, selected: String(x.id) === String(secili) }, x.ad)));
}

// ---------- Her sayfada ortak ----------

// Çıkış bağlantısı: oturumu sunucuda (Redis) siler, giriş sayfasına döner
document.getElementById("cikis")?.addEventListener("click", async (olay) => {
    olay.preventDefault();
    await istek("/cikis", { method: "POST" }).catch(() => {});
    location.href = "login.html";
});

// Giriş yapmış kullanıcı (sayfa başına bir kez sorulur)
let benSozu = null;
export function ben() {
    benSozu ??= istek("/ben");
    return benSozu;
}

// Yönetici olmayan kullanıcıya menüde yönetici bağlantısını gösterme.
// (Asıl yetki kontrolü API'dedir; bu yalnızca görünüm içindir.)
if (!location.pathname.endsWith("login.html")) {
    ben().then((k) => {
        if (k.rol !== "yonetici") {
            document.querySelector('nav a[href="adminPanel.html"]')?.closest("li")?.remove();
        }
    }).catch(() => {});
}
