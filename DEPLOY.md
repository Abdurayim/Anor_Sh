# Botni VPS serverga o'rnatish

Ubuntu 22.04 yoki 24.04, root huquqi bilan SSH. 1 GB RAM va 10 GB disk yetarli.

Bot **polling** rejimida ishlaydi: domen, SSL sertifikat yoki ochiq port kerak emas. Webhook ixtiyoriy, u 10-bo'limda.

Barcha buyruqlar serverda `root` sifatida bajariladi. Bot esa root'dan emas, alohida `anorbot` foydalanuvchisidan ishlaydi.

Natijada serverda shunday tuzilma bo'ladi:

```
/opt/anor-bot/
├── src/                 # GitHub'dan olingan kod
├── bin/parent-bot       # yig'ilgan bot
├── .env                 # token va admin raqamlari (faqat anorbot o'qiydi)
├── data/parent_bot.db   # SQLite baza
├── backups/             # kunlik zaxira nusxalar
└── temp_docs/           # vaqtinchalik DOCX fayllar (bot o'zi yaratadi)
```

---

## 1. Serverni tayyorlash

```bash
apt update && apt upgrade -y
apt install -y git build-essential sqlite3 tzdata ca-certificates curl
timedatectl set-timezone Asia/Tashkent
```

`build-essential` (gcc) majburiy: SQLite drayveri C kodda yozilgan va `CGO` bilan yig'iladi.

## 2. Go o'rnatish

Loyiha Go 1.24.5 talab qiladi.

```bash
cd /tmp
curl -LO https://go.dev/dl/go1.24.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
source /etc/profile.d/go.sh
go version   # go version go1.24.5 linux/amd64
```

ARM server bo'lsa, `amd64` o'rniga `arm64` yuklab oling. Arxitekturani `uname -m` buyrug'i ko'rsatadi.

## 3. Foydalanuvchi va papkalar

```bash
useradd --system --home-dir /opt/anor-bot --shell /usr/sbin/nologin anorbot
mkdir -p /opt/anor-bot/{bin,data,backups,temp_docs}
```

## 4. Kodni olish

**Repo public bo'lsa:**

```bash
git clone -b feature/filiallar-va-tuzatishlar https://github.com/Abdurayim/Anor_Sh.git /opt/anor-bot/src
```

**Repo private bo'lsa**, avval faqat o'qish huquqli deploy key yarating:

```bash
ssh-keygen -t ed25519 -N "" -f /root/.ssh/anor_deploy
cat /root/.ssh/anor_deploy.pub
```

Chiqqan kalitni GitHub'ga qo'shing: repo → **Settings → Deploy keys → Add deploy key**. "Allow write access" belgisini qo'ymang. Keyin:

```bash
cat >> /root/.ssh/config <<'EOF'
Host github-anor
    HostName github.com
    User git
    IdentityFile /root/.ssh/anor_deploy
EOF

git clone -b feature/filiallar-va-tuzatishlar git@github-anor:Abdurayim/Anor_Sh.git /opt/anor-bot/src
```

> Branch `main` ga merge qilingandan keyin `-b feature/filiallar-va-tuzatishlar` o'rniga `-b main` yozing.

## 5. Botni yig'ish

```bash
cd /opt/anor-bot/src
CGO_ENABLED=1 go build -o /opt/anor-bot/bin/parent-bot ./cmd/bot
```

Migratsiyalar binary ichiga joylangan. Bot birinchi ishga tushganda bazani o'zi yaratadi va yangilaydi.

> `cmd/seed` faqat **demo ma'lumot** uchun. Ishlayotgan serverda uni ishga tushirmang.

## 6. `.env` fayli

