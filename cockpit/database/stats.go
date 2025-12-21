package database

type UserCount struct {
	Count int64
}

type TrainingCount struct {
	Count int64
}

func GetUserCount() (int64, error) {
	if DB == nil {
		return 0, nil
	}
	var result UserCount
	err := DB.Raw(`SELECT count(*) as count FROM users`).Scan(&result).Error
	return result.Count, err
}

func GetTrainingCount() (int64, error) {
	if DB == nil {
		return 0, nil
	}
	var result TrainingCount
	err := DB.Raw(`SELECT count(*) as count FROM trainings`).Scan(&result).Error
	return result.Count, err
}
