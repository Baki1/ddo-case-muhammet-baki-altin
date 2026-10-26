-- Demo kullanıcılar (TESTING.md ile birebir uyumlu)
INSERT INTO "yn-ej-users" (email, password_hash, role) VALUES
('ogrenci@ddo-case.invalid', '123456', 'ogrenci'),
('akademisyen@ddo-case.invalid', '123456', 'akademisyen'),
('personel@ddo-case.invalid', '123456', 'kutuphaneci'),
('satinalma@ddo-case.invalid', '123456', 'satinalma')
ON CONFLICT (email) DO NOTHING;

-- Örnek Kitaplar (q=algoritma aramasını ve Türkçe FTS'yi kanıtlamak için)
INSERT INTO "yn-books" (title, author, isbn, subject) VALUES
('İleri Algoritma Analizi ve Tasarımı', 'Prof. Dr. Kemal İnan', '9789750802942', 'Bilgisayar Mühendisliği'),
('Veri Yapıları ve Algoritmalar', 'Doç. Dr. Selin Şahin', '9789753638029', 'Bilgisayar Mühendisliği'),
('Saatleri Ayarlama Enstitüsü', 'Ahmet Hamdi Tanpınar', '9789754700114', 'Türk Edebiyatı'),
('Kürk Mantolu Madonna', 'Sabahattin Ali', '9780134190440', 'Türk Edebiyatı'),
('Dağıtık Sistemler ve Veritabanı Mimarisi', 'Martin Kleppmann', '9781492056355', 'Yazılım Mühendisliği')
ON CONFLICT (isbn) DO NOTHING;

-- Örnek Fiziksel Nüshalar (Barkodlar)
INSERT INTO "yn-copies" (book_id, barcode, status) VALUES
(1, 'BC-1001', 'available'),
(1, 'BC-1002', 'available'),
(2, 'BC-1003', 'available'),
(3, 'BC-1004', 'available'),
(4, 'BC-1005', 'available')
ON CONFLICT (barcode) DO NOTHING;