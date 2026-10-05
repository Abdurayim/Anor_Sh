package services

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/config"
	"parent-bot/internal/models"
	"parent-bot/internal/repository"
	"parent-bot/internal/state"
)

// BotService is the main bot service
type BotService struct {
	Bot                 *tgbotapi.BotAPI
	Config              *config.Config
	UserRepo            *repository.UserRepository
	ComplaintRepo       *repository.ComplaintRepository
	ProposalRepo        *repository.ProposalRepository
	TimetableRepo       *repository.TimetableRepository
	AnnouncementRepo    *repository.AnnouncementRepository
	AdminRepo           *repository.AdminRepository
	BranchRepo          *repository.BranchRepository
	ClassRepo           *repository.ClassRepository
	TeacherRepo         *repository.TeacherRepository
	StudentRepo         *repository.StudentRepository
	TestResultRepo      *repository.TestResultRepository
	AttendanceRepo      *repository.AttendanceRepository
	StateManager        *state.Manager
	TelegramService     *TelegramService
	UserService         *UserService
	ComplaintService    *ComplaintService
	ProposalService     *ProposalService
	TimetableService    *TimetableService
	AnnouncementService *AnnouncementService
	DocumentService     *DocumentService
	TeacherService      *TeacherService
	StudentService      *StudentService
	TestResultService   *TestResultService
	AttendanceService   *AttendanceService
}

// NewBotService creates a new bot service
func NewBotService(cfg *config.Config, db *sql.DB) (*BotService, error) {
	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return NewBotServiceWithBot(cfg, db, bot), nil
}

// NewBotServiceWithBot wires repositories and services around an existing Telegram client
// (bot may be nil in tests that never talk to Telegram).
func NewBotServiceWithBot(cfg *config.Config, db *sql.DB, bot *tgbotapi.BotAPI) *BotService {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	proposalRepo := repository.NewProposalRepository(db)
	timetableRepo := repository.NewTimetableRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	branchRepo := repository.NewBranchRepository(db)
	classRepo := repository.NewClassRepository(db)
	teacherRepo := repository.NewTeacherRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	testResultRepo := repository.NewTestResultRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)

	// Initialize state manager
	stateManager := state.NewManager(db)

	// Initialize services
	telegramService := NewTelegramService(bot)
	userService := NewUserService(userRepo)
	complaintService := NewComplaintService(complaintRepo, userRepo)
	proposalService := NewProposalService(proposalRepo, userRepo)
	timetableService := NewTimetableService(timetableRepo, classRepo)
	announcementService := NewAnnouncementService(announcementRepo)
	documentService := NewDocumentService("./temp_docs") // temp directory for generated documents
	teacherService := NewTeacherService(db)
	studentService := NewStudentService(db)
	testResultService := NewTestResultService(db)
	attendanceService := NewAttendanceService(db)

	return &BotService{
		Bot:                 bot,
		Config:              cfg,
		UserRepo:            userRepo,
		ComplaintRepo:       complaintRepo,
		ProposalRepo:        proposalRepo,
		TimetableRepo:       timetableRepo,
		AnnouncementRepo:    announcementRepo,
		AdminRepo:           adminRepo,
		BranchRepo:          branchRepo,
		ClassRepo:           classRepo,
		TeacherRepo:         teacherRepo,
		StudentRepo:         studentRepo,
		TestResultRepo:      testResultRepo,
		AttendanceRepo:      attendanceRepo,
		StateManager:        stateManager,
		TelegramService:     telegramService,
		UserService:         userService,
		ComplaintService:    complaintService,
		ProposalService:     proposalService,
		TimetableService:    timetableService,
		AnnouncementService: announcementService,
		DocumentService:     documentService,
		TeacherService:      teacherService,
		StudentService:      studentService,
		TestResultService:   testResultService,
		AttendanceService:   attendanceService,
	}
}

// SetWebhook registers the webhook URL; Telegram will send secret in
// X-Telegram-Bot-Api-Secret-Token on every request when it is not empty.
func (s *BotService) SetWebhook(webhookURL, secret string) error {
	// telegram-bot-api v5.5.1 does not know secret_token, so the request is built by hand.
	params := tgbotapi.Params{"url": webhookURL}
	if secret != "" {
		params["secret_token"] = secret
	}
	if _, err := s.Bot.MakeRequest("setWebhook", params); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return nil
}

// RemoveWebhook removes webhook (for polling mode)
func (s *BotService) RemoveWebhook() error {
	_, err := s.Bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	return nil
}

// InitializeAdmins syncs the admins table with .env: every configured phone becomes an
// active admin (the super admin or a branch admin), .env admins no longer configured are
// deactivated. Branch admins added by the super admin in the bot are kept.
func (s *BotService) InitializeAdmins() error {
	configured := s.Config.Admin.AllPhones()

	if phone := s.Config.Admin.SuperAdminPhone; phone != "" {
		if err := s.AdminRepo.UpsertFromEnv(phone, "Super admin", models.RoleSuperAdmin, 0); err != nil {
			return err
		}
		configured = append(configured, phone)
	}

	for _, code := range config.BranchCodes {
		phones := s.Config.Admin.BranchPhones[code]
		if len(phones) == 0 {
			continue
		}
		branch, err := s.BranchRepo.GetByCode(code)
		if err != nil {
			return err
		}
		if branch == nil {
			return fmt.Errorf("branch %q not found in database", code)
		}
		for _, phone := range phones {
			if err := s.AdminRepo.UpsertFromEnv(phone, "Admin "+branch.NameUz, models.RoleAdmin, branch.ID); err != nil {
				return err
			}
		}
	}

	return s.AdminRepo.DeactivateEnvAdminsExcept(configured)
}

// MaxAdminsPerBranch limits how many admins a branch can have.
const MaxAdminsPerBranch = 3

