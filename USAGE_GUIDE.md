# Parent Complaint Bot - Usage Guide

## For End Users (Parents)

### Getting Started

1. **Find the bot** on Telegram (ask your school admin for the bot username)
2. **Start a conversation**: Send `/start`
3. **Choose your language**: Uzbek 🇺🇿 or Russian 🇷🇺

### Registration Process

After starting the bot, you'll go through a simple registration:

#### Step 1: Language Selection
- Tap **"🇺🇿 O'zbek"** for Uzbek
- Tap **"🇷🇺 Русский"** for Russian

#### Step 2: Phone Number
- **Method 1**: Tap the "📱 Share Phone Number" button (recommended)
- **Method 2**: Type your number manually: `+998901234567`

**Requirements**:
- Must start with `+998`
- Must be a valid Uzbek number
- Format: `+998XXXXXXXXX` (13 digits total)

#### Step 3: Child's Name
- Enter your child's full name
- Example: `Akmal Rahimov`

**Requirements**:
- 2-100 characters
- Only letters allowed
- No numbers or special symbols

#### Step 4: Child's Class
- Enter the class your child attends
- Example: `9A` or `11B`

**Requirements**:
- Grade: 1-11
- Letter: A-Z
- Format: `9A`, `10B`, etc.

### Submitting a Complaint

Once registered, you'll see the main menu:

#### Method 1: Using Button
1. Tap **"✍️ Shikoyat yuborish"** (Submit complaint)
2. Write your complaint (minimum 10 characters)
3. Review the preview
4. Tap **"✅ Tasdiqlash"** (Confirm) to send

#### Method 2: Using Command
1. Send `/complaint`
2. Follow the same steps as above

### What Happens After Submission?

1. ✅ Your complaint is automatically converted to a DOCX document
2. 📄 The document includes:
   - Your child's name and class
   - Your phone number
   - Full complaint text
   - Timestamp
3. 🔔 All school admins receive a notification
4. 📥 Admins get the document for review
5. ⏳ You can track the status in "My Complaints"

### Viewing Your Complaints

**Method 1**: Tap **"📋 Mening shikoyatlarim"** (My complaints)
**Method 2**: Send `/mycomplaints` (not implemented yet)

You'll see:
- ⏳ Pending complaints
- ✅ Reviewed complaints
- Date and preview of each

### Viewing Your Settings

**Method 1**: Tap **"⚙️ Sozlamalar"** (Settings)

You'll see:
- Your child's name
- Class
- Phone number
- Language preference

### Available Commands

- `/start` - Start the bot or reset
- `/help` - Get help information
- `/complaint` - Submit a new complaint
- `/admin` - Admin panel (admins only)

---

## For Administrators

### Admin Access

Admins are identified by phone numbers configured in the server's `.env` file:
```env
ADMIN_PHONES=+998901234567,+998907654321,+998909876543
```

**Maximum 3 admins allowed.**

### Receiving Notifications

When a parent submits a complaint, you receive:
1. 🔔 Notification message with:
   - Parent's name (child's name)
   - Class
   - Phone number
   - Telegram username (if available)
2. 📄 DOCX file ready for download

### Using Admin Panel

Send `/admin` to access the admin panel with these options:

#### 1. View Users (👥 Foydalanuvchilar)
- Lists all registered parents
- Shows: Name, Class, Phone, Registration date
- Limited to last 20 users
- Use API for full list

#### 2. View Complaints (📋 Shikoyatlar)
- Lists all submitted complaints
- Shows: Status, Parent name, Preview, Date
- Status indicators:
  - ⏳ Pending
  - ✅ Reviewed
  - 📦 Archived

#### 3. View Statistics (📊 Statistika)
- Total registered users
- Total complaints
- Pending complaints
- Reviewed complaints
- Completion percentage

### Using Admin API

The bot provides REST API endpoints for advanced management:

#### Get All Users
```bash
curl http://your-server:8080/api/admin/users
```

Response:
```json
{
  "users": [
    {
      "id": 1,
      "telegram_id": 123456789,
      "telegram_username": "john_doe",
      "phone_number": "+998901234567",
      "child_name": "Akmal Rahimov",
      "child_class": "9A",
      "language": "uz",
      "registered_at": "2025-10-27T10:30:00Z"
    }
  ]
}
```

#### Get All Complaints
```bash
curl http://your-server:8080/api/admin/complaints
```

