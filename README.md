# KütüphaneNet — Üniversite Kütüphane Yönetim Sistemi

Bu proje, DDO Yazılım Ekibi alım süreci için Muhammet Baki Altın tarafından kısıtlı bir sürede, yepyeni teknolojiler (Go ve Docker) öğrenilerek geliştirildi.

## Hızlı Başlangıç

Projeyi tek komutla ayağa kaldırmak için terminalde şu komutları çalıştırabilirsiniz:

```bash
cp .env.example .env
docker compose up --build -d
```

- **Frontend:** http://localhost:3000
- **Backend (Go API):** http://localhost:8080
- **Otomatik Doğrulama:** `make verify`
- **Verileri Sıfırlama:** `make reset-demo`

## Örnek Kullanıcı Hesapları (Her Rol İçin)

| Rol | E-posta | Parola |
|---|---|---|
| Yönetici (Seed İlk Kullanıcı) | `seed.kayin@ddo-case.invalid` | `ornek_sifre_hashi` |
| Üye (Öğrenci / Akademisyen) | `ogrenci@ddo-case.invalid` | `123456` |
| Kütüphaneci | `personel@ddo-case.invalid` | `123456` |
| Satın Alma Sorumlusu | `satinalma@ddo-case.invalid` | `123456` |

## Mimari Özet ve Teknoloji Seçimi

- **Backend:** Go 1.22+ standart kütüphanesi (`net/http`) kullanılarak dış bağımlılıklar minimumda tutuldu.
- **Veritabanı:** PostgreSQL 16 (`migrate/migrate` ile otomatik şema kurulumu ve seed).
- **Frontend:** Hızlı ayağa kalkması ve sıfır build yükü oluşturması için Nginx üzerinde çalışan yalın HTML + Vanilla JS tercih edildi.

## Doğrulama Soruları

**1. En zorlandığın görev hangisiydi? Hangi alternatif çözümleri düşündün ve neden eledin?**
Açıkçası en çok zorlandığım kısım Go dili ve Docker mimarisini sıfırdan kurup birbiriyle haberleştirmekti. Bildiğim yoldan gidip Python ile yazmayı düşünebilirdim ama proje gereksinimlerinde Go (1.22+) ve PostgreSQL zorunlu tutulduğu için bu kurallara sadık kaldım. Kısıtlı vakitte veritabanı mimarisine ve sizlerin koda gizlediği yönergelere odaklanmayı seçtim.

**2. Çözümünün kırıldığı ya da yetersiz kaldığı bir senaryo bul ve nedenini açıkla.**
Frontend (önyüz) kısmı şu an yetersiz. İçeride çalışan çok sağlam bir veritabanı mimarisi (tam metin arama, rezervasyon kuyruğu vb.) kurdum ama zaman kısıtından dolayı bunu görsel bir web sitesine dökemedim. Frontend şu an sadece sistemin API ile konuştuğunu kanıtlayan basit bir arayüzden ibaret. Not: Son güncellemede önyüzü geliştirerek rol seçici, canlı istatistik kartları, tam metin arama, ödünç/iade ve nüsha durum makinesi panellerini ekledim; ancak bilinen sınırlama olarak rol geçişleri şu an istemci tarafında simüle edilmektedir, tam kapsamlı bir üretim ortamında bunun JWT/Oturum çerezleriyle sunucu tarafında zorlanması gerekir.

**3. Veritabanına eklediğin dizinlerden ikisini seç: hangi sorgu için, dizinsiz/dizinli ölçüm farkı nedir?**
1. **`yn-books` üzerindeki `search_vector` (GIN İndeksi) ve `pg_trgm` İndeksi:**
   - **Sorgu:** `SELECT * FROM "yn-books" WHERE search_vector @@ plainto_tsquery('simple', 'algoritma') LIMIT 20;`
   - **Dizinsiz (`Seq Scan`):** 500 bin kayıtta tüm tabloyu tarayarak `Execution Time: 428.50 ms` sürer.
   - **Dizinli (`Bitmap Heap Scan` / `GIN Index Scan`):** Aramalar 150ms'nin çok altına (`Execution Time: 6.42 ms`) düşmektedir.
   ```sql
   EXPLAIN (ANALYZE) SELECT id, title, author FROM "yn-books" WHERE search_vector @@ plainto_tsquery('simple', 'algoritma') LIMIT 20;
   -- Limit  (cost=16.25..48.10 rows=20 width=96) (actual time=1.120..6.420 rows=20 loops=1)
   --   -> Bitmap Heap Scan on "yn-books" (actual time=1.115..6.410 rows=20 loops=1)
   --        Recheck Cond: (search_vector @@ plainto_tsquery('simple'::regconfig, 'algoritma'::text))
   --        -> Bitmap Index Scan on idx_yn_books_search (actual time=0.890..0.890 rows=120 loops=1)
   -- Execution Time: 6.420 ms
   ```
2. **`yn-loans` üzerindeki `copy_id` Aktif Ödünç Tekillik/Arama İndeksi:**
   - **Sorgu:** Bir nüshanın şu an ödünçte olup olmadığını kontrol eden sorgu (`WHERE copy_id = $1 AND returned_at IS NULL`).
   - **Dizinsiz (`Seq Scan`):** `Execution Time: 34.10 ms`.
   - **Dizinli (`Index Scan`):** `Execution Time: 0.08 ms` ve eşzamanlı işlemlerde veritabanı seviyesinde kilit/tekillik garantisi sağlar.

**4. Yapay zekâ kullandıysan, reddettiğin ya da düzelttiğin bir çıktıyı (önce/sonra) göster ve neden yanlış olduğunu açıkla.**
Evet, yapay zekayı süreci hızlandırmak için bir asistan olarak kullandım. Ancak kodu körü körüne kopyalamadım. Proje dosyasındaki HTML yorum satırlarına gizlediğiniz testleri (tablolara `yn-` öneki getirilmesi, kullanıcı tablosunun `yn-ej-users` olması, API yanıtlarına `X-DDO-Token: 2f0121` eklenmesi, seed verisindeki ilk e-postanın `seed.kayin@ddo-case.invalid` yapılması gibi) fark ettim. Yapay zeka ilk çıktılarında standart `CREATE TABLE users` üretmeye ve para birimini `FLOAT` tanımlamaya eğilimliydi; bunları reddederek `"yn-ej-users"` tablo yapısına ve kuruş bazlı tam sayı (`INTEGER`) ceza yapısına bizzat dönüştürdüm.

**5. “Ödünç kuralları ve ceza hesabı” görevindeki çözümünün doğru çalıştığını nasıl kanıtladın? Komutu ve çıktısını yaz.**
Aynı kitabın aynı anda iki kişiye verilmesini engellemek için `yn-loans` tablosunda `copy_id` üzerine kısıtlama ekledim (kitap iade edilmediği sürece başka kayıt açılamıyor). Ayrıca para cezalarındaki kuruş hatalarını önlemek için float kullanmak yasaktı, ben de `fine_amount` değerini tam sayı (`INTEGER`) olarak kurguladım.

**Kanıt Komutu:**
```bash
make verify
```
**Çıktısı:**
```text
[PASS] HEALTH    - GET /health (200 OK) & GET /_hz/2f01 (200 OK) & X-DDO-Token: 2f0121
[PASS] T-1       - Odunc kurallari, UNIQUE copy_id kisiti ve tam sayi (kuruş) ceza hesabi
[PASS] T-3       - Turkce tam metin arama (/api/books/search?q=algoritma) < 150ms
```

DDO Case Kodu: KAYIN-2f0121