# Extending the Parent Complaint Bot

This guide shows you how to add new features and handlers to the bot.

## Architecture Overview

The bot uses a handler-based architecture where:
1. **Updates** come from Telegram (via webhook or polling)
2. **Router** (`handlers/webhook.go`) parses and routes updates
3. **State Manager** tracks conversation states
4. **Handlers** process messages and callbacks
5. **Services** handle business logic
6. **Repositories** interact with database

## Adding a New Command

### Example: Adding `/stats` command for users

#### Step 1: Add Translation Keys

Edit `internal/i18n/uzbek.go`:
```go
const (
    MsgUserStats = "user_stats"
    // ... other constants
)

var uzbek = map[string]string{
    MsgUserStats: "📊 Sizning statistikangiz:\n\n" +
                  "Jami shikoyatlar: %d\n" +
                  "Ko'rilgan: %d\n" +
                  "Kutilmoqda: %d",
    // ... other translations
}
```

Edit `internal/i18n/russian.go`:
```go
var russian = map[string]string{
    MsgUserStats: "📊 Ваша статистика:\n\n" +
                  "Всего жалоб: %d\n" +
                  "Рассмотрено: %d\n" +
                  "Ожидание: %d",
    // ... other translations
}
```

#### Step 2: Create Handler

Create `internal/handlers/stats.go`:
```go
package handlers

import (
    "fmt"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "parent-bot/internal/i18n"
    "parent-bot/internal/models"
    "parent-bot/internal/services"
)

// HandleStatsCommand shows user statistics
func HandleStatsCommand(botService *services.BotService, message *tgbotapi.Message) error {
    telegramID := message.From.ID
    chatID := message.Chat.ID

    // Get user
    user, err := botService.UserService.GetUserByTelegramID(telegramID)
    if err != nil {
        return err
    }

    if user == nil {
        lang := i18n.LanguageUzbek
        text := i18n.Get(i18n.ErrNotRegistered, lang)
        return botService.TelegramService.SendMessage(chatID, text, nil)
    }

    lang := i18n.GetLanguage(user.Language)

    // Get statistics
    totalComplaints, _ := botService.ComplaintService.CountUserComplaints(user.ID)

    pendingCount := 0
    reviewedCount := 0

    complaints, _ := botService.ComplaintService.GetUserComplaints(user.ID, 1000, 0)
    for _, c := range complaints {
        if c.Status == models.StatusPending {
            pendingCount++
        } else if c.Status == models.StatusReviewed {
            reviewedCount++
        }
    }

    // Format message
    text := fmt.Sprintf(
        i18n.Get(i18n.MsgUserStats, lang),
        totalComplaints,
        reviewedCount,
        pendingCount,
    )

    return botService.TelegramService.SendMessage(chatID, text, nil)
}
```

#### Step 3: Register Command

Edit `internal/handlers/webhook.go`:
```go
func HandleCommand(botService *services.BotService, message *tgbotapi.Message) error {
    switch message.Command() {
    case "start":
        return HandleStart(botService, message)
    case "help":
        return HandleHelp(botService, message)
    case "complaint":
        return HandleComplaintCommand(botService, message)
    case "stats":  // NEW
        return HandleStatsCommand(botService, message)
    case "admin":
        return HandleAdminCommand(botService, message)
    default:
        return HandleStart(botService, message)
    }
}
```

#### Step 4: Test

```bash
go build -o parent-bot cmd/bot/main.go
./parent-bot

# In Telegram, send: /stats
```

## Adding a New Callback Button

### Example: Adding "Delete Complaint" button

#### Step 1: Add Button to Keyboard

Edit `internal/utils/keyboard.go`:
```go
// MakeComplaintActionsKeyboard creates actions for a complaint
func MakeComplaintActionsKeyboard(complaintID int, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData(
                "🗑 O'chirish / Удалить",
                fmt.Sprintf("delete_complaint_%d", complaintID),
            ),
            tgbotapi.NewInlineKeyboardButtonData(
                "◀️ Orqaga / Назад",
                "back_to_menu",
            ),
        ),
    )
}
```

#### Step 2: Handle Callback

Edit `internal/handlers/router.go`:
```go
func HandleCallbackQuery(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
    data := callback.Data

    // ... existing callbacks

    // Delete complaint
    if strings.HasPrefix(data, "delete_complaint_") {
        return HandleDeleteComplaint(botService, callback)
    }

    return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Unknown action")
}
```

#### Step 3: Implement Handler

Add to `internal/handlers/complaint.go`:
```go
// HandleDeleteComplaint handles complaint deletion
func HandleDeleteComplaint(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
    // Parse complaint ID from callback data
    parts := strings.Split(callback.Data, "_")
    if len(parts) != 3 {
        return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Invalid data")
    }

    complaintID, err := strconv.Atoi(parts[2])
    if err != nil {
        return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Invalid ID")
    }

    // Verify ownership
    user, _ := botService.UserService.GetUserByTelegramID(callback.From.ID)
    complaint, _ := botService.ComplaintService.GetComplaintByID(complaintID)

    if complaint == nil || complaint.UserID != user.ID {
        return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Not found")
    }

    // Delete (or mark as archived)
    _ = botService.ComplaintService.UpdateComplaintStatus(complaintID, models.StatusArchived)

    lang := i18n.GetLanguage(user.Language)
    _ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "Deleted")

    text := "✅ Shikoyat o'chirildi / Жалоба удалена"
    return botService.TelegramService.SendMessage(callback.Message.Chat.ID, text, utils.MakeMainMenuKeyboard(lang))
}
```

## Adding Multi-Step Conversations

### Example: Edit Child Information

#### Step 1: Add States

