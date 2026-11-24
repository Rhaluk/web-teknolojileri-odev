// Günün menüsü: GET /api/yemekhaneler, GET /api/menu
import { istek, el, mesaj, tl, gunMetni, bugun, secenekleriDoldur } from "./api.js";

const form = document.getElementById("menu-secim");
const baslik = document.getElementById("sayfa-baslik");
const hataKutusu = document.getElementById("menu-hata");
const bolum = document.getElementById("gunun-menusu");
const liste = document.getElementById("yemek-listesi");
const menuYok = document.getElementById("menu-yok");

// Bir yemeğin satırı: kategori, ad, kalori, vejetaryen ve alerjen etiketleri
function yemekSatiri(y) {
    return el("li", {},
        el("span", { class: "kategori" }, y.kategori),
        el("span", { class: "yemek" }, y.ad),
        el("span", { class: "kalori" }, y.kalori + " kcal"),
        el("span", { class: "etiketler" },
            y.vejetaryen && el("span", { class: "rozet rozet-vejetaryen" }, "vejetaryen"),
            y.alerjenler.map((a) => el("span", { class: "rozet" }, a)),
        ),
    );
}

async function menuyuGetir() {
    mesaj(hataKutusu, null);
    const yemekhane = form.yemekhane_id.value;
    const tarih = form.tarih.value || bugun();
    const ogun = form.ogun_id.value;
    baslik.textContent = "Günün menüsü · " + gunMetni(tarih);

    try {
        const menu = await istek(`/menu?yemekhane_id=${yemekhane}&ogun_id=${ogun}&tarih=${tarih}`);
        liste.replaceChildren(...menu.yemekler.map(yemekSatiri));
        document.getElementById("toplam-kalori").textContent = menu.toplam_kalori + " kcal";
        document.getElementById("fiyat").textContent = tl(menu.fiyat);
        document.getElementById("bos-yer").textContent = menu.bos_yer;
        bolum.hidden = false;
        menuYok.hidden = true;
    } catch (hata) {
        bolum.hidden = true;
        if (hata.durum === 404) {
            menuYok.hidden = false;
        } else {
            menuYok.hidden = true;
            mesaj(hataKutusu, hata.message);
        }
    }
}

// Seçimlerden biri değişince menü yeniden istenir
form.addEventListener("change", menuyuGetir);
form.addEventListener("submit", (olay) => olay.preventDefault());

// Sayfa açılışı: yemekhaneleri doldur, bugünün menüsünü getir
// index.html?tarih=2026-11-18 ile başka bir gün açılabilir
form.tarih.value = new URLSearchParams(location.search).get("tarih") ?? bugun();
istek("/yemekhaneler")
    .then((yemekhaneler) => {
        secenekleriDoldur(form.yemekhane_id, yemekhaneler, yemekhaneler[0]?.id);
        return menuyuGetir();
    })
    .catch((hata) => mesaj(hataKutusu, hata.message));
