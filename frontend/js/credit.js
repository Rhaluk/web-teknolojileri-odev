// Bakiye: GET /api/bakiye, POST /api/bakiye/yukle
import { istek, el, mesaj, tl, saatMetni, ben } from "./api.js";

const bakiye = document.getElementById("bakiye");
const tbody = document.querySelector("#hareketler tbody");
const form = document.getElementById("yukleme-formu");
const mesajKutusu = document.getElementById("yukleme-mesaj");

function satir(h) {
    const arti = h.tutar >= 0;
    return el("tr", {},
        el("td", { "data-baslik": "tarih" }, saatMetni(h.olusturma)),
        el("td", { "data-baslik": "açıklama" }, h.aciklama),
        el("td", { "data-baslik": "tutar", class: arti ? "tutar-arti" : "tutar-eksi" },
            (arti ? "+" : "−") + tl(Math.abs(h.tutar))),
    );
}

// Hızlı seçim tutar kutusunu doldurur; kutuya elle yazınca seçim kalkar
form.addEventListener("change", (olay) => {
    if (olay.target.name === "hizli") form.tutar.value = olay.target.value;
});
form.tutar.addEventListener("input", () => {
    form.querySelectorAll('input[name="hizli"]').forEach((r) => { r.checked = r.value === form.tutar.value; });
});

form.addEventListener("submit", async (olay) => {
    olay.preventDefault();
    const dugme = form.querySelector('button[type="submit"]');
    dugme.disabled = true;
    mesaj(mesajKutusu, null);
    const tutar = Number(form.tutar.value);
    try {
        // 10–1000 TL sınırını sp_bakiye_yukle denetler; burada ayrıca kontrol yok
        const sonuc = await istek("/bakiye/yukle", { method: "POST", govde: { tutar } });
        // Bakiye kaynaktan döndü; listeyi replikadan yeniden okumak yerine
        // yeni hareketi başa ekliyoruz (replikasyon gecikmesi).
        bakiye.textContent = tl(sonuc.bakiye);
        tbody.prepend(satir({ olusturma: new Date(), aciklama: "Bakiye yükleme", tutar }));
        mesaj(mesajKutusu, `${tl(tutar)} yüklendi. Yeni bakiyeniz: ${tl(sonuc.bakiye)}`, "basari");
    } catch (hata) {
        mesaj(mesajKutusu, hata.message);
    } finally {
        dugme.disabled = false;
    }
});

ben().then((k) => { document.getElementById("selam").textContent = "merhaba " + k.ad + " " + k.soyad; }).catch(() => {});

istek("/bakiye")
    .then((b) => {
        bakiye.textContent = tl(b.bakiye);
        tbody.replaceChildren(...b.hareketler.map(satir));
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
