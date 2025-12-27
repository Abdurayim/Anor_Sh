package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"parent-bot/internal/models"
)

// Manager manages user conversation states
type Manager struct {
	db    *sql.DB
	cache map[int64]*models.UserState // In-memory cache for faster access
	mu    sync.RWMutex
}

// NewManager creates a new state manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:    db,
		cache: make(map[int64]*models.UserState),
	}
}

// Set sets user state
func (m *Manager) Set(telegramID int64, state string, data *models.StateData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal state data: %w", err)
	}

	query := `
		INSERT INTO user_states (telegram_id, state, data)
		VALUES (?, ?, ?)
		ON CONFLICT (telegram_id)
		DO UPDATE SET state = excluded.state, data = excluded.data, updated_at = CURRENT_TIMESTAMP
	`

	_, err = m.db.Exec(query, telegramID, state, dataJSON)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	// Update cache
	m.cache[telegramID] = &models.UserState{
		TelegramID: telegramID,
		State:      state,
		Data:       dataJSON,
	}

	return nil
}

// Get gets user state
func (m *Manager) Get(telegramID int64) (*models.UserState, error) {
	m.mu.RLock()

	// Check cache first
	if state, ok := m.cache[telegramID]; ok {
		m.mu.RUnlock()
		return state, nil
	}
	m.mu.RUnlock()

	// Query database
	query := `
		SELECT telegram_id, state, data, updated_at
		FROM user_states
		WHERE telegram_id = ?
	`

	var state models.UserState
	err := m.db.QueryRow(query, telegramID).Scan(
		&state.TelegramID,
		&state.State,
		&state.Data,
		&state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Update cache
	m.mu.Lock()
	m.cache[telegramID] = &state
	m.mu.Unlock()

	return &state, nil
}

// GetData gets and parses state data
func (m *Manager) GetData(telegramID int64) (*models.StateData, error) {
	state, err := m.Get(telegramID)
	if err != nil {
		return nil, err
	}

	if state == nil || state.Data == nil {
		return &models.StateData{}, nil
	}

	var data models.StateData
	err = json.Unmarshal(state.Data, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state data: %w", err)
	}

	return &data, nil
}

// Delete deletes user state
func (m *Manager) Delete(telegramID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	query := `DELETE FROM user_states WHERE telegram_id = ?`
	_, err := m.db.Exec(query, telegramID)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	// Remove from cache
	delete(m.cache, telegramID)

	return nil
}

// Clear clears user state (sets to registered)
func (m *Manager) Clear(telegramID int64) error {
	return m.Set(telegramID, models.StateRegistered, &models.StateData{})
}

// CleanOldStates removes states older than specified hours
func (m *Manager) CleanOldStates(hours int) error {
	// SQLite uses datetime() function, not INTERVAL
	query := `
		DELETE FROM user_states
		WHERE updated_at < datetime('now', '-' || ? || ' hours')
	`

	_, err := m.db.Exec(query, hours)
	if err != nil {
		return fmt.Errorf("failed to clean old states: %w", err)
	}

	// Clear cache
	m.mu.Lock()
	m.cache = make(map[int64]*models.UserState)
	m.mu.Unlock()

	return nil
}

// GetState returns just the state string
func (m *Manager) GetState(telegramID int64) (string, error) {
	state, err := m.Get(telegramID)
	if err != nil {
		return "", err
	}

	if state == nil {
		return models.StateStart, nil
	}

	return state.State, nil
}
