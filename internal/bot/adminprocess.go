package bot

import (
	"sync"
	"time"
)

type AdminProcessType string

const (
	ProcessTypeSuspension   AdminProcessType = "suspension"
	ProcessTypeBan          AdminProcessType = "ban"
	ProcessTypeUnban        AdminProcessType = "unban"
	ProcessTypeAdmitToGreen AdminProcessType = "admit_to_green"
)

type AdminProcess struct {
	Type      AdminProcessType
	AdminID   int64
	Duration  string
	CreatedAt time.Time
}

type AdminProcessStore struct {
	processes map[int64]*AdminProcess
	mu        sync.RWMutex
}

func NewAdminProcessStore() *AdminProcessStore {
	return &AdminProcessStore{
		processes: make(map[int64]*AdminProcess),
	}
}

func (s *AdminProcessStore) Set(adminID int64, processType AdminProcessType, duration string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processes[adminID] = &AdminProcess{
		Type:      processType,
		AdminID:   adminID,
		Duration:  duration,
		CreatedAt: time.Now(),
	}
}

func (s *AdminProcessStore) Get(adminID int64) (*AdminProcess, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.processes[adminID]
	return p, exists
}

func (s *AdminProcessStore) Clear(adminID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.processes, adminID)
}

func (s *AdminProcessStore) CleanupExpired(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, p := range s.processes {
		if now.Sub(p.CreatedAt) > maxAge {
			delete(s.processes, id)
		}
	}
}
