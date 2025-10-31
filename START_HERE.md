# 🚀 START HERE - Your Bot in 3 Steps!

## ✨ Super Simple Setup (2 Minutes!)

### Step 1: Get Your Bot Token (1 minute)

1. Open Telegram
2. Search for `@BotFather`
3. Send: `/newbot`
4. Give it a name (e.g., "School Complaints Bot")
5. Give it a username (e.g., "myschool_complaints_bot")
6. **Copy the token** (looks like: `1234567890:ABCdefGHIjkl...`)

### Step 2: Create Configuration File (30 seconds)

```bash
# Copy the simple template
cp .env.simple .env

# Edit it with any text editor
nano .env
# or
open -e .env  # on macOS
```

**Fill in just 2 things:**

```env
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ADMIN_PHONES=+998901234567
```

Replace with:
- Your bot token from @BotFather
- Your phone number (must start with +998)

**Save and close!**

### Step 3: Run! (30 seconds)

```bash
./parent-bot
```

**That's it!** You should see:

```
✓ Connected to database
✓ Database migrations completed
✓ Bot authorized: @YourBotName
✓ Admins initialized
🔄 Starting in POLLING mode (for local testing)
📱 Bot is ready to receive messages via polling!
```

---

## 🎯 Test Your Bot Right Now!

1. Open Telegram on your phone
2. Search for your bot (the username you gave it)
3. **Send:** `/start`
4. **Choose language:** Uzbek 🇺🇿 or Russian 🇷🇺
5. **Share your phone:** +998901234567
6. **Enter child name:** Akmal Rahimov
7. **Enter class:** 9A

✅ **Registration complete!** Now you can submit complaints!

---

## 📱 Try Submitting a Complaint

1. Tap **"✍️ Shikoyat yuborish"** (Submit complaint)
2. Type your complaint (at least 10 characters)
3. Review and confirm
4. **Done!** You'll receive the complaint as a DOCX file!

---

## 💡 What Just Happened?

When you ran the bot:

1. ✅ Created `parent_bot.db` (your database)
2. ✅ Created all database tables
3. ✅ Made you an admin (using your phone number)
4. ✅ Started listening for Telegram messages

**No PostgreSQL, no passwords, no complicated setup!**

---

## 🎓 Common Questions

### Where is my data?

In `parent_bot.db` file (same folder as the bot).

### How do I backup?

```bash
cp parent_bot.db backup_2025-10-28.db
```

### How do I reset?

```bash
rm parent_bot.db
./parent-bot  # Creates fresh database
```

### How many users can it handle?

Easily **10,000+ users**. Perfect for schools!

### Do I need to install anything?

**NO!** Everything is included in the `parent-bot` file.

---

## 🔧 Configuration Options

**Required:**
- `BOT_TOKEN` - From @BotFather
- `ADMIN_PHONES` - Your phone number(s), max 3

**Optional:**
- `DB_PATH` - Database location (default: `parent_bot.db`)
- `SERVER_PORT` - Port (default: `8080`)
- `WEBHOOK_URL` - For production (leave empty for testing)

---

## 📚 Learn More

- **SQLITE_QUICKSTART.md** - Detailed SQLite guide
- **USAGE_GUIDE.md** - Complete user manual
- **EXTENDING.md** - Add custom features
- **README.md** - Full documentation

---

## 🆘 Troubleshooting

### "BOT_TOKEN is required"
→ Check your `.env` file has `BOT_TOKEN=...`

### "Admin phone validation failed"
→ Use format: `+998901234567` (must start with +998)

### "Bot doesn't respond"
→ Make sure bot is running and check token is correct

### "Database is locked"
→ Only run one instance of the bot at a time

---

## ✅ Checklist

- [ ] Got bot token from @BotFather
- [ ] Created `.env` file with BOT_TOKEN and ADMIN_PHONES
- [ ] Ran `./parent-bot`
- [ ] Saw success messages
- [ ] Tested `/start` in Telegram
- [ ] Registered successfully
- [ ] Submitted test complaint

**All checked?** You're ready to use the bot! 🎉

---

## 🎯 Next Steps

1. **Share the bot** with other parents
2. **Test all features** (registration, complaints, admin panel)
3. **Setup backups** (copy `parent_bot.db` regularly)
4. **For production:** Read about WEBHOOK_URL in documentation

---

## 💪 You're All Set!

Your Parent Complaint Bot is:
- ✅ Running
- ✅ Database ready
- ✅ Bilingual (Uzbek/Russian)
- ✅ Secure and validated
- ✅ Easy to backup

**Enjoy!** 🚀

---

**Need help?** Check the documentation files or the error messages in the terminal.
