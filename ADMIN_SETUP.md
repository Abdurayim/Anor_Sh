# 👨‍💼 Admin Setup Guide

## 🎯 Understanding Admins

**Maximum Admins:** 3 (enforced by database)

Admins can:
- ✅ Receive all complaint notifications
- ✅ Access admin panel via `/admin` command
- ✅ View all users and complaints
- ✅ Download complaint documents
- ✅ Access REST API endpoints

---

## 🔐 How Admin Authentication Works

Admins are identified by **phone numbers** configured in `.env`:

```env
ADMIN_PHONES=+998901234567,+998907654321,+998909876543
```

When the bot starts:
1. Reads admin phone numbers from `.env`
2. Adds them to the `admins` table in database
3. Matches incoming messages against these phone numbers

---

## 📝 Adding Admins

### Option 1: Before First Run (Recommended)

Edit `.env` file before starting the bot:

```env
# One admin
ADMIN_PHONES=+998901234567

# Two admins (no spaces!)
ADMIN_PHONES=+998901234567,+998907654321

# Three admins (maximum)
ADMIN_PHONES=+998901234567,+998907654321,+998909876543
```

Then run:
```bash
./parent-bot
```

### Option 2: After Bot is Running

1. **Stop the bot** (Ctrl+C)
2. **Edit `.env`** file and update `ADMIN_PHONES`
3. **Restart the bot:**
   ```bash
   ./parent-bot
   ```

---

## ⚠️ Important Rules

### Phone Number Format

**Must follow this format:**
- ✅ `+998901234567` (13 characters)
- ✅ Starts with `+998`
- ✅ Valid operator codes: 90, 91, 93, 94, 95, 97, 98, 99, 33, 88, 77

**Invalid formats:**
- ❌ `998901234567` (missing +)
- ❌ `+8 (998) 90-123-45-67` (has spaces/dashes)
- ❌ `+998 90 123 45 67` (has spaces)

### Multiple Admins

**Correct:**
```env
# NO spaces after commas!
ADMIN_PHONES=+998901234567,+998907654321,+998909876543
```

**Wrong:**
```env
# These will NOT work:
ADMIN_PHONES=+998901234567, +998907654321  # Has spaces
ADMIN_PHONES=+998901234567; +998907654321  # Wrong separator
ADMIN_PHONES=+998901234567
+998907654321  # Multiple lines
```

### Maximum Limit

The bot **enforces a maximum of 3 admins**:

- ✅ Trying to add 1-3 admins: Works fine
- ❌ Trying to add 4+ admins: **Error!** Bot will log a warning

---

## 🧪 Testing Admin Access

### Step 1: Register Admin Phone

Admins must first register in the bot like regular users:

1. Open Telegram
2. Find your bot
3. Send `/start`
4. Complete registration using the same phone number in `ADMIN_PHONES`

### Step 2: Verify Admin Access

Send `/admin` command:

- ✅ **Admin:** See admin panel with buttons
- ❌ **Not Admin:** Get "Access denied" message

---

## 📊 Admin Panel Features

Once you send `/admin`, you'll see:

### 👥 View Users
- Shows all registered parents
- Displays: Name, Class, Phone, Registration date
- Limited to 20 latest users (use API for more)

### 📋 View Complaints
- Shows all submitted complaints
- Displays: Status, Parent name, Preview, Date
- Status indicators:
  - ⏳ Pending
  - ✅ Reviewed
  - 📦 Archived

### 📊 View Statistics
- Total registered users
- Total complaints
- Pending complaints count
- Reviewed complaints count
- Completion percentage

---

## 🔔 Receiving Notifications

When a parent submits a complaint, **all admins** receive:

1. **Notification message:**
   ```
   🔔 Yangi shikoyat keldi!

   ID: #123
   Username: @parent_user
   ```

2. **DOCX Document** with:
   - Parent's name (child's name)
   - Class
   - Phone number
   - Full complaint text
   - Timestamp

The document filename format:
```
Shikoyat_ChildName_9A_sinf_2025-10-28.docx
```

---

## 🔧 Managing Admins

### View Current Admins

Using SQLite:
```bash
sqlite3 parent_bot.db "SELECT * FROM admins;"
```

Using API:
```bash
curl http://localhost:8080/api/admin/stats
```

### Remove an Admin

1. **Stop the bot**
2. **Edit `.env`** - remove the phone number
3. **Delete from database:**
   ```bash
   sqlite3 parent_bot.db "DELETE FROM admins WHERE phone_number = '+998901234567';"
   ```
4. **Restart the bot**

### Change Admin Phone Number

1. **Stop the bot**
2. **Edit `.env`** with new number
3. **Update database:**
   ```bash
   sqlite3 parent_bot.db "UPDATE admins SET phone_number = '+998909999999' WHERE phone_number = '+998901234567';"
   ```
4. **Restart the bot**

---

