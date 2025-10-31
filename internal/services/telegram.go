package services

import (
	"fmt"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramService handles Telegram file operations
type TelegramService struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramService creates a new Telegram service
func NewTelegramService(bot *tgbotapi.BotAPI) *TelegramService {
	return &TelegramService{bot: bot}
}

// UploadDocument uploads a document to Telegram and returns file_id
// This stores the file in Telegram's cloud, not on our server
func (s *TelegramService) UploadDocument(chatID int64, docPath, filename string) (fileID string, err error) {
	// Get file info for debugging
	fileInfo, err := os.Stat(docPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	fmt.Printf("[DEBUG] Uploading file: %s, size: %d bytes\n", filename, fileInfo.Size())

	// Open the file
	file, err := os.Open(docPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create document from file reader with proper filename
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   filename,
		Reader: file,
	})

	// Send document
	fmt.Printf("[DEBUG] Sending document to Telegram...\n")
	msg, err := s.bot.Send(doc)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	// Extract file_id from message
	if msg.Document != nil {
		fmt.Printf("[DEBUG] Document uploaded successfully, file_id: %s, file_size: %d\n",
			msg.Document.FileID, msg.Document.FileSize)
		return msg.Document.FileID, nil
	}

	return "", fmt.Errorf("no document in response")
}

// SendDocumentByFileID sends a document using file_id
func (s *TelegramService) SendDocumentByFileID(chatID int64, fileID, caption string) error {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
	doc.Caption = caption

	_, err := s.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}

	return nil
}

// SendMessage sends a text message
func (s *TelegramService) SendMessage(chatID int64, text string, replyMarkup interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	_, err := s.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// AnswerCallbackQuery answers a callback query
func (s *TelegramService) AnswerCallbackQuery(callbackQueryID string, text string) error {
	callback := tgbotapi.NewCallback(callbackQueryID, text)
	_, err := s.bot.Request(callback)
	if err != nil {
		return fmt.Errorf("failed to answer callback: %w", err)
	}

	return nil
}

// EditMessage edits an existing message
func (s *TelegramService) EditMessage(chatID int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	_, err := s.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (s *TelegramService) DeleteMessage(chatID int64, messageID int) error {
	msg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := s.bot.Request(msg)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// GetFile gets file info from Telegram
func (s *TelegramService) GetFile(fileID string) (*tgbotapi.File, error) {
	file, err := s.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &file, nil
}

// DownloadFile downloads a file from Telegram
func (s *TelegramService) DownloadFile(fileID, savePath string) error {
	file, err := s.GetFile(fileID)
	if err != nil {
		return err
	}

	link := file.Link(s.bot.Token)

	// Download file using http client
	client := &http.Client{}
	resp, err := client.Get(link)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Save to file
	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = out.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// NotifyAdmins sends a notification to all admins
func (s *TelegramService) NotifyAdmins(adminTelegramIDs []int64, message string) error {
	for _, adminID := range adminTelegramIDs {
		err := s.SendMessage(adminID, message, nil)
		if err != nil {
			// Log error but continue notifying other admins
			fmt.Printf("Failed to notify admin %d: %v\n", adminID, err)
		}
	}

	return nil
}

// SendDocumentToAdmins sends a document to all admins
func (s *TelegramService) SendDocumentToAdmins(adminTelegramIDs []int64, fileID, caption string) error {
	for _, adminID := range adminTelegramIDs {
		err := s.SendDocumentByFileID(adminID, fileID, caption)
		if err != nil {
			// Log error but continue sending to other admins
			fmt.Printf("Failed to send document to admin %d: %v\n", adminID, err)
		}
	}

	return nil
}
