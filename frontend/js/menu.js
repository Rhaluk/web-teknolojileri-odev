// Haftanın menüsü ve rezervasyon: GET /api/menuler/haftalik, POST /api/rezervasyonlar
import { istek, el, mesaj, tl, gunMetni, bugun, secenekleriDoldur } from "./api.js";

const secim = document.getElementById("yemekhane-secim");
const tbody = document.querySelector("#weekly-menu-table tbody");
const mesajKutusu = document.getElementById("rezervasyon-mesaj");
const haftaBos = document.getElementById("hafta-bos");
const aralik = document.getElementById("hafta-araligi");

// Haftanın pazartesisi (YYYY-AA-GG) ve ona gün ekleme
function pazartesi(tarih) {
    const [y, a, g] = tarih.split("-").map(Number);
    const t = new Date(y, a - 1, g);
    t.setDate(t.getDate() - ((t.getDay() + 6) % 7));
    return metin(t);
}
function gunEkle(tarih, gun) {
    const [y, a, g] = tarih.split("-").map(Number);
    return metin(new Date(y, a - 1, g + gun));
}
const metin = (t) => [t.getFullYear(), String(t.getMonth() + 1).padStart(2, "0"), String(t.getDate()).padStart(2, "0")].join("-");

let hafta = pazartesi(bugun());

// Son sütun: rezervasyonu varsa rozet, alınabiliyorsa düğme, değilse pasif düğme.
// Alınabilir mi kararını API verir (son saat + boş yer); arayüz yalnızca gösterir.
function islemHucresi(m) {
    if (m.rezervasyon_durumu === "aktif") {
        return el("span", { class: "rozet rozet-aktif" }, m.porsiyon === "yarim" ? "rezervasyonun var (yarım)" : "rezervasyonun var");
    }
    if (m.rezervasyon_durumu === "kullanildi") {
        return el("span", { class: "rozet rozet-kullanildi" }, "kullanıldı");
    }
    if (!m.rezerve_edilebilir) {
        return el("button", { class: "dugme", disabled: true }, m.bos_yer > 0 ? "süre doldu" : "yer yok");
    }
    const dugme = el("button", { class: "dugme", "data-menu-id": m.menu_id, onclick: rezervasyonYap }, "rezervasyon yap");
    // Aşama 4: yarım porsiyon yalnızca API yarım fiyat verdiyse (öğle/akşam) seçilebilir
    if (m.yarim_fiyat == null) return dugme;
    return el("div", { class: "rezervasyon-secim" },
        el("label", { class: "gorunmez", for: "porsiyon-" + m.menu_id }, "Porsiyon"),
        el("select", { id: "porsiyon-" + m.menu_id, name: "porsiyon" },
            el("option", { value: "tam" }, "tam · " + tl(m.fiyat)),
            el("option", { value: "yarim" }, "yarım · " + tl(m.yarim_fiyat)),
        ),
        dugme,
    );
}

function satir(m) {
    return el("tr", {},
        el("td", { "data-baslik": "Gün" }, gunMetni(m.tarih)),
        el("td", { "data-baslik": "Öğün" }, m.ogun),
        el("td", { "data-baslik": "Yemekler", class: "yemekler" }, m.yemekler.join(", "),
            m.alerjen_uyarisi && el("span", { class: "rozet rozet-uyari", title: "Profilindeki bir alerjeni içeriyor" }, "⚠ alerjen")),
        el("td", { "data-baslik": "Kcal" }, m.toplam_kalori),
        el("td", { "data-baslik": "Fiyat", class: "fiyat" }, tl(m.fiyat)),
        el("td", { "data-baslik": "Boş yer" }, m.bos_yer),
        el("td", {}, islemHucresi(m)),
    );
}

async function haftayiGetir() {
    aralik.textContent = gunMetni(hafta) + " – " + gunMetni(gunEkle(hafta, 6));
    try {
        const liste = await istek(`/menuler/haftalik?yemekhane_id=${secim.yemekhane_id.value}&baslangic=${hafta}`);
        tbody.replaceChildren(...liste.map(satir));
        haftaBos.hidden = liste.length > 0;
    } catch (hata) {
        tbody.replaceChildren();
        mesaj(mesajKutusu, hata.message);
    }
}

async function rezervasyonYap(olay) {
    const dugme = olay.currentTarget;
    dugme.disabled = true;
    mesaj(mesajKutusu, null);
    try {
        // Yanıttaki bakiye kaynak veritabanından gelir (replikasyon gecikmesi yok)
        const secim = dugme.closest("td").querySelector('select[name="porsiyon"]');
        const govde = { menu_id: Number(dugme.dataset.menuId), porsiyon: secim?.value ?? "tam" };
        const sonuc = await istek("/rezervasyonlar", { method: "POST", govde });
        const porsiyon = sonuc.porsiyon === "yarim" ? "yarım porsiyon, " : "";
        mesaj(mesajKutusu, `Rezervasyon yapıldı (${porsiyon}${tl(sonuc.ucret)}). Kalan bakiyeniz: ${tl(sonuc.bakiye)}`, "basari");
        // Tabloyu hemen yeniden istemiyoruz: liste replikadan okunur ve yazma
        // oraya birkaç milisaniye geç ulaşır (replikasyon gecikmesi), eski hali
        // görünürdü. Satırı kaynaktan gelen bu başarılı yanıtla güncelliyoruz.
        const bosYer = dugme.closest("tr").querySelector('[data-baslik="Boş yer"]');
        bosYer.textContent = Number(bosYer.textContent) - 1;
        (dugme.closest(".rezervasyon-secim") ?? dugme).replaceWith(el("span", { class: "rozet rozet-aktif" },
            sonuc.porsiyon === "yarim" ? "rezervasyonun var (yarım)" : "rezervasyonun var"));
    } catch (hata) {
        // "Bakiye yetersiz." gibi procedure mesajları olduğu gibi gösterilir
        mesaj(mesajKutusu, hata.message);
        dugme.disabled = false;
    }
}

document.getElementById("onceki-hafta").addEventListener("click", () => { hafta = gunEkle(hafta, -7); haftayiGetir(); });
document.getElementById("sonraki-hafta").addEventListener("click", () => { hafta = gunEkle(hafta, 7); haftayiGetir(); });
secim.addEventListener("change", haftayiGetir);

istek("/yemekhaneler")
    .then((yemekhaneler) => {
        secenekleriDoldur(secim.yemekhane_id, yemekhaneler, yemekhaneler[0]?.id);
        return haftayiGetir();
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
