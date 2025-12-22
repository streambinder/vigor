package database

import "github.com/streambinder/vigor/model"

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

func GetReports() ([]model.Report, error) {
	if DB == nil {
		return nil, nil
	}
	var reports []model.Report
	err := DB.Order("created_at DESC").Limit(100).
		Preload("User").
		Find(&reports).Error
	return reports, err
}

func GetTrainings() ([]model.Training, error) {
	if DB == nil {
		return nil, nil
	}
	var trainings []model.Training
	err := DB.Order("created_at DESC").Limit(100).
		Preload("User").
		Preload("Routines.Blocks.Activities").
		Find(&trainings).Error
	return trainings, err
}
