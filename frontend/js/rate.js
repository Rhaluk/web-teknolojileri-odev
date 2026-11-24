// Değerlendirme: GET /api/degerlendirmeler, POST /api/degerlendirmeler
import { istek, el, mesaj, gunMetni } from "./api.js";

const tbody = document.querySelector("#rate-table tbody");
const mesajKutusu = document.getElementById("puan-mesaj");

// Salt okunur yıldızlar: --puan değişkeni rate.css'teki maskeyi doldurur
const yildizlar = (puan) => el("div", { class: "stars-div", style: "--puan: " + puan, role: "img", "aria-label": `5 üzerinden ${puan} yıldız` });

// Puanlanabilir yıldızlar: 5 radio + label, altında maskeli div (rate.css)
function yildizSecimi(formId, menuId) {
    const kutu = el("fieldset", { class: "yildiz-kutu", form: formId }, el("legend", { class: "gorunmez" }, "Puan"));
    for (let n = 1; n <= 5; n++) {
        const id = `p${menuId}-${n}`;
        kutu.append(
            el("input", { type: "radio", id, name: "puan", value: n, form: formId, required: n === 1 }),
            el("label", { for: id }, el("span", { class: "gorunmez" }, n + " yıldız")),
        );
    }
    kutu.append(el("div", { class: "stars-div", "aria-hidden": "true" }));
    return kutu;
}

function satir(d) {
    const ust = [
        el("td", { "data-baslik": "tarih" }, gunMetni(d.tarih) + " · " + d.ogun),
        el("td", { "data-baslik": "menu" }, d.yemekhane),
    ];
    // Daha önce puanlanmış: yıldızlar yalnızca gösterilir
    if (d.puan != null) {
        return el("tr", {}, ...ust,
            el("td", { "data-baslik": "Yorum" }, el("span", { class: "soluk" }, d.yorum ?? "—")),
            el("td", { "data-baslik": "Puan" }, yildizlar(d.puan)),
            el("td", {}, el("span", { class: "rozet rozet-kullanildi" }, "puanlandı")),
        );
    }
    const formId = "form-" + d.menu_id;
    return el("tr", {}, ...ust,
        el("td", { "data-baslik": "Yorum" },
            el("label", { for: "yorum-" + d.menu_id, class: "gorunmez" }, "Yorum"),
            el("input", { type: "text", id: "yorum-" + d.menu_id, name: "yorum", form: formId, maxlength: 500, placeholder: "isteğe bağlı" }),
        ),
        el("td", { "data-baslik": "Puan" }, yildizSecimi(formId, d.menu_id)),
        el("td", {},
            el("form", { id: formId, "data-menu-id": d.menu_id, onsubmit: (olay) => gonder(olay, d) },
                el("button", { type: "submit", class: "dugme" }, "gönder"),
            ),
        ),
    );
}

// d: satırın API'den gelen verisi; başarıdan sonra satır puanlanmış haliyle yeniden çizilir
async function gonder(olay, d) {
    olay.preventDefault();
    const form = olay.currentTarget;
    const dugme = form.querySelector("button");
    dugme.disabled = true;
    mesaj(mesajKutusu, null);
    // form="" ile bağlı alanlar form.elements içinde gelir
    const yorum = form.elements.yorum.value.trim();
    const govde = { menu_id: d.menu_id, puan: Number(form.elements.puan.value), yorum };
    try {
        await istek("/degerlendirmeler", { method: "POST", govde });
        mesaj(mesajKutusu, "Değerlendirmeniz kaydedildi, teşekkürler.", "basari");
        // Replikadan yeniden okumak yerine satırı puanlanmış haline çeviriyoruz
        form.closest("tr").replaceWith(satir({ ...d, puan: govde.puan, yorum: yorum || null }));
    } catch (hata) {
        mesaj(mesajKutusu, hata.message);
        dugme.disabled = false;
    }
}

istek("/degerlendirmeler")
    .then((liste) => {
        tbody.replaceChildren(...liste.map(satir));
        document.getElementById("puan-bos").hidden = liste.length > 0;
    })
    .catch((hata) => mesaj(mesajKutusu, hata.message));
