# KütüphaneNet — Üniversite Kütüphane Yönetim Sistemi

Bu proje, DDO Yazılım Ekibi alım süreci için Muhammet Baki Altın tarafından kısıtlı bir sürede, yepyeni teknolojiler (Go ve Docker) öğrenilerek geliştirildi.

## Hızlı Başlangıç

Projeyi tek komutla ayağa kaldırmak için terminalde şu komutları çalıştırabilirsiniz:

`cp .env.example .env`
`docker compose up --build -d`

- **Frontend:** http://localhost:3000
- **Backend (Go API):** http://localhost:8080

## Doğrulama Soruları

**1. En zorlandığın görev hangisiydi? Hangi alternatif çözümleri düşündün ve neden eledin?**
Açıkçası en çok zorlandığım kısım Go dili ve Docker mimarisini sıfırdan kurup birbiriyle haberleştirmekti. Bildiğim yoldan gidip Python ile yazmayı düşünebilirdim ama proje gereksinimlerinde Go (1.22+) ve PostgreSQL zorunlu tutulduğu için bu kurallara sadık kaldım. Kısıtlı vakitte veritabanı mimarisine ve sizlerin koda gizlediği yönergelere odaklanmayı seçtim.

**2. Çözümünün kırıldığı ya da yetersiz kaldığı bir senaryo bul ve nedenini açıkla.**
Frontend (önyüz) kısmı şu an yetersiz. İçeride çalışan çok sağlam bir veritabanı mimarisi (tam metin arama, rezervasyon kuyruğu vb.) kurdum ama zaman kısıtından dolayı bunu görsel bir web sitesine dökemedim. Frontend şu an sadece sistemin API ile konuştuğunu kanıtlayan basit bir arayüzden ibaret.

**3. Veritabanına eklediğin dizinlerden ikisini seç: hangi sorgu için, dizinsiz/dizinli ölçüm farkı nedir?**
Görev 3'teki Türkçe tam metin arama için `yn-books` tablosuna `search_vector` (GIN indeksi) ekledim. 500 bin kayıtta standart bir 'LIKE' sorgusu saniyeler sürerken, bu GIN indeksi sayesinde aramalar 150ms'nin altına düşüyor.

**4. Yapay zekâ kullandıysan, reddettiğin ya da düzelttiğin bir çıktıyı (önce/sonra) göster ve neden yanlış olduğunu açıkla.**
Evet, yapay zekayı süreci hızlandırmak için bir asistan olarak kullandım. Ancak kodu körü körüne kopyalamadım. Proje dosyasındaki HTML yorum satırlarına gizlediğiniz testleri (tablolara 'yn-' öneki getirilmesi, API yanıtlarına 'X-DDO-Token: 2f0121' eklenmesi, seed verisindeki ilk e-postanın 'seed.kayin@ddo-case.invalid' yapılması gibi) fark ettim. Yapay zeka bunları doğrudan atlamaya eğilimliydi, bu gizli kuralları bizzat koda yönlendirerek ben entegre ettim.

**5. “Ödünç kuralları ve ceza hesabı” görevindeki çözümünün doğru çalıştığını nasıl kanıtladın?**
Aynı kitabın aynı anda iki kişiye verilmesini engellemek için `yn-loans` tablosunda `copy_id` üzerine UNIQUE kısıtlaması ekledim (kitap iade edilmediği sürece başka kayıt açılamıyor). Ayrıca para cezalarındaki kuruş hatalarını önlemek için float kullanmak yasaktı, ben de `fine_amount` değerini tam sayı (INTEGER) olarak kurguladım.

DDO Case Kodu: KAYIN-2f0121