Response:
```json
{
  "complaints": [
    {
      "id": 1,
      "user_id": 1,
      "complaint_text": "...",
      "telegram_file_id": "BQACAgIAAxkBAAI...",
      "filename": "Shikoyat_Akmal_Rahimov_9A_sinf_2025-10-27.docx",
      "created_at": "2025-10-27T11:00:00Z",
      "status": "pending",
      "user_telegram_id": 123456789,
      "phone_number": "+998901234567",
      "child_name": "Akmal Rahimov",
      "child_class": "9A"
    }
  ]
}
```

#### Get Statistics
```bash
curl http://your-server:8080/api/admin/stats
```

Response:
```json
{
  "total_users": 150,
  "total_complaints": 45,
  "pending_complaints": 12
}
```

### Downloading Complaint Files

Files are stored in Telegram's cloud. To download:

1. **From Telegram**: Click the file in the notification
2. **From API**: Use `telegram_file_id` with Telegram Bot API

### Managing Complaint Status

Currently through database only:
```sql
-- Mark as reviewed
UPDATE complaints SET status = 'reviewed' WHERE id = 123;

-- Archive old complaints
UPDATE complaints SET status = 'archived'
WHERE created_at < NOW() - INTERVAL '1 year';
```

*(Future update: Add status management buttons)*

### Best Practices for Admins

1. **Check notifications daily** for new complaints
2. **Download documents immediately** for record-keeping
3. **Update status** after reviewing each complaint
4. **Export data regularly** via API for backup
5. **Monitor statistics** to identify trends
6. **Respond to parents** directly via phone if needed

---

## Common Issues & Solutions

### For Users

#### "Invalid phone number format"
- ✅ Use format: `+998901234567`
- ✅ Include `+998` prefix
- ✅ Total 13 characters
- ❌ Don't use: `998...`, `+8...`, spaces, dashes

#### "Name contains invalid characters"
- ✅ Use only letters: `Akmal Rahimov`
- ❌ Don't use: numbers, @, +, _, etc.

#### "Invalid class format"
- ✅ Use format: `9A`, `11B`
- ❌ Don't use: `9-A`, `ninth`, `9 A`

#### "Complaint text too short"
- Minimum 10 characters required
- Be specific and clear
- Provide details

#### "Already registered"
- You can only register once
- Use `/start` to access main menu
- Contact admin if you need to change information

### For Admins

#### "Not receiving notifications"
- Check your phone number in server configuration
- Make sure you've started the bot at least once
- Verify bot has permission to send messages

#### "Cannot access admin panel"
- Verify your phone number matches configuration
- Register in the bot first (if not registered)
- Check server logs for errors

#### "API returns empty data"
- No users/complaints may exist yet
- Check database connection
- Verify server is running

---

## Tips & Tricks

### For Users

- **Save the bot link** for quick access
- **Screenshot important** complaint submissions
- **Check status regularly** in "My Complaints"
- **Be specific** in complaints for better resolution
- **Use proper language** (polite and professional)

### For Admins

- **Create a Telegram group** for all admins
- **Forward critical complaints** to the group
- **Set up notifications** on your phone
- **Export data monthly** for archives
- **Monitor response times** via statistics
- **Create templates** for common responses

---

## Security & Privacy

### Data Protection

- ✅ Phone numbers are validated and secured
- ✅ Only admins can view user data
- ✅ Files stored securely in Telegram cloud
- ✅ SQL injection prevention
- ✅ XSS protection

### What Admins Can See

- User's phone number
- Child's name and class
- Telegram username (if set)
- All complaint texts
- Submission dates

### What Other Users Can See

- **Nothing** - Users can only see their own data

---

## Support

### For Users

If you experience issues:
1. Try `/start` to reset
2. Contact school administration
3. Check this guide for solutions

### For Admins

If you experience technical issues:
1. Check server logs
2. Verify configuration
3. Review database status
4. Consult technical documentation

---

## Frequently Asked Questions

### Can I change my information after registration?
Currently, contact an admin to update your information. (Feature coming soon)

### Can I delete a complaint?
Yes, future updates will add this feature.

### How long are complaints stored?
Indefinitely, unless archived by admins.

### What file format are complaints?
DOCX (Microsoft Word format), compatible with all office software.

### Can I submit multiple complaints?
Yes, submit as many as needed.

### Do I get notifications when status changes?
Not yet, but this feature is planned.

### Is my data safe?
Yes, the bot implements security best practices and data protection measures.

---

**Last Updated**: October 2025
**Version**: 1.0.0
