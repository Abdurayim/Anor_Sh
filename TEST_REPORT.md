# Test hisoboti: kamchiliklar va qo'shimchalar

Sana: 2026-10-01 · Bot: `@zizichayna_one_bot` · Rejim: polling · DB: SQLite (`parent_bot.db`, faqat 006 migratsiya)

## 1. Nima qilindi

- `.env` fayli `../Anor_bot/.env` dan nusxalandi (o'sha token va admin telefoni).
- Bot build qilindi va ishga tushirildi. DB yaratildi, admin qo'shildi.
- Demo data uchun **`cmd/seed/main.go`** yozildi. Uni qayta ishga tushirsa ham dublikat yaratmaydi:
  ```bash
  go run ./cmd/bot     # birinchi marta: sxema yaratiladi
  go run ./cmd/seed    # demo data qo'shiladi
  ```
  | Jadval | Soni | Izoh |
  |---|---|---|
  | classes | 6 | 5-A, 5-B, 6-A, 7-A, 9-B, 11-A |
  | teachers | 3 | Dilnoza Karimova `+998901112233`, Jasur Toshmatov `+998935556677`, Malika Yusupova `+998998887766` |
  | students | 18 | har bir sinfda 2–4 ta |
  | users (ota-onalar) | 4 | Telegram ID'lari soxta (`900000001`…). Ularga xabar yuborilmaydi, bu kutilgan holat |
  | test_results | 144 | 4 fan × 2 sana × 18 o'quvchi |
  | attendance | 180 | oxirgi 10 ta ish kuni |
  | announcements | 3 | 2 tasi admindan, 1 tasi o'qituvchidan (5-A, 5-B uchun) |
  | complaints / proposals | 2 / 2 | |
- **O'z-o'zini test qilish.** Telegram orqali foydalanuvchi sifatida yozib bo'lmaydi. Shuning uchun DB nusxasida 40 ta repository metodi va DOCX generatsiyasi chaqirib ko'rildi. Kod qo'lda ham audit qilindi.
  - ✅ Sinflar, o'qituvchilar, o'quvchilar, ota-onalar, davomat, baholar, e'lonlar, shikoyat va takliflarni o'qish ishlaydi.
  - ✅ Shikoyat, taklif va sinf baholari bo'yicha DOCX fayllar yaratiladi (~9 KB).
  - ❌ Quyidagi bo'limlardagi buglar topildi.

### Botni o'zingiz sinash
1. Botga `/start` yuboring, tilni tanlang va `.env` dagi admin raqamini ulashing. Siz admin bo'lasiz.
2. Agar o'qituvchi rolini sinamoqchi bo'lsangiz, admin panelidan **o'z** raqamingiz bilan o'qituvchi qo'shing (yoki seed'dagi telefonni o'zingiznikiga almashtiring).
3. Ota-ona rolini sinash uchun boshqa Telegram akkaunt kerak. Ro'yxatdan o'tishda istalgan sinfdagi o'quvchini tanlang.

---

## 2. Kritik (xavfsizlik)

| # | Joy | Muammo |
|---|---|---|
| K1 | [registration.go:56-60](internal/handlers/registration.go#L56-L60), [admin_link.go:50](internal/handlers/admin_link.go#L50) | Telefon raqami **matn ko'rinishida** qabul qilinadi va `Contact.UserID == From.ID` tekshirilmaydi. Kim admin raqamini yozib yuborsa, **admin bo'lib oladi**. O'qituvchi raqamini yozsa, o'qituvchi akkauntini egallaydi. ✅ Tasdiqlandi. |
| K2 | [main.go:136-170](cmd/bot/main.go#L136-L170) | `/api/admin/*` endpointlarida autentifikatsiya yo'q. Barcha telefonlar va shikoyatlar ochiq turibdi. |
| K3 | [main.go:118](cmd/bot/main.go#L118) | Webhook `secret_token` ni tekshirmaydi. URL'ni bilgan odam admin nomidan soxta update yubora oladi. |
| K4 | [admin.go:1320](internal/handlers/admin.go#L1320) va boshqa admin callback'lari | Callback handlerlarda admin tekshiruvi yo'q. Masalan, o'qituvchini o'chirish bir bosishda, tasdiqsiz bajariladi. ✅ Tasdiqlandi. |

## 3. Yuqori darajadagi buglar

| # | Joy | Muammo |
|---|---|---|
| Y1 | [student_repo.go:246](internal/repository/student_repo.go#L246) + 006 sxema | `parent_students.student_id UNIQUE`, 007 migratsiya esa ishga tushirilmaydi. Natijada **ikkinchi ota-ona** (masalan, onasi) bog'lanmaydi: `INSERT OR IGNORE` jim qoladi, bot esa "bog'landi" deb javob beradi. ✅ Testda tasdiqlandi: 0 qator yozildi. |
| Y2 | [complaint_repo.go:21](internal/repository/complaint_repo.go#L21), [proposal_repo.go:21](internal/repository/proposal_repo.go#L21) | Tanlangan farzandning `student_id` si bazaga **yozilmaydi**. Admin shikoyat qaysi bola haqida ekanini ko'rmaydi. ✅ Tasdiqlandi. |
| Y3 | [models/complaint.go:45](internal/models/complaint.go#L45) | `StatusArchived = "archived"` qiymati DB CHECK cheklovida yo'q, shuning uchun status o'zgartirilganda xato beradi. ✅ Testda tasdiqlandi. |
| Y4 | [router.go:599](internal/handlers/router.go#L599) | `data[:17] == "view_grades_class_"`: literal 18 belgidan iborat, shuning uchun shart **hech qachon** bajarilmaydi. ✅ Tasdiqlandi. |
| Y5 | router.go | `complaints_page_%d` va `proposals_page_%d` uchun marshrut yo'q, sahifalash ishlamaydi. |
| Y6 | [announcement_repo.go:21](internal/repository/announcement_repo.go#L21) | `teacher_id` va `announcement_classes` yozilmaydi. O'qituvchi o'z e'lonini tahrirlay/o'chira olmaydi, sinfga mo'ljallangan e'lon esa hammaga ketadi. |
| Y7 | [announcement_repo.go:289](internal/repository/announcement_repo.go#L289) | Ichma-ich `db.Query` va `MaxOpenConns(1)` birga kelganda **deadlock** bo'ladi. Hozir bu joyga yo'l yo'q, lekin ulansa bot qotib qoladi. |
| Y8 | attendance.go, teacher_management.go, test_results.go | O'qituvchi **istalgan sinf** bilan ishlay oladi: `teacher_classes` tekshirilmaydi. `/edit_grade` va `/delete_grade` boshqalarning baholariga ham ta'sir qiladi. |
| Y9 | [attendance.go:443](internal/handlers/attendance.go#L443) | Har qanday ota-ona istalgan sinf davomatini ko'ra oladi. |
| Y10 | student_selection.go | Istalgan foydalanuvchi istalgan o'quvchini o'ziga bog'lab, uning baholarini ko'ra oladi. 006 sxemada bu haqiqiy ota-onani butunlay bloklab qo'yadi. |
| Y11 | [test_result_repo.go:244](internal/repository/test_result_repo.go#L244) | `/edit_grade` fan nomini bo'sh satrga almashtirib yuboradi: `COALESCE("", …)` ishlamaydi. |
| Y12 | [timetable.go:103](internal/handlers/timetable.go#L103) | Rasm ko'rinishidagi jadval `sendDocument` orqali yuboriladi, natijada ota-onaga yetib bormaydi. |
| Y13 | test_results.go:350, admin.go:254 | HTML rejimdagi xabarda `<ID>` va `<sinf nomi>` bor. Telegram xabarni rad etadi va foydalanuvchiga hech narsa kelmaydi. |
| Y14 | admin.go:742 | Sinf tasdiqsiz o'chiriladi. Kaskad bo'yicha barcha o'quvchilar, baholar va davomat ham o'chib ketadi. |
| Y15 | main.go | `recover()` yo'q, bitta panic butun botni to'xtatadi. Polling ketma-ket ishlaydi, shuning uchun DOCX yaratilayotganda hamma kutib qoladi. Webhook'da esa poyga holatlari bor (Confirm'ni ikki marta bossa ikki shikoyat yaratiladi). |

## 4. O'rta / past darajadagi kamchiliklar

- **HTML escape yo'q.** E'lon, ism yoki fanda `&` yoki `<` bo'lsa, butun broadcast muvaffaqiyatsiz tugaydi.
- **Shikoyat matni buziladi.** `validator/text.go` SQL so'zlarini olib tashlaydi va HTML-escape qiladi. Natijada `o'quvchi` so'zi DOCX'da `o&#39;quvchi` bo'lib chiqadi.
- **Broadcast.** Faqat birinchi 1000 foydalanuvchiga yuboriladi. Rate-limit (≈30 msg/s) va 429 qayta urinishi yo'q.
- **4096 belgi limiti.** Butun sinf baholari yoki davomati shu limitdan oshsa, xabar yuborilmaydi.
- **Eksport.** "Oxirgi 7 kun" va "Hammasi" bir xil ishlaydi. Custom sana kiritilsa ham e'tiborga olinmaydi.
- **Davomat.** Qayta saqlanganda ota-onalarga qayta xabar ketadi. "Yo'q"dan "bor"ga tuzatilganda esa xabar ketmaydi. Yozuvlar tranzaksiyasiz bajariladi.
- **Vaqt zonasi.** Ba'zi joylarda server vaqti, ba'zilarida Asia/Tashkent ishlatiladi. `LoadLocation` xatosi e'tiborsiz qoldirilgan.
- **Callback data 64 baytdan oshishi mumkin.** Sinf nomi uzunligi cheklanmagan, uzun nom klaviaturani buzadi.
- **Temp fayllar.** Nomlar faqat sanaga bog'liq, shuning uchun parallel yaratilganda to'qnashadi. `CleanTempDirectory` va `CleanOldStates` hech qachon chaqirilmaydi.
- **O'chirilgan o'quvchilar.** Soft-delete qilingan o'quvchi "Farzandlarim"da ko'rinishda qoladi va 4 ta bola limitiga hisoblanadi.
- **O'qituvchi va admin huquqini olib tashlab bo'lmaydi.** O'qituvchini deaktivatsiya qilib bo'lmaydi. `.env` dan olib tashlangan admin DB'da qolib ketadi.
- **Matnlar.** Ruscha matnlarda lotin harflari bor: "Ошибка базы danных". Ko'p matn i18n'siz, faqat o'zbekcha yozilgan.
- **Tartib.** `test_results.go` map ustida aylanadi, shuning uchun o'quvchilar tartibi tasodifiy chiqadi.
- **Admin tugmasi yo'qoladi.** Muvaffaqiyat ekranlaridan keyin admin tugmasi ko'rinmay qoladi.
- **Sahifalash yo'q.** Foydalanuvchilar, shikoyatlar va o'qituvchilar ro'yxatida faqat "...va yana N ta" deb chiqadi.
- **Ishlatilmaydigan funksiyalar.** O'qituvchining "mening e'lonlarim", sinf baho/davomat ko'rinishi va o'qituvchi sozlamalari routerga ulanmagan.

## 5. Tavsiya etiladigan qo'shimchalar

1. **Shikoyat/taklif statusini boshqarish.** Admin javob yozsa va status o'zgartirsa, ota-onaga xabar boradi.
2. **Bola bog'lashni tasdiqlash.** Admin tasdiqlaydi yoki har bir o'quvchiga maxsus kod beriladi. Bu Y10 muammosini yopadi.
3. **O'qituvchini sinfga biriktirish UI'si.** Biriktirish har bir o'qituvchi amalida tekshiriladi.
4. **Haqiqiy eksport (XLSX/DOCX) sana oralig'i bilan.** `pkg/docx` da `GenerateClassTestResults` va `GenerateTodayAttendance` allaqachon tayyor.
5. **Sinfga mo'ljallangan e'lonlar** va jadval o'zgarganda push xabar.
6. **Sozlamalarda tilni almashtirish** (`UserRepo.Update` mavjud).
7. **Xavfsizlik:** webhook secret, API uchun token, har bir foydalanuvchi update'larini ketma-ket qayta ishlash, broadcast navbati.
8. **Rejalashtirilgan vazifalar:** temp va state tozalash, "bugun davomat olinmadi" eslatmasi, haftalik baho xulosasi.
9. **Ota-ona tomonidan sabab bildirish** ("bugun kasal"), uy vazifasi, ota-ona va o'qituvchi o'rtasida yozishma, o'chirishlar audit logi.
10. **Migratsiya tizimi.** `main.go` faqat 006 ni ishga tushiradi. Versiya jadvali bilan 007+ migratsiyalarni ketma-ket bajaradigan mexanizm kerak.

## 6. Birinchi navbatda tuzatish tartibi

1. K1: telefonni faqat `request_contact` orqali qabul qilish va `Contact.UserID == From.ID` ni tekshirish.
2. K4 va Y8–Y10: barcha callback'larda rolni tekshirish.
3. Y1: 007 migratsiyani ishga tushirish (migratsiya runner orqali).
4. Y2, Y3, Y4, Y5, Y11, Y12, Y13: har biri kichik, bir necha qatorlik tuzatish.
5. K2, K3: API auth va webhook secret.

---

## 7. Yangilanish (2026-10-02, branch `feature/filiallar-va-tuzatishlar`)

### Filiallar: Olmazor va Sergeli
- Yangi `branches` jadvali qo'shildi (`009_branches.sql`). Har bir sinf, o'qituvchi, ota-ona, admin va e'lon bitta filialga tegishli.
- O'quvchilar, baholar, davomat va dars jadvali filialni sinfidan oladi. Shikoyat va takliflar esa ota-onadan oladi.
- Sinf nomi endi har bir filial ichida alohida unikal: ikkala filialda ham "5-A" bo'lishi mumkin.
- Har bir filialning o'z admini bor. Ular `.env` da beriladi: `ADMIN_PHONES_OLMAZOR`, `ADMIN_PHONES_SERGELI`. Eski `ADMIN_PHONES` o'zgaruvchisi Olmazor deb hisoblanadi.
- Admin faqat o'z filialini ko'radi va boshqaradi: sinflar, o'quvchilar, ota-onalar, o'qituvchilar, baholar, davomat, e'lonlar, shikoyat va takliflar, statistika va eksport.
- O'qituvchi faqat o'z filiali sinflari bilan ishlaydi.
- Ota-ona ro'yxatdan o'tishda filialni tanlaydi. Shundan keyin faqat shu filial sinflari va o'quvchilarini ko'radi.
- Shikoyat va taklif faqat ota-onaning filiali adminlariga boradi.
- E'lonlar:
  - admin e'loni o'z filialining barcha ota-onalariga boradi;
  - o'qituvchi e'loni faqat tanlangan sinflar ota-onalariga boradi.
- Eski ma'lumotlar Olmazor filialiga o'tkazildi. Hech narsa yo'qolmagani testda tekshirildi.
- API: `/api/admin/*?branch=olmazor|sergeli`.

### Tuzatilgan buglar
| # | Holat |
|---|---|
| K1 | ✅ Telefon faqat **o'z kontaktini ulashish** tugmasi orqali qabul qilinadi (`Contact.UserID == From.ID`). Admin endi telefon raqami bilan emas, faqat bog'langan Telegram ID bilan taniladi. |
| K2 | ✅ `/api/admin/*` uchun `API_TOKEN` kerak (`Authorization: Bearer ...`). U bo'sh bo'lsa, API o'chiq. |
| K3 | ✅ `WEBHOOK_SECRET` qo'shildi, webhook header'ni tekshiradi. |
| K4, Y8, Y9, Y10 | ✅ Barcha callback va buyruqlar bitta joyda tekshiriladi: [auth.go](internal/handlers/auth.go). Rol va filial tekshiriladi, ota-ona esa faqat o'z farzandi bilan ishlay oladi. |
| Y1 | ✅ Migratsiya runner yozildi (`schema_migrations`, migratsiyalar binary ichida). 007 qo'llandi, endi ikkinchi ota-ona bog'lanadi. |
| Y2 | ✅ Shikoyat va taklifda `student_id` saqlanadi. Admin ro'yxatida farzand va sinf ko'rinadi. |
| Y3 | ✅ Statuslar CHECK cheklovga moslandi (`resolved` / `implemented`). |
| Y4, Y5 | ✅ `view_grades_class_` marshruti ishlaydi. Shikoyat va takliflar sahifalanadi. |
| Y6 | ✅ E'londa `teacher_id`, `branch_id` va `announcement_classes` saqlanadi (tranzaksiyada). |
| Y7 | ✅ Ichma-ich query olib tashlandi, deadlock endi bo'lmaydi. |
| Y11 | ✅ `/edit_grade` fan nomini o'chirmaydi (`NULLIF`). O'qituvchi faqat o'z filiali baholarini o'zgartira oladi. |
| Y12 | ✅ Rasm ko'rinishidagi dars jadvali rasm sifatida yuboriladi. |
| Y13 | ✅ HTML'dagi `<ID>` kabi matnlar escape qilindi. E'lon matni ham escape qilinadi. |
| Y14 | ✅ Sinf yoki o'qituvchini o'chirishdan oldin tasdiq so'raladi. |
| Y15 | ✅ Panic'dan himoya (`recover`) qo'shildi. Bitta foydalanuvchining update'lari ketma-ket qayta ishlanadi, shuning uchun Confirm'ni ikki marta bosish dublikat yaratmaydi. |
| Boshqa | ✅ Broadcast'da 1000 limiti yo'q, sekundiga ~20 ta xabar yuboriladi. ✅ 4096 belgidan uzun xabarlar bo'linadi. ✅ `TruncateText` kirill harflarini buzmaydi. ✅ `.env` dan olib tashlangan admin huquqini yo'qotadi. ✅ Deaktiv o'qituvchi tizimga kira olmaydi. ✅ Soft-delete qilingan o'quvchi "Farzandlarim"da ko'rinmaydi. ✅ Ruscha matndagi lotin harflari tuzatildi. |

### Testlar
`go test ./...` buyrug'i quyidagilarni tekshiradi:
- ruxsatlar: 23 ta holat, jumladan boshqa filialning sinfi, o'qituvchisi, o'quvchisi va e'loni;
- filiallar bo'yicha ajratish;
- ota-onaga qaysi e'lonlar ko'rinishi;
- `.env` bilan admin sinxronizatsiyasi;
- uzun xabarni bo'lish.

Migratsiya eski bazaning nusxasida ham sinab ko'rildi.

### Hali qolgan ishlar (bo'lim 4–5 dan)
- Eksportda "oxirgi 7 kun" va custom sana ishlamaydi.
- Davomat qayta saqlanganda ota-onalarga qayta xabar ketadi.
- Vaqt zonasi bir xil emas.
- Temp fayllar tozalanmaydi.
- Shikoyat matni haddan tashqari sanitizatsiya qilinadi.
- O'qituvchini sinfga biriktirish UI'si yo'q. Hozir o'qituvchi o'z filialining barcha sinflarini ko'radi.
- Shikoyat statusini o'zgartirish UI'si yo'q.
- Bo'lim 5 dagi qo'shimchalar.

---

## 8. Super admin (2026-10-03)

- `.env` dagi `SUPER_ADMIN_PHONE` raqamiga **super admin** roli beriladi (`010_super_admin.sql`).
- Super admin **faqat ko'radi**: umumiy dashboard (jami va har filial bo'yicha), filial tafsilotlari, so'nggi shikoyat va takliflar.
  - Filial tafsilotlarida sinflar, o'quvchilar soni va adminlar ko'rinadi.
  - Dashboard'da ota-onalar, o'quvchilar, o'qituvchilar, sinflar, adminlar, bugungi davomat, 30 kunlik o'rtacha baho va shikoyat/takliflar soni bor.
- Super admin sinf, o'quvchi, o'qituvchi yoki e'lon qo'sha olmaydi. Testda tekshirildi.
- Super admin **filial adminlarini qo'shadi**: filialni tanlaydi va raqamni kiritadi. Bitta filialda ko'pi bilan 3 ta admin bo'ladi.
  - O'qituvchi yoki boshqa admin raqami admin qilib qo'shilmaydi.
  - Bot orqali qo'shilgan adminni tasdiq bilan o'chirish mumkin.
  - `.env` dagi adminlarni faqat `.env` dan o'chirish mumkin.
- Bot orqali qo'shilgan adminlar qayta ishga tushirishda saqlanib qoladi. `.env` dagilar esa avvalgidek `.env` bilan sinxronlanadi.
- Filial admini o'qituvchini faqat o'z filialiga qo'shadi. Admin raqamini o'qituvchi qilib bo'lmaydi.
- Kirish: `/start`, keyin "📱 raqamni ulashish" tugmasi, keyin "👑 Super admin paneli" tugmasi (yoki `/superadmin` buyrug'i).
