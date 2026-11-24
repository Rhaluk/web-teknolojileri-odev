// Giriş sayfası: e-posta + şifre → POST /api/giris
// Başarılıysa oturum çerezi sunucu tarafından yazılır (HttpOnly).
import { istek, mesaj } from "./api.js";

const form = document.getElementById("giris-formu");
const hataKutusu = document.getElementById("giris-hata");

form.addEventListener("submit", async (olay) => {
    olay.preventDefault(); // sayfa yenilenmesin
    mesaj(hataKutusu, null);

    const dugme = form.querySelector("button");
    dugme.disabled = true;
    try {
        const kullanici = await istek("/giris", {
            method: "POST",
            govde: { eposta: form.eposta.value.trim(), sifre: form.sifre.value },
        });
        location.href = kullanici.rol === "yonetici" ? "adminPanel.html" : "index.html";
    } catch (hata) {
        mesaj(hataKutusu, hata.message);
        form.sifre.value = "";
        form.sifre.focus();
    } finally {
        dugme.disabled = false;
    }
});
