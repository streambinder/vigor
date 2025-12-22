package database

import "time"

type countResult struct {
	Count int64
}

type avgResult struct {
	Avg float64
}

func GetUserCount() (int64, error) {
	if DB == nil {
		return 0, nil
	}
	var result countResult
	err := DB.Raw(`SELECT count(*) as count FROM users`).Scan(&result).Error
	return result.Count, err
}

func GetTrainingCount() (int64, error) {
	if DB == nil {
		return 0, nil
	}
	var result countResult
	err := DB.Raw(`SELECT count(*) as count FROM trainings`).Scan(&result).Error
	return result.Count, err
}

func GetAvgTrainingsPerDay() (float64, error) {
	if DB == nil {
		return 0, nil
	}
	var result avgResult
	err := DB.Raw(`
		SELECT COALESCE(AVG(daily_count), 0) as avg FROM (
			SELECT DATE(created_at) as day, COUNT(*) as daily_count
			FROM trainings
			WHERE created_at > NOW() - INTERVAL '30 days'
			GROUP BY day
		) daily_counts
	`).Scan(&result).Error
	return result.Avg, err
}

type Report struct {
	ID         string
	Content    string
	TrainingID *string
	ActivityID *string
	UserID     string
	UserEmail  string
	CreatedAt  time.Time
}

func GetReports() ([]Report, error) {
	if DB == nil {
		return nil, nil
	}
	var reports []Report
	err := DB.Raw(`
		SELECT r.id, r.content, r.training_id, r.activity_id, r.user_id, u.email as user_email, r.created_at
		FROM reports r
		JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
		LIMIT 100
	`).Scan(&reports).Error
	return reports, err
}