// AddBranchAdmin lets the super admin add an admin to a branch. The new admin links their
// Telegram account by sharing their contact in the bot.
func (s *BotService) AddBranchAdmin(superAdmin *models.Admin, phone, name string, branchID int) error {
	if superAdmin == nil || !superAdmin.IsSuperAdmin() {
		return fmt.Errorf("only the super admin can add admins")
	}
	branch, err := s.BranchRepo.GetByID(branchID)
	if err != nil || branch == nil {
		return fmt.Errorf("filial topilmadi / филиал не найден")
	}
	if existing, _ := s.AdminRepo.GetByPhoneNumber(phone); existing != nil {
		return fmt.Errorf("bu raqam allaqachon admin / этот номер уже администратор")
	}
	if teacher, _ := s.TeacherRepo.GetByPhoneNumber(phone); teacher != nil && teacher.IsActive {
		return fmt.Errorf("bu raqam o'qituvchiga tegishli / этот номер принадлежит учителю")
	}
	if n, err := s.AdminRepo.Count(branchID); err == nil && n >= MaxAdminsPerBranch {
		return fmt.Errorf("filialda %d tadan ortiq admin bo'lmaydi / в филиале не больше %d админов", MaxAdminsPerBranch, MaxAdminsPerBranch)
	}
	if err := s.AdminRepo.CreateBranchAdmin(phone, name, branchID, superAdmin.ID); err != nil {
		if err == repository.ErrAdminExists {
			return fmt.Errorf("bu raqam allaqachon admin / этот номер уже администратор")
		}
		return err
	}
	return nil
}

// RemoveBranchAdmin lets the super admin remove a branch admin. Admins configured in .env
// cannot be removed here (they would come back on the next start).
func (s *BotService) RemoveBranchAdmin(superAdmin *models.Admin, adminID int) error {
	if superAdmin == nil || !superAdmin.IsSuperAdmin() {
		return fmt.Errorf("only the super admin can remove admins")
	}
	admin, err := s.AdminRepo.GetByID(adminID)
	if err != nil || admin == nil || admin.IsSuperAdmin() {
		return fmt.Errorf("admin topilmadi / админ не найден")
	}
	if admin.Source == models.AdminSourceEnv {
		return fmt.Errorf(".env dagi adminni faqat .env dan o'chirish mumkin / админ из .env удаляется только в .env")
	}
	return s.AdminRepo.Deactivate(adminID)
}

// GetAdminTelegramIDs returns Telegram IDs of linked admins of a branch (branchID 0 = all branches)
func (s *BotService) GetAdminTelegramIDs(branchID int) ([]int64, error) {
	admins, err := s.AdminRepo.GetAll(branchID)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, admin := range admins {
		if admin.TelegramID != nil {
			ids = append(ids, *admin.TelegramID)
		}
	}

	return ids, nil
}

// GetAdmin returns the active *branch* admin linked to this Telegram account, or nil.
// An account is linked only after the admin shared their own contact (see LinkAdminByContact),
// so typing someone else's phone number never grants admin rights.
func (s *BotService) GetAdmin(telegramID int64) *models.Admin {
	admin := s.getLinkedAdmin(telegramID)
	if admin == nil || admin.IsSuperAdmin() {
		return nil
	}
	return admin
}

// GetSuperAdmin returns the super admin linked to this Telegram account, or nil.
func (s *BotService) GetSuperAdmin(telegramID int64) *models.Admin {
	admin := s.getLinkedAdmin(telegramID)
	if admin == nil || !admin.IsSuperAdmin() {
		return nil
	}
	return admin
}

func (s *BotService) getLinkedAdmin(telegramID int64) *models.Admin {
	if telegramID == 0 {
		return nil
	}
	admin, err := s.AdminRepo.GetByTelegramID(telegramID)
	if err != nil {
		return nil
	}
	return admin
}

// GetLinkedAdmin returns any active admin (branch admin or super admin) linked to this Telegram account, or nil.
func (s *BotService) GetLinkedAdmin(telegramID int64) *models.Admin {
	return s.getLinkedAdmin(telegramID)
}

// IsAdmin reports whether the Telegram account is a linked, active admin (branch admin or super admin).
// The phone number argument is ignored: phone numbers are not proof of identity.
func (s *BotService) IsAdmin(_ string, telegramID int64) (bool, error) {
	return s.getLinkedAdmin(telegramID) != nil, nil
}

// LinkAdminByContact links the Telegram account to an admin (branch or super) if the
// verified phone belongs to an active admin. It returns the admin or nil.
func (s *BotService) LinkAdminByContact(verifiedPhone string, telegramID int64) (*models.Admin, error) {
	admin, err := s.AdminRepo.GetByPhoneNumber(verifiedPhone)
	if err != nil || admin == nil {
		return nil, err
	}
	if err := s.AdminRepo.UpdateTelegramID(verifiedPhone, telegramID); err != nil {
		return nil, err
	}
	return s.AdminRepo.GetByTelegramID(telegramID)
}

// BranchName returns the localized branch name, or "" if unknown.
func (s *BotService) BranchName(branchID int, lang string) string {
	branch, err := s.BranchRepo.GetByID(branchID)
	if err != nil || branch == nil {
		return ""
	}
	return branch.Name(lang)
}

// ClassInBranch reports whether a class belongs to the branch.
func (s *BotService) ClassInBranch(classID, branchID int) bool {
	ok, err := s.ClassRepo.BelongsToBranch(classID, branchID)
	return err == nil && ok
}

// StudentInBranch reports whether a student's class belongs to the branch.
func (s *BotService) StudentInBranch(studentID, branchID int) bool {
	student, err := s.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		return false
	}
	return s.ClassInBranch(student.ClassID, branchID)
}
