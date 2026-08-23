package database

import (
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

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

func GetCompletedTrainingCount() (int64, error) {
	if DB == nil {
		return 0, nil
	}
	var result countResult
	err := DB.Raw(`SELECT count(*) as count FROM trainings WHERE completed_at IS NOT NULL`).Scan(&result).Error
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

func GetAvgCompletedTrainingsPerDay() (float64, error) {
	if DB == nil {
		return 0, nil
	}
	var result avgResult
	err := DB.Raw(`
		SELECT COALESCE(AVG(daily_count), 0) as avg FROM (
			SELECT DATE(completed_at) as day, COUNT(*) as daily_count
			FROM trainings
			WHERE completed_at IS NOT NULL AND completed_at > NOW() - INTERVAL '30 days'
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
		Preload("Training.Routines.Blocks.Activities").
		Preload("Training.LLMSteps", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Activity").
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
		Preload("LLMSteps", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Find(&trainings).Error
	return trainings, err
}

func GetUsers() ([]model.User, error) {
	if DB == nil {
		return nil, nil
	}
	var users []model.User
	err := DB.Order("created_at DESC").Limit(100).
		Preload("Profile").
		Find(&users).Error
	return users, err
}

type ModelQualityPoint struct {
	Model string
	Count int64
}

func GetBadQualityPerModel() ([]ModelQualityPoint, error) {
	if DB == nil {
		return nil, nil
	}
	var results []ModelQualityPoint
	err := DB.Raw(`
		SELECT s.model AS model, COUNT(DISTINCT f.training_id) AS count
		FROM training_feedbacks f
		JOIN llm_steps s ON s.training_id = f.training_id AND s.position = 0
		WHERE f.quality = false
		  AND s.model IS NOT NULL
		  AND s.model != ''
		GROUP BY s.model
		ORDER BY count DESC
	`).Scan(&results).Error
	return results, err
}

func DeleteReport(id string) error {
	if DB == nil {
		return nil
	}
	return DB.Delete(&model.Report{}, "id = ?", id).Error
}

// GetUserNames resolves user UUIDs to display names (first_name) from profiles
func GetUserNames(userIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if DB == nil || len(userIDs) == 0 {
		return result, nil
	}

	var profiles []struct {
		UserID    string `gorm:"column:user_id"`
		FirstName string `gorm:"column:first_name"`
	}
	err := DB.Raw(`SELECT user_id::text, first_name FROM profiles WHERE user_id::text IN ?`, userIDs).Scan(&profiles).Error
	if err != nil {
		return result, err
	}

	for _, p := range profiles {
		name := p.FirstName
		if name == "" {
			name = p.UserID[:8]
		}
		result[p.UserID] = name
	}
	return result, nil
}
