// Yönetici paneli: menü oluşturma/düzenleme ve menü raporu
// GET /api/yemekhaneler, /api/yonetici/yemekler, /api/yonetici/menuler(/{id})
// POST /api/yonetici/menuler, PUT /api/yonetici/menuler/{id}
import { istek, el, mesaj, gunMetni, bugun, secenekleriDoldur } from "./api.js";

const form = document.getElementById("menu-formu");
const mesajKutusu = document.getElementById("menu-mesaj");
const duzenlenen = document.getElementById("duzenlenen");
const filtre = document.getElementById("rapor-filtre");
const tbody = document.querySelector("#rapor-tablo tbody");

const sayiBicimi = new Intl.NumberFormat("tr-TR", { minimumFractionDigits: 1, maximumFractionDigits: 2 });

function gunEkle(tarih, gun) {
    const [y, a, g] = tarih.split("-").map(Number);
    const t = new Date(y, a - 1, g + gun);
    return [t.getFullYear(), String(t.getMonth() + 1).padStart(2, "0"), String(t.getDate()).padStart(2, "0")].join("-");
}

// ---------- Menü formu ----------

// Yemekleri kategoriye göre gruplayıp onay kutusu çipleri üretir
function yemekGruplari(yemekler) {
    const gruplar = new Map();
    for (const y of yemekler) {
        if (!gruplar.has(y.kategori)) gruplar.set(y.kategori, []);
        gruplar.get(y.kategori).push(y);
    }
    return [...gruplar].map(([kategori, liste]) =>
        el("fieldset", { class: "kategori-grubu" },
            el("legend", {}, kategori),
            el("div", { class: "secim-kutulari" },
                liste.map((y) => el("label", {},
                    el("input", { type: "checkbox", name: "yemek_id", value: y.id }),
                    " " + y.ad + " ",
                    el("span", { class: "soluk" }, y.kalori),
                )),
            ),
        ));
}

// "temizle" ya da başarılı kayıt sonrası form yeni menü moduna döner
function yeniMenuModu() {
    form.menu_id.value = "";
    duzenlenen.textContent = "yeni";
}
form.addEventListener("reset", () => setTimeout(yeniMenuModu));

async function duzenle(menuId) {
    mesaj(mesajKutusu, null);
    try {
        const m = await istek("/yonetici/menuler/" + menuId);
        form.menu_id.value = m.id;
        form.yemekhane_id.value = m.yemekhane_id;
        form.tarih.value = m.tarih;
        form.ogun_id.value = m.ogun_id;
        form.ogrenci_fiyati.value = m.ogrenci_fiyati;
        form.personel_fiyati.value = m.personel_fiyati;
        const secili = new Set(m.yemek_idler.map(String));
        form.querySelectorAll('input[name="yemek_id"]').forEach((k) => { k.checked = secili.has(k.value); });
        duzenlenen.textContent = "#" + m.id;
        form.scrollIntoView({ behavior: "smooth" });
    } catch (hata) {
        mesaj(mesajKutusu, hata.message);
    }
}

form.addEventListener("submit", async (olay) => {
    olay.preventDefault();
    const dugme = form.querySelector('button[type="submit"]');
    dugme.disabled = true;
    mesaj(mesajKutusu, null);

    const id = form.menu_id.value;
    const govde = {
        yemekhane_id: Number(form.yemekhane_id.value),
        tarih: form.tarih.value,
        ogun_id: Number(form.ogun_id.value),
        ogrenci_fiyati: Number(form.ogrenci_fiyati.value),
        personel_fiyati: Number(form.personel_fiyati.value),
        yemek_idler: [...form.querySelectorAll('input[name="yemek_id"]:checked')].map((k) => Number(k.value)),
    };
    try {
        const m = id
            ? await istek("/yonetici/menuler/" + id, { method: "PUT", govde })
            : await istek("/yonetici/menuler", { method: "POST", govde });
        mesaj(mesajKutusu, `Menü #${m.id} ${id ? "güncellendi" : "oluşturuldu"}.`, "basari");
        form.reset();
        // Rapor replikadan okunur; yeni yazma oraya birkaç milisaniye geç
        // ulaşabileceği için listeyi kısa bir gecikmeyle yeniliyoruz.
        setTimeout(raporuGetir, 500);
    } catch (hata) {
        // "Bu yemekhane, tarih ve öğün için zaten bir menü var." gibi
        mesaj(mesajKutusu, hata.message);
    } finally {
        dugme.disabled = false;
    }
});

// ---------- Rapor ----------

function dolulukHucresi(r) {
    return el("span", { class: "doluluk" },
        el("span", { class: "doluluk-cubuk", style: `--oran: ${r.doluluk_yuzde}%` }),
        el("span", {}, `${r.rezerve} / ${r.kapasite} `, el("span", { class: "soluk" }, `(%${sayiBicimi.format(r.doluluk_yuzde)})`)),
    );
}

function puanHucresi(r) {
    if (r.ortalama_puan == null) return el("span", { class: "soluk" }, "henüz oy yok");
    const p = sayiBicimi.format(r.ortalama_puan);
    return el("span", { class: "puan" },
        el("span", { class: "stars-div kucuk", style: `--puan: ${r.ortalama_puan}`, role: "img", "aria-label": `5 üzerinden ${p}` }),
        ` ${p} `,
        el("span", { class: "soluk" }, `(${r.oy_sayisi} oy)`),
    );
}

function satir(r) {
    return el("tr", {},
        el("td", { "data-baslik": "tarih" }, gunMetni(r.tarih)),
        el("td", { "data-baslik": "yemekhane" }, r.yemekhane),
        el("td", { "data-baslik": "öğün" }, r.ogun),
        el("td", { "data-baslik": "doluluk" }, dolulukHucresi(r)),
        el("td", { "data-baslik": "ortalama puan" }, puanHucresi(r)),
        el("td", {}, el("button", { type: "button", class: "dugme dugme-ikincil", onclick: () => duzenle(r.menu_id) }, "düzenle")),
    );
}

async function raporuGetir() {
    try {
        const liste = await istek(`/yonetici/menuler?baslangic=${filtre.baslangic.value}&bitis=${filtre.bitis.value}`);
        tbody.replaceChildren(...liste.map(satir));
    } catch (hata) {
        mesaj(mesajKutusu, hata.message);
    }
}
filtre.addEventListener("change", raporuGetir);

// ---------- Açılış ----------

filtre.baslangic.value = gunEkle(bugun(), -7);
filtre.bitis.value = gunEkle(bugun(), 7);

Promise.all([istek("/yemekhaneler"), istek("/yonetici/yemekler")])
    .then(([yemekhaneler, yemekler]) => {
        secenekleriDoldur(form.yemekhane_id, yemekhaneler, yemekhaneler[0]?.id);
        document.getElementById("yemek-gruplari").replaceChildren(...yemekGruplari(yemekler));
        return raporuGetir();
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