```bash
cat > /opt/anor-bot/.env <<'EOF'
BOT_TOKEN=BU_YERGA_BOTFATHER_TOKENI

# Super admin: barcha filiallar statistikasini ko'radi, filial adminlarini qo'shadi
SUPER_ADMIN_PHONE=+998XXXXXXXXX

# Filial adminlari (ixtiyoriy, super admin ularni botdan ham qo'sha oladi)
ADMIN_PHONES_OLMAZOR=
ADMIN_PHONES_SERGELI=

DB_PATH=/opt/anor-bot/data/parent_bot.db

# Polling rejimi uchun bo'sh qoladi
WEBHOOK_URL=
WEBHOOK_SECRET=
API_TOKEN=
SERVER_PORT=8080
GIN_MODE=release
EOF

# Bot faqat o'z ma'lumotlariga yoza oladi; kod va binary root'niki bo'lib qoladi
chown anorbot:anorbot /opt/anor-bot /opt/anor-bot/.env
chown -R anorbot:anorbot /opt/anor-bot/data /opt/anor-bot/backups /opt/anor-bot/temp_docs
chmod 600 /opt/anor-bot/.env
```

Barcha o'zgaruvchilar tavsifi bilan [.env.example](.env.example) faylida.

## 7. Kompyuterdagi botni to'xtatish

Bitta token bilan bir vaqtda **faqat bitta bot** ishlashi mumkin. Aks holda loglarda `Conflict: terminated by other getUpdates request` xatosi chiqadi va xabarlar ikki bot o'rtasida bo'linib ketadi.

Kompyuteringizda ishlab turgan botni (`go run ./cmd/bot` yoki IDE) to'xtating.

**Mavjud bazani ko'chirish (ixtiyoriy).** Kompyuterdagi bazada demo ma'lumotlar bor, shuning uchun serverda toza bazadan boshlash tavsiya etiladi. Baribir ko'chirmoqchi bo'lsangiz, botni to'xtatgandan keyin kompyuterda:

```bash
sqlite3 parent_bot.db ".backup parent_bot.db.copy"
scp parent_bot.db.copy root@SERVER_IP:/opt/anor-bot/data/parent_bot.db
```

Keyin serverda:

```bash
chown anorbot:anorbot /opt/anor-bot/data/parent_bot.db
```

## 8. systemd xizmati

Bot serverda doim ishlashi, qayta yuklanishdan keyin o'zi yonishi va yiqilsa qayta ishga tushishi uchun:

```bash
cat > /etc/systemd/system/anor-bot.service <<'EOF'
[Unit]
Description=Anor maktab Telegram boti
After=network-online.target
Wants=network-online.target

[Service]
User=anorbot
Group=anorbot
WorkingDirectory=/opt/anor-bot
ExecStart=/opt/anor-bot/bin/parent-bot
Restart=always
RestartSec=5
Environment=TZ=Asia/Tashkent

# Xavfsizlik: bot faqat o'z papkasiga yoza oladi
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/opt/anor-bot

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now anor-bot
```

`WorkingDirectory` muhim: bot `.env` ni va `temp_docs` ni shu papkadan topadi.

## 9. Tekshirish

```bash
systemctl status anor-bot      # "active (running)" bo'lishi kerak
journalctl -u anor-bot -f      # jonli loglar (Ctrl+C bilan chiqiladi)
```

Logda quyidagilar chiqishi kerak:

```
✓ Database schema is up to date
✓ Bot authorized: @BOT_NOMI
✓ Admins initialized
📱 Bot is ready to receive messages via polling!
```

Keyin Telegram'da tekshiring:
1. Super admin raqami ochilgan akkauntdan botga `/start` yuboring.
2. Tilni tanlang va **"📱 raqamni ulashish"** tugmasini bosing. "👑 Super admin paneli" chiqishi kerak.
3. Panelda "➕ Admin qo'shish" orqali Olmazor va Sergeli adminlarini qo'shing.

**Firewall.** Polling rejimida tashqaridan kiruvchi port kerak emas, faqat SSH ochiq qoladi:

```bash
ufw allow OpenSSH
ufw enable
```

## 10. Webhook rejimi (ixtiyoriy)

Polling bitta server uchun yetarli. Webhook kerak bo'lsa, sizga domen kerak (masalan `bot.misol.uz`, DNS A yozuvi server IP'siga). Keyin:

