// Profil: GET /api/alerjenler, GET/PUT /api/profil/alerjenler
import { istek, el, mesaj, ben } from "./api.js";

const form = document.getElementById("alerjen-formu");
const liste = document.getElementById("alerjen-listesi");
const mesajKutusu = document.getElementById("profil-mesaj");

form.addEventListener("submit", async (olay) => {
    olay.preventDefault();
    const dugme = form.querySelector("button");
    dugme.disabled = true;
    mesaj(mesajKutusu, null);
    const alerjen_idler = [...form.querySelectorAll('input[name="alerjen_id"]:checked')].map((k) => Number(k.value));
    try {
        await istek("/profil/alerjenler", { method: "PUT", govde: { alerjen_idler } });
        mesaj(mesajKutusu, alerjen_idler.length
            ? `${alerjen_idler.length} alerjen kaydedildi. Menüde bu alerjenleri içeren yemekler işaretlenecek.`
            : "Alerjen seçimin kaldırıldı.", "basari");
    } catch (hata) {
        mesaj(mesajKutusu, hata.message);
    } finally {
        dugme.disabled = false;
    }
});

ben().then((k) => {
    document.getElementById("profil-ad").textContent = k.ad + " " + k.soyad;
    document.getElementById("profil-bilgi").textContent = k.eposta;
}).catch(() => {});

Promise.all([istek("/alerjenler"), istek("/profil/alerjenler")])
    .then(([alerjenler, profil]) => {
        const secili = new Set(profil.alerjen_idler);
        liste.replaceChildren(...alerjenler.map((a) => el("label", {},
            el("input", { type: "checkbox", name: "alerjen_id", value: a.id, checked: secili.has(a.id) }),
            " " + a.ad,
        )));
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
