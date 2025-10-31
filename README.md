# Parent Complaint Bot

Telegram bot for automating parent complaints in schools. Built with Go, Gin, and PostgreSQL.

## Features

- **Bilingual Support**: Uzbek (primary) and Russian languages
- **User Registration**: Phone number, child name, and class validation
- **Complaint Submission**: Text complaints converted to DOCX format
- **Telegram Cloud Storage**: Files stored in Telegram's cloud using file_id (no local storage)
- **Admin Panel**: View users, complaints, and statistics
- **Data Validation**: Comprehensive input validation to prevent injection attacks
- **Fast Database Queries**: Optimized with indexes for quick retrieval
- **Secure**: Phone number validation, SQL injection prevention, XSS protection

## Architecture

```
parent-bot/
├── cmd/bot/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection & migrations
│   ├── models/           # Data models
│   ├── handlers/         # Telegram update handlers
│   ├── api/              # Gin API routes
│   ├── middleware/       # Authentication & rate limiting
│   ├── validator/        # Input validation
│   ├── services/         # Business logic
│   ├── repository/       # Database queries
│   ├── i18n/             # Internationalization
│   ├── state/            # User state management
│   └── utils/            # Helper functions
├── pkg/docx/             # DOCX document generation
└── temp/                 # Temporary files (auto-cleaned)
```

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))

## Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd parent-bot
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Setup PostgreSQL

Create a database:

```sql
CREATE DATABASE parent_bot;
```

### 4. Configure environment

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```env
BOT_TOKEN=your_telegram_bot_token_here
WEBHOOK_URL=https://your-domain.com/webhook

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=parent_bot

SERVER_PORT=8080
GIN_MODE=release

# Admin phone numbers (max 3, comma-separated)
ADMIN_PHONES=+998901234567,+998907654321
```

### 5. Run migrations

Migrations run automatically on startup, or manually:

```bash
psql -U postgres -d parent_bot -f internal/database/migrations/001_initial.sql
```

### 6. Run the bot

```bash
go run cmd/bot/main.go
```

Or build and run:

```bash
go build -o parent-bot cmd/bot/main.go
./parent-bot
```

## Usage

### For Users

1. Start the bot: `/start`
2. Choose language (Uzbek/Russian)
3. Share phone number (+998XXXXXXXXX)
4. Enter child's name
5. Enter child's class (e.g., 9A, 11B)
6. Submit complaints via the main menu

### For Admins

Admins are identified by phone numbers configured in `.env`.

**Commands**:
- View all registered users
- View all complaints
- Download complaint documents
- View statistics

**API Endpoints**:
- `GET /api/admin/users` - List all users
- `GET /api/admin/complaints` - List all complaints
- `GET /api/admin/stats` - View statistics

## Validation Rules

### Phone Number
- Format: `+998XXXXXXXXX` (exactly 13 characters)
- Must start with +998
- Valid operator codes: 90, 91, 93, 94, 95, 97, 98, 99, 33, 88, 77

### Name
- Length: 2-100 characters
- Only letters (Latin, Cyrillic), spaces, hyphens, apostrophes
- No numbers or special characters (+, @, _, %, $, etc.)

### Class
- Format: `[1-11][A-Z]` (e.g., 9A, 11B)
- Grade: 1-11
- Letter: A-Z or А-Я

### Complaint Text
- Length: 10-5000 characters
- Sanitized to prevent SQL injection and XSS

## Database Schema

### Users
- `telegram_id` - Unique Telegram user ID (indexed)
- `phone_number` - Unique phone number (indexed)
- `child_name` - Child's full name
- `child_class` - Class (indexed)
- `language` - Preferred language (uz/ru)

### Complaints
- `user_id` - Foreign key to users (indexed)
- `complaint_text` - Complaint content
- `telegram_file_id` - File stored in Telegram cloud (indexed)
- `filename` - Document filename
- `status` - pending/reviewed/archived (indexed)

### Admins
- `phone_number` - Unique admin phone (indexed)
- `telegram_id` - Admin's Telegram ID (indexed)

## File Storage Strategy

Files are **NOT** stored on the server. Instead:

1. Generate DOCX → `temp/` folder
2. Upload to Telegram → Get `file_id`
3. Save `file_id` in database
4. Delete local temp file
5. Retrieve later using `file_id`

**Benefits**:
- Zero server storage costs
- Automatic file backup
- Fast retrieval
- No disk space concerns

## Security Features

- **Input Sanitization**: All user inputs sanitized
- **SQL Injection Prevention**: Parameterized queries
- **XSS Protection**: HTML escaping
- **Phone Validation**: Strict format checking
- **Rate Limiting**: Prevent spam (configurable)
- **Admin Authentication**: Phone-based verification

## Performance Optimizations

- **Database Indexes**: Fast queries on telegram_id, phone_number, class, status
- **Connection Pooling**: 25 max connections, 5 idle
- **In-Memory Caching**: Admin phones, user states
- **Prepared Statements**: Pre-compiled queries
- **View Optimization**: `v_complaints_with_user` for admin dashboard

## Deployment

### Using systemd

Create `/etc/systemd/system/parent-bot.service`:

```ini
[Unit]
Description=Parent Complaint Bot
After=network.target postgresql.service

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/parent-bot
ExecStart=/path/to/parent-bot/parent-bot
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable parent-bot
sudo systemctl start parent-bot
```

### Using Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o parent-bot cmd/bot/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/parent-bot .
COPY --from=builder /app/internal ./internal
CMD ["./parent-bot"]
```

Build and run:

```bash
docker build -t parent-bot .
docker run -d --env-file .env -p 8080:8080 parent-bot
```

## API Reference

### Health Check
```
GET /health
Response: {"status": "healthy"}
```

### Admin Endpoints

**List Users**
```
GET /api/admin/users
Response: {"users": [...]}
```

**List Complaints**
```
GET /api/admin/complaints
Response: {"complaints": [...]}
```

**Statistics**
```
GET /api/admin/stats
Response: {
  "total_users": 150,
  "total_complaints": 45,
  "pending_complaints": 12
}
```

## Troubleshooting

### Bot not responding
- Check bot token in `.env`
- Verify webhook is set correctly
- Check logs for errors

### Database connection failed
- Verify PostgreSQL is running
- Check database credentials in `.env`
- Ensure database exists

### File upload failed
- Check Telegram API limits
- Verify bot has document send permissions
- Check temp directory permissions

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

[MIT License](LICENSE)

## Support

For issues and questions, please open an issue on GitHub.

---

**Note**: This bot is designed for educational institutions. Ensure compliance with local data protection regulations when handling user data.
