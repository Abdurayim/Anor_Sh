# Quick Start Guide

## Prerequisites

Before you begin, ensure you have:

1. **Go 1.21+** installed
2. **PostgreSQL 14+** running
3. **Telegram Bot Token** from [@BotFather](https://t.me/BotFather)

## Step-by-Step Setup

### 1. Create Your Bot

Talk to [@BotFather](https://t.me/BotFather) on Telegram:

```
/newbot
# Follow prompts to name your bot
# Save the bot token you receive
```

### 2. Setup PostgreSQL Database

```bash
# Login to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE parent_bot;

# Exit psql
\q
```

### 3. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your favorite editor
nano .env
```

Minimum required configuration in `.env`:

```env
# Your bot token from BotFather
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Database credentials
DB_PASSWORD=your_postgres_password

# Admin phone numbers (max 3)
ADMIN_PHONES=+998901234567,+998907654321
```

### 4. Run the Bot

```bash
# Make sure PostgreSQL is running
sudo systemctl status postgresql

# Run the bot
go run cmd/bot/main.go
```

You should see:

```
✓ Connected to database
✓ Database migrations completed
✓ Bot authorized: @YourBotName
✓ Admins initialized
🚀 Server starting on :8080
📱 Bot is ready to receive messages!
```

### 5. Test Your Bot

Open Telegram and:

1. Find your bot by username
2. Send `/start`
3. Choose language (Uzbek/Russian)
4. Share your phone number
5. Complete registration

## Common Issues

### "Bot token is invalid"

- Check your `BOT_TOKEN` in `.env`
- Make sure there are no extra spaces
- Verify the token in BotFather

### "Failed to connect to database"

- Check PostgreSQL is running: `sudo systemctl status postgresql`
- Verify database credentials in `.env`
- Ensure `parent_bot` database exists

### "Admin phone validation failed"

- Phone format must be: `+998XXXXXXXXX`
- Use valid Uzbek operator codes: 90, 91, 93, 94, 95, 97, 98, 99, 33, 88, 77
- Example: `+998901234567`

## Next Steps

1. **Production Deployment**: See `README.md` for systemd or Docker setup
2. **Webhook Setup**: For production, set `WEBHOOK_URL` in `.env`
3. **Admin Features**: Admins can access `/api/admin/*` endpoints
4. **Monitoring**: Check `/health` endpoint for service status

## Testing Complaint Submission

After registration:

1. Tap "✍️ Shikoyat yuborish" (Submit complaint)
2. Write your complaint (min 10 characters)
3. Confirm submission
4. Document is generated and sent to admins

Admins will receive:
- Notification with complaint details
- Downloadable DOCX file with full complaint

## API Endpoints

Access admin APIs (no authentication yet, add in production):

```bash
# List all users
curl http://localhost:8080/api/admin/users

# List all complaints
curl http://localhost:8080/api/admin/complaints

# Get statistics
curl http://localhost:8080/api/admin/stats
```

## Development Mode

For development with hot reload:

```bash
# Install air (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## Production Checklist

Before deploying to production:

- [ ] Set `GIN_MODE=release` in `.env`
- [ ] Configure `WEBHOOK_URL` for your domain
- [ ] Setup HTTPS with Let's Encrypt
- [ ] Add authentication to admin API endpoints
- [ ] Configure firewall rules
- [ ] Setup database backups
- [ ] Enable PostgreSQL SSL mode
- [ ] Monitor logs and errors
- [ ] Test all workflows thoroughly

## Support

If you encounter issues:

1. Check logs in the terminal
2. Verify all environment variables
3. Test database connection manually
4. Review `README.md` for detailed documentation

Happy bot building! 🚀
