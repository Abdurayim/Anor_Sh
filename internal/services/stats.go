package services

import (
	"time"

	"parent-bot/internal/models"
)

// BranchStats is the dashboard summary of one branch (or of all branches when BranchID is 0).
type BranchStats struct {
	BranchID          int
	Parents           int
	Students          int
	Teachers          int
	Classes           int
	Admins            int
	PresentToday      int
	AbsentToday       int
	Complaints        int
	PendingComplaints int
	Proposals         int
	PendingProposals  int
	AvgScore30d       float64 // average numeric grade over the last 30 days
	Scores30d         int
}

// AttendanceRate returns today's attendance in percent, or -1 if attendance was not taken.
func (b BranchStats) AttendanceRate() float64 {
	total := b.PresentToday + b.AbsentToday
	if total == 0 {
		return -1
	}
	return float64(b.PresentToday) / float64(total) * 100
}

// GetBranchStats collects dashboard numbers for a branch (branchID 0 = all branches).
// Errors of single counters are ignored so one failing query does not hide the dashboard.
func (s *BotService) GetBranchStats(branchID int) BranchStats {
	st := BranchStats{BranchID: branchID}
	st.Parents, _ = s.UserService.CountUsers(branchID)
	st.Students, _ = s.StudentService.CountStudents(branchID)
	st.Teachers, _ = s.TeacherService.CountTeachers(branchID)
	st.Classes, _ = s.ClassRepo.Count(branchID)
	st.Admins, _ = s.AdminRepo.Count(branchID)
	st.Complaints, _ = s.ComplaintService.CountComplaints(branchID)
	st.PendingComplaints, _ = s.ComplaintService.CountComplaintsByStatus(branchID, models.StatusPending)
	st.Proposals, _ = s.ProposalService.CountProposals(branchID)
	st.PendingProposals, _ = s.ProposalService.CountProposalsByStatus(branchID, models.StatusPending)

	if today, err := s.AttendanceService.GetTodayAttendanceAllClasses(branchID); err == nil {
		for _, a := range today {
			if a.Status == "present" {
				st.PresentToday++
			} else {
				st.AbsentToday++
			}
		}
	}

	since := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	st.AvgScore30d, st.Scores30d, _ = s.TestResultRepo.AverageNumericScore(branchID, since)
	return st
}