## 🌐 Admin API Endpoints

Admins can access these REST endpoints:

### Get All Users
```bash
curl http://localhost:8080/api/admin/users
```

Returns JSON with all registered users.

### Get All Complaints
```bash
curl http://localhost:8080/api/admin/complaints
```

Returns JSON with complaints + user info.

### Get Statistics
```bash
curl http://localhost:8080/api/admin/stats
```

Returns:
```json
{
  "total_users": 150,
  "total_complaints": 45,
  "pending_complaints": 12
}
```

### Health Check
```bash
curl http://localhost:8080/health
```

---

## 🔒 Security Notes

### Current Implementation

⚠️ **Admin API has NO authentication** (yet)

Anyone who can access your server can call the API endpoints.

**For production, you should:**
1. Add API key authentication
2. Use HTTPS
3. Restrict API access by IP
4. Add rate limiting

### Telegram Admin Commands

✅ **Secure** - Only users with matching phone numbers can use `/admin`

The bot checks:
1. User's Telegram ID
2. User's registered phone number
3. Matches against `ADMIN_PHONES` list

---

## 📱 Admin Workflow Example

### Morning Routine

```bash
# 1. Check for new complaints
# Open Telegram - check bot messages

# 2. View pending complaints
# Send /admin → Click "📋 Shikoyatlar"

# 3. Download complaint documents
# Click on document files sent by bot

# 4. Review and respond
# Contact parent via phone if needed
```

### Using API for Reports

```bash
# Export all users to file
curl http://localhost:8080/api/admin/users > users.json

# Export all complaints
curl http://localhost:8080/api/admin/complaints > complaints.json

# Check statistics
curl http://localhost:8080/api/admin/stats
```

---

## ❓ Troubleshooting

### "Access denied" when using /admin

**Check:**
1. ✅ Is your phone number in `.env` ADMIN_PHONES?
2. ✅ Did you register in the bot with that phone?
3. ✅ Is the format correct? (+998XXXXXXXXX)
4. ✅ Did you restart bot after changing .env?

**Fix:**
```bash
# 1. Stop bot
# 2. Edit .env
ADMIN_PHONES=+998901234567  # Your actual phone

# 3. Check database
sqlite3 parent_bot.db "SELECT * FROM admins;"

# 4. If not there, delete and recreate database
rm parent_bot.db
./parent-bot
```

### Not receiving notifications

**Check:**
1. ✅ Have you started the bot (`/start`)?
2. ✅ Is your Telegram ID linked to admin phone?
3. ✅ Check bot logs for errors

**Fix:**
```bash
# Check if admin is in database
sqlite3 parent_bot.db "SELECT * FROM admins WHERE phone_number = '+998901234567';"

# Verify telegram_id is set
# If NULL, re-register in the bot
```

### "Maximum 3 admins allowed" error

**This is intentional!** The system only allows 3 admins.

**To add a new admin:**
1. Remove an existing admin from `.env`
2. Delete from database
3. Add new admin to `.env`
4. Restart bot

---

## 💡 Best Practices

### 1. Document Your Admins

Keep a record:
```
Admin 1: School Principal - +998901234567
Admin 2: Vice Principal - +998907654321
Admin 3: Secretary - +998909876543
```

### 2. Regular Backups

Admins should ensure database is backed up:
```bash
# Daily backup script
cp parent_bot.db backups/backup_$(date +%Y-%m-%d).db
```

### 3. Monitor Activity

Check statistics regularly:
```bash
# Quick stats check
sqlite3 parent_bot.db "
SELECT
  (SELECT COUNT(*) FROM users) as total_users,
  (SELECT COUNT(*) FROM complaints) as total_complaints,
  (SELECT COUNT(*) FROM complaints WHERE status='pending') as pending;
"
```

### 4. Respond Quickly

Set up notifications:
- Keep Telegram notifications enabled
- Check bot messages at least twice daily
- Download and review complaints promptly

---

## 📞 Admin Support

### For Technical Issues

1. Check logs when running bot
2. Verify database integrity:
   ```bash
   sqlite3 parent_bot.db "PRAGMA integrity_check;"
   ```
3. Review documentation files

### For Configuration Help

Refer to:
- **START_HERE.md** - Quick setup
- **SQLITE_QUICKSTART.md** - Database guide
- **USAGE_GUIDE.md** - Complete manual

---

## 🎓 Advanced: Programmatic Admin Management

For advanced users who want to manage admins via code:

```go
// Add admin
admin, err := botService.AdminRepo.Create("+998901234567", "John Doe")

// Check if admin
isAdmin, err := botService.IsAdmin("+998901234567", telegramID)

// Get all admins
admins, err := botService.AdminRepo.GetAll()

// Delete admin
err := botService.AdminRepo.Delete("+998901234567")
```

---

**You're all set!** With up to 3 admins, your school can efficiently manage parent complaints. 🎉
