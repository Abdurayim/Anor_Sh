# 🚀 SQLite Quick Start - SUPER SIMPLE!

## Why SQLite?

✅ **No installation needed!**
✅ **No passwords or configuration!**
✅ **Just one file - easy backups!**
✅ **Fast enough for thousands of users!**
✅ **Perfect for school use!**

---

## 🎯 3-Step Setup (5 Minutes!)

### Step 1: Get Your Bot Token

1. Open Telegram, find **@BotFather**
2. Send: `/newbot`
3. Follow the prompts
4. Copy your token: `1234567890:ABCdefGHIjkl...`

### Step 2: Create `.env` File

```bash
cd /Users/abdurayim/Desktop/PROJECTS/parent-bot
cp .env.example .env
nano .env  # or any text editor
```

**Paste this (replace with YOUR values):**

```env
# Get this from @BotFather
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Your phone number (will be admin)
ADMIN_PHONES=+998901234567

# That's it! These are optional:
DB_PATH=parent_bot.db
SERVER_PORT=8080
GIN_MODE=debug
```

**That's all you need!** No database passwords, no PostgreSQL setup!

### Step 3: Run the Bot!

```bash
./parent-bot
```

**Expected output:**
```
✓ Connected to database
✓ Database migrations completed
✓ Bot authorized: @YourBotName
✓ Admins initialized
🔄 Starting in POLLING mode (for local testing)
📱 Bot is ready to receive messages via polling!
```

---

## 🎉 Test It Now!

1. Open Telegram
2. Search for your bot
3. Send: `/start`
4. Follow the registration!

---

## 📁 What Just Happened?

When you run the bot for the first time:

1. ✅ Creates `parent_bot.db` file (your database!)
2. ✅ Creates all tables automatically
3. ✅ Adds you as an admin
4. ✅ Starts listening for messages

**The database is just a single file!** Easy to:
- 📦 Backup (just copy `parent_bot.db`)
- 📤 Transfer (copy to another computer)
- 🗑️ Reset (delete the file and restart)

---

## 🔧 Common Questions

### Where is my data stored?

In `parent_bot.db` file in the same folder as your bot.

### How do I backup my data?

```bash
# Simply copy the file!
cp parent_bot.db parent_bot_backup_2025-10-27.db

# Or use a script
DATE=$(date +%Y-%m-%d)
cp parent_bot.db "backups/parent_bot_$DATE.db"
```

### How do I reset everything?

```bash
# Delete the database file
rm parent_bot.db

# Run the bot again - it will create a fresh database
./parent-bot
```

### Can I view the database?

Yes! Install SQLite browser:

**macOS:**
```bash
brew install --cask db-browser-for-sqlite
```

**Ubuntu/Linux:**
```bash
sudo apt install sqlitebrowser
```

Then open `parent_bot.db` with the browser!

### How many users can it handle?

**Easily 10,000+ users** for a school bot. SQLite is perfect for:
- ✅ Small to medium schools (500-5,000 students)
- ✅ Single server deployments
- ✅ Low concurrent writes (complaint submissions)
- ✅ Many reads (viewing data)

---

## 📊 Performance Comparison

| Database | Setup Time | For 500 Parents | For 5,000 Parents |
|----------|-----------|----------------|------------------|
| SQLite | 0 minutes | ✅ Perfect | ✅ Great |
| PostgreSQL | 30+ minutes | ✅ Overkill | ✅ Good |

**For your school bot: SQLite is the smart choice!**

---

## 🔒 Database File Location

You can change where the database is stored:

```env
# Same folder (default)
DB_PATH=parent_bot.db

# Absolute path
DB_PATH=/var/lib/parent-bot/database.db

# In a data folder
DB_PATH=data/complaints.db
```

Just make sure the folder exists!

---

## 🚀 Migration from PostgreSQL

If you already have data in PostgreSQL and want to switch:

```bash
# 1. Export from PostgreSQL
pg_dump parent_bot > backup.sql

# 2. Convert to SQLite (manual process)
# Use tools like pgloader or convert manually

# 3. Or start fresh with SQLite
# (Easier for new deployments)
```

---

## 💡 Pro Tips

### 1. Automatic Backups

Create a backup script:

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y-%m-%d_%H-%M-%S)
cp parent_bot.db "backups/parent_bot_$DATE.db"
echo "Backup created: parent_bot_$DATE.db"
```

Run daily with cron:
```bash
# Run backup every day at 2 AM
0 2 * * * /path/to/backup.sh
```

### 2. View Data Quickly

```bash
# Install sqlite3 command-line tool
brew install sqlite3  # macOS
sudo apt install sqlite3  # Linux

# View users
sqlite3 parent_bot.db "SELECT * FROM users;"

# Count complaints
sqlite3 parent_bot.db "SELECT COUNT(*) FROM complaints;"

# View latest complaints
sqlite3 parent_bot.db "SELECT * FROM complaints ORDER BY created_at DESC LIMIT 5;"
```

### 3. Database Browser (GUI)

Download: https://sqlitebrowser.org/

Open your `parent_bot.db` file and you can:
- View all tables
- Edit data manually
- Run queries
- Export to CSV/Excel

---

## ⚠️ Important Notes

### When to Use SQLite

✅ School with 100-5,000 parents
✅ Single server deployment
✅ Simple setup needed
✅ Easy backups required
✅ Low concurrent writes

### When to Consider PostgreSQL

⚠️ Multiple servers (horizontal scaling)
⚠️ Very high concurrent writes (>100/second)
⚠️ Need advanced database features
⚠️ Multiple applications accessing same database

**For 99% of schools: SQLite is perfect!**

---

## 🎓 Learn More

### SQLite Commands

```bash
# Open database
sqlite3 parent_bot.db

# Show tables
.tables

# Describe table structure
.schema users

# Pretty output
.mode column
.headers on

# Run query
SELECT * FROM users;

# Exit
.exit
```

### Useful Queries

```sql
-- Total users
SELECT COUNT(*) FROM users;

-- Total complaints
SELECT COUNT(*) FROM complaints;

-- Pending complaints
SELECT COUNT(*) FROM complaints WHERE status = 'pending';

-- Recent registrations
SELECT child_name, child_class, registered_at
FROM users
ORDER BY registered_at DESC
LIMIT 10;

-- Most active complainers
SELECT u.child_name, COUNT(c.id) as complaint_count
FROM users u
LEFT JOIN complaints c ON u.id = c.user_id
GROUP BY u.id
ORDER BY complaint_count DESC
LIMIT 10;
```

---

## 🎉 You're Done!

That's it! SQLite is:
- ✅ Simple
- ✅ Fast
- ✅ Reliable
- ✅ Perfect for your school bot

No PostgreSQL headaches!
No password management!
Just one file!

**Start the bot and enjoy!** 🚀

---

## 📞 Troubleshooting

### "Database is locked"

This happens if multiple processes try to write simultaneously. SQLite handles this automatically by waiting. If it persists:

```bash
# Check if bot is already running
ps aux | grep parent-bot

# Kill old process if found
killall parent-bot
```

### "Permission denied"

```bash
# Make sure you have write permissions
chmod 666 parent_bot.db
```

### "No such table"

The database file exists but tables weren't created. Delete and restart:

```bash
rm parent_bot.db
./parent-bot
```

---

**Questions?** Check the main README.md or other documentation files!
