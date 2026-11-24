// Rezervasyonlarım: GET /api/rezervasyonlar, DELETE /api/rezervasyonlar/{id}
import { istek, el, mesaj, tl, gunMetni } from "./api.js";

const aktifListe = document.getElementById("aktif-liste");
const gecmisListe = document.getElementById("gecmis-liste");
const mesajKutusu = document.getElementById("iptal-mesaj");

const rozetler = {
    aktif: el("span", { class: "rozet rozet-aktif" }, "aktif"),
    kullanildi: el("span", { class: "rozet rozet-kullanildi" }, "kullanıldı"),
    iptal: el("span", { class: "rozet rozet-iptal" }, "iptal"),
};

// Satırın altındaki işlem. İptal edilebilir mi kararını API verir.
function islem(r) {
    if (r.durum === "aktif") {
        return r.iptal_edilebilir
            ? el("button", { "data-rezervasyon-id": r.id, onclick: iptalEt }, "rezervasyon iptal")
            : el("button", { disabled: true }, "iptal süresi doldu");
    }
    if (r.durum === "kullanildi") {
        return r.degerlendirildi
            ? el("span", { class: "soluk" }, "değerlendirildi")
            : el("a", { class: "dugme", href: "rate.html" }, "puanla");
    }
    return el("span", { class: "soluk" }, "ücret iade edildi");
}

function satir(r) {
    return el("li", {},
        el("span", { class: "tarih" }, gunMetni(r.tarih) + " · " + r.ogun + (r.porsiyon === "yarim" ? " · yarım porsiyon" : "")),
        el("span", { class: "yer" }, r.yemekhane),
        el("span", { class: "fiyat" }, tl(r.ucret)),
        rozetler[r.durum].cloneNode(true),
        islem(r),
    );
}

function bosMu() {
    document.getElementById("aktif-bos").hidden = aktifListe.children.length > 0;
    document.getElementById("gecmis-bos").hidden = gecmisListe.children.length > 0;
}

async function iptalEt(olay) {
    const dugme = olay.currentTarget;
    if (!confirm("Bu rezervasyonu iptal etmek istiyor musunuz?")) return;
    dugme.disabled = true;
    mesaj(mesajKutusu, null);
    try {
        const sonuc = await istek("/rezervasyonlar/" + dugme.dataset.rezervasyonId, { method: "DELETE" });
        mesaj(mesajKutusu, `Rezervasyon iptal edildi, ${tl(sonuc.iade)} iade edildi. Bakiyeniz: ${tl(sonuc.bakiye)}`, "basari");
        // Listeyi replikadan hemen yeniden okumak yerine (replikasyon gecikmesi)
        // satırı iptal edilmiş olarak geçmiş listesinin başına taşıyoruz.
        const li = dugme.closest("li");
        li.querySelector(".rozet").replaceWith(rozetler.iptal.cloneNode(true));
        dugme.replaceWith(el("span", { class: "soluk" }, "ücret iade edildi"));
        gecmisListe.prepend(li);
        bosMu();
    } catch (hata) {
        // "Bu rezervasyon için iptal süresi doldu." gibi mesajlar olduğu gibi
        mesaj(mesajKutusu, hata.message);
        dugme.disabled = false;
    }
}

istek("/rezervasyonlar")
    .then((liste) => {
        aktifListe.replaceChildren(...liste.filter((r) => r.durum === "aktif").map(satir));
        gecmisListe.replaceChildren(...liste.filter((r) => r.durum !== "aktif").map(satir));
        bosMu();
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
