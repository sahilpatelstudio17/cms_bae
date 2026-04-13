package services

import (
	"cms/internal/repositories"
	"time"
)

type DashboardService struct {
	employeeRepo   *repositories.EmployeeRepository
	taskRepo       *repositories.TaskRepository
	attendanceRepo *repositories.AttendanceRepository
}

func NewDashboardService(
	employeeRepo *repositories.EmployeeRepository,
	taskRepo *repositories.TaskRepository,
	attendanceRepo *repositories.AttendanceRepository,
) *DashboardService {
	return &DashboardService{
		employeeRepo:   employeeRepo,
		taskRepo:       taskRepo,
		attendanceRepo: attendanceRepo,
	}
}

type DashboardStats struct {
	TotalEmployees    int64                   `json:"total_employees"`
	OpenTasks         int64                   `json:"open_tasks"`
	AttendanceToday   int64                   `json:"attendance_today"`
	AttendanceSummary []AttendanceSummaryItem `json:"attendance_summary"`
}

type AttendanceSummaryItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
	Color string `json:"color"`
}

func (s *DashboardService) GetDashboardStats(companyID uint) (*DashboardStats, error) {
	// Get total employees for company
	employees, total, err := s.employeeRepo.List(companyID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Get open tasks for company
	tasks, _, err := s.taskRepo.List(companyID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Count open tasks (status = "open" or "pending")
	var openTasks int64
	for _, task := range tasks {
		if task.Status == "open" || task.Status == "pending" {
			openTasks++
		}
	}

	// Get today's attendance for company
	today := time.Now().Format("2006-01-02")
	attendance, _, err := s.attendanceRepo.FindByDateAndCompany(companyID, today, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Calculate attendance summary
	var present, absent, onLeave int64
	employeeMap := make(map[uint]bool)

	for _, emp := range employees {
		employeeMap[emp.ID] = true
	}

	for _, record := range attendance {
		if record.InTime != nil {
			// Employee has checked in
			present++
		} else {
			// Employee has a record but hasn't checked in
			absent++
		}
		delete(employeeMap, record.EmployeeID)
	}

	// Mark unmarked employees (no record for today) as absent
	absent += int64(len(employeeMap))

	// Calculate attendance percentage
	totalEmp := total
	if totalEmp == 0 {
		totalEmp = 1 // Avoid division by zero
	}
	attendancePercent := (present * 100) / totalEmp

	stats := &DashboardStats{
		TotalEmployees:  total,
		OpenTasks:       openTasks,
		AttendanceToday: attendancePercent,
		AttendanceSummary: []AttendanceSummaryItem{
			{
				Label: "Present",
				Value: present,
				Color: "bg-emerald-500",
			},
			{
				Label: "Absent",
				Value: absent,
				Color: "bg-rose-500",
			},
			{
				Label: "On Leave",
				Value: onLeave,
				Color: "bg-amber-500",
			},
		},
	}

	return stats, nil
}