```bash
apt install -y nginx certbot python3-certbot-nginx

cat > /etc/nginx/sites-available/anor-bot <<'EOF'
server {
    server_name bot.misol.uz;
    location /webhook { proxy_pass http://127.0.0.1:8080; }
    location /health  { proxy_pass http://127.0.0.1:8080; }
}
EOF
ln -s /etc/nginx/sites-available/anor-bot /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
certbot --nginx -d bot.misol.uz

ufw allow 'Nginx Full'
```

`.env` da quyidagilarni to'ldiring:

```
WEBHOOK_URL=https://bot.misol.uz
WEBHOOK_SECRET=<openssl rand -hex 32 natijasi>
```

So'ng botni qayta ishga tushiring: `systemctl restart anor-bot`.

`/api/admin/*` endpointlari tashqariga ochilmagan. Ular kerak bo'lsa, `API_TOKEN` ni to'ldiring va nginx'ga `location /api` qo'shing.

## 11. Kunlik zaxira nusxa

```bash
cat > /usr/local/bin/anor-bot-backup <<'EOF'
#!/bin/sh
set -e
DIR=/opt/anor-bot/backups
cd "$DIR"
sqlite3 /opt/anor-bot/data/parent_bot.db ".backup $DIR/parent_bot-$(date +%F).db"
find "$DIR" -name 'parent_bot-*.db' -mtime +14 -delete
EOF
chmod +x /usr/local/bin/anor-bot-backup

echo '0 3 * * * anorbot /usr/local/bin/anor-bot-backup' > /etc/cron.d/anor-bot-backup
```

Zaxira har kuni soat 03:00 da olinadi, oxirgi 14 kunlik nusxa saqlanadi. `.backup` buyrug'i bot ishlab turganda ham bazani xavfsiz nusxalaydi.

Zaxirani serverdan tashqariga ham ko'chirib turing, masalan haftada bir marta `scp` bilan kompyuterga.

Zaxiradan tiklash:

```bash
systemctl stop anor-bot
cp /opt/anor-bot/backups/parent_bot-2026-10-03.db /opt/anor-bot/data/parent_bot.db
rm -f /opt/anor-bot/data/parent_bot.db-wal /opt/anor-bot/data/parent_bot.db-shm
chown anorbot:anorbot /opt/anor-bot/data/parent_bot.db
systemctl start anor-bot
```

## 12. Yangilash (yangi kod chiqqanda)

```bash
sudo -u anorbot /usr/local/bin/anor-bot-backup   # avval zaxira
cd /opt/anor-bot/src
git pull
CGO_ENABLED=1 go build -o /opt/anor-bot/bin/parent-bot.new ./cmd/bot \
  && mv /opt/anor-bot/bin/parent-bot.new /opt/anor-bot/bin/parent-bot \
  && systemctl restart anor-bot
journalctl -u anor-bot -n 30
```

Yangi migratsiyalar qayta ishga tushganda avtomatik qo'llanadi.

## 13. Ko'p uchraydigan muammolar

| Belgi | Sabab va yechim |
|---|---|
| `Conflict: terminated by other getUpdates request` | Shu token bilan boshqa joyda ham bot ishlayapti (kompyuter, boshqa server). Ortiqchasini to'xtating. |
| `BOT_TOKEN is required` | `.env` topilmadi. Fayl `/opt/anor-bot/.env` da bo'lishi va service'da `WorkingDirectory=/opt/anor-bot` turishi kerak. |
| `unable to open database file` / `readonly database` | `chown -R anorbot:anorbot /opt/anor-bot/data` buyrug'ini bajaring. |
| Yig'ishda `gcc: not found` yoki `cgo` xatosi | `apt install build-essential`, keyin `CGO_ENABLED=1` bilan qayta yig'ing. |
| Admin panel ochilmaydi | Raqam `.env` da **+998** bilan yozilganini tekshiring, botni qayta ishga tushiring. Raqamni qo'lda yozmasdan **tugma** orqali ulashing. |
| Bot javob bermayapti | `systemctl status anor-bot` va `journalctl -u anor-bot -n 100` ni ko'ring. |