Edit `internal/models/state.go`:
```go
const (
    // ... existing states
    StateEditingChildName  = "editing_child_name"
    StateEditingChildClass = "editing_child_class"
)
```

#### Step 2: Create Handler

```go
// HandleEditInfoCommand initiates info editing
func HandleEditInfoCommand(botService *services.BotService, message *tgbotapi.Message) error {
    user, _ := botService.UserService.GetUserByTelegramID(message.From.ID)
    lang := i18n.GetLanguage(user.Language)

    // Ask what to edit
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Ism / Имя", "edit_name"),
            tgbotapi.NewInlineKeyboardButtonData("Sinf / Класс", "edit_class"),
        ),
    )

    text := "Nimani o'zgartirmoqchisiz? / Что хотите изменить?"
    return botService.TelegramService.SendMessage(message.Chat.ID, text, keyboard)
}

// In callback handler
if data == "edit_name" {
    stateData := &models.StateData{Language: user.Language}
    _ = botService.StateManager.Set(callback.From.ID, models.StateEditingChildName, stateData)

    text := "Yangi ismni kiriting / Введите новое имя:"
    return botService.TelegramService.SendMessage(callback.Message.Chat.ID, text, nil)
}

// In router
case models.StateEditingChildName:
    return HandleEditChildName(botService, message, stateData)
```

## Adding New Service Methods

### Example: Search Complaints by Keyword

#### Step 1: Add Repository Method

Edit `internal/repository/complaint_repo.go`:
```go
// SearchComplaintsByText searches complaints by text
func (r *ComplaintRepository) SearchComplaintsByText(searchText string, limit, offset int) ([]*models.Complaint, error) {
    query := `
        SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status
        FROM complaints
        WHERE complaint_text ILIKE $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := r.db.Query(query, "%"+searchText+"%", limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to search complaints: %w", err)
    }
    defer rows.Close()

    var complaints []*models.Complaint
    for rows.Next() {
        var c models.Complaint
        err := rows.Scan(&c.ID, &c.UserID, &c.ComplaintText, &c.TelegramFileID, &c.Filename, &c.CreatedAt, &c.Status)
        if err != nil {
            return nil, err
        }
        complaints = append(complaints, &c)
    }

    return complaints, nil
}
```

#### Step 2: Add Service Method

Edit `internal/services/complaint.go`:
```go
// SearchComplaints searches complaints by keyword
func (s *ComplaintService) SearchComplaints(keyword string, limit, offset int) ([]*models.Complaint, error) {
    complaints, err := s.repo.SearchComplaintsByText(keyword, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to search complaints: %w", err)
    }
    return complaints, nil
}
```

#### Step 3: Use in Handler

```go
// HandleSearchCommand searches complaints
func HandleSearchCommand(botService *services.BotService, message *tgbotapi.Message) error {
    // Get search query from message text
    keyword := strings.TrimPrefix(message.Text, "/search ")

    if keyword == "" {
        return botService.TelegramService.SendMessage(
            message.Chat.ID,
            "Qidiruv so'zini kiriting / Введите поисковый запрос",
            nil,
        )
    }

    // Search
    complaints, _ := botService.ComplaintService.SearchComplaints(keyword, 10, 0)

    // Format results
    text := fmt.Sprintf("🔍 Topildi: %d\n\n", len(complaints))
    for i, c := range complaints {
        preview := utils.TruncateText(c.ComplaintText, 50)
        text += fmt.Sprintf("%d. %s\n", i+1, preview)
    }

    return botService.TelegramService.SendMessage(message.Chat.ID, text, nil)
}
```

## Common Patterns

### 1. Protected Admin-Only Commands

```go
func HandleAdminOnlyCommand(botService *services.BotService, message *tgbotapi.Message) error {
    user, _ := botService.UserService.GetUserByTelegramID(message.From.ID)
    isAdmin, _ := botService.IsAdmin(user.PhoneNumber, message.From.ID)

    if !isAdmin {
        return botService.TelegramService.SendMessage(
            message.Chat.ID,
            "❌ Access denied",
            nil,
        )
    }

    // Admin logic here
}
```

### 2. Confirmation Pattern

```go
// Step 1: Ask for confirmation
keyboard := tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "confirm_action_123"),
        tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel_action"),
    ),
)

// Step 2: Handle callback
if strings.HasPrefix(data, "confirm_action_") {
    // Extract ID and process
    id := strings.TrimPrefix(data, "confirm_action_")
    // Do action
}
```

### 3. Pagination Pattern

```go
func MakePaginationKeyboard(page, totalPages int) tgbotapi.InlineKeyboardMarkup {
    var buttons []tgbotapi.InlineKeyboardButton

    if page > 1 {
        buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("◀️ Prev", fmt.Sprintf("page_%d", page-1)))
    }

    buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", page, totalPages), "ignore"))

    if page < totalPages {
        buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Next ▶️", fmt.Sprintf("page_%d", page+1)))
    }

    return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
}
```

## Testing Your Changes

1. **Build**: `go build -o parent-bot cmd/bot/main.go`
2. **Run**: `./parent-bot` (polling mode)
3. **Test in Telegram**: Send messages to your bot
4. **Check logs**: Monitor console output
5. **Debug**: Add `log.Printf()` statements

## Best Practices

1. **Always validate user input** using validators
2. **Use transactions** for multi-step database operations
3. **Handle errors gracefully** and show user-friendly messages
4. **Log errors** for debugging
5. **Keep handlers thin** - move logic to services
6. **Use constants** for magic strings
7. **Add translations** for all user-facing text
8. **Test edge cases** (empty input, special characters, etc.)

## Need Help?

- Check existing handlers for examples
- Review service layer for available methods
- Look at repository layer for database queries
- Refer to Telegram Bot API docs for new features

Happy extending! 🚀
