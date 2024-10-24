# URL Kısaltma Servisi

Golang, Fiber, GORM, Redis ve PostgreSQL kullanarak geliştirilmiş basit ve verimli bir URL kısaltma servisidir. Bu servis, kullanıcıların uzun URL'leri kısaltmalarını, kısaltılmış URL'ler hakkında istatistikler görmelerini ve ihtiyaç duyulmadığında URL'leri silmelerini sağlar.

## Özellikler

- **URL Kısaltma**: Kullanıcılar, uzun URL'leri göndererek kısaltılmış versiyonlarını alabilirler.
- **Orijinal URL'ye Yönlendirme**: Servis, kullanıcıları kısaltılmış URL'den orijinal URL'ye yönlendirir.
- **İstatistikler**: Kullanıcılar, her kısaltılmış URL için erişim sayısı gibi istatistikleri görüntüleyebilirler.
- **URL Silme**: Kullanıcılar, ihtiyaç duydukları zaman kısaltılmış URL'lerini silebilirler.
- **Son Kullanma Tarihi**: Kısaltılmış URL'ler belirli bir süre sonra sona erebilir.
- **Benzersiz Kısaltılmış URL'ler**: Her orijinal URL yalnızca benzersiz bir kısaltılmış URL'ye kısaltılabilir.

## Kullanılan Teknolojiler

- **Golang**: Servisin ana programlama dili.
- **Fiber**: Golang için hızlı bir HTTP framework'ü.
- **GORM**: PostgreSQL veritabanı ile etkileşimde bulunan bir ORM.
- **Redis**: Önbellekleme ve URL sona erme yönetimi için kullanılır.
- **PostgreSQL**: Orijinal ve kısaltılmış URL'lerin saklandığı veritabanı.

## API Uç Noktaları

- `POST /shorten`: Uzun bir URL'yi kısalt.
- `GET /:short_url`: Orijinal URL'ye yönlendir.
- `GET /:short_url/stats`: Kısaltılmış URL için istatistikleri göster.
- `DELETE /:short_url`: Kısaltılmış URL'yi sil.

## Başlarken

### Gereksinimler

- Go 1.20 veya üzeri
- PostgreSQL
- Redis

### Kurulum

1. Depoyu klonlayın:
   ```bash
   git clone https://github.com/KadirErbas/url-shortener-app-with-go.git
   cd url-shortener-app-with-go

2. .env dosyasını ayarlayın:
   ```bash
   POSTGRES_HOST={ip adresiniz}
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD={şifreniz}
   POSTGRES_DB={db adı}
   POSTGRES_PORT=5432
   
   REDIS_HOST={ip adresiniz}
   REDIS_PORT=6380
   REDIS_PASSWORD={şifreniz}
   REDIS_DB=0
   
   ACTIVE_TIME_MINUTE=6
   
3. Uygulamayı çalıştırın:
   ```bash
   go run main.go
## Docker ile Çalıştırma
1. Docker ve Docker Compose'un yüklü olduğundan emin olun.
2. Docker Konteynerlerini başlatın:
   ```bash
   sudo docker-compose up --build

## Notlar
Uygulumanın çalışabilmesi için Postgresql ve Redis servislerinin çalışır durumda olması gerektiğini unutmayın.
