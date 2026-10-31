package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"gorm.io/gorm"
)

// knowledgeExerciseAnimation is one generated exercise animation in the
// knowledge database, keyed by the catalog exercise it illustrates. The
// table is image-owned: the backend reads it and never migrates it.
type knowledgeExerciseAnimation struct {
	ExerciseID string `gorm:"primaryKey;column:exercise_id"`
	Animation  string `gorm:"column:animation"`
}

// TableName pins the image-owned table name.
func (knowledgeExerciseAnimation) TableName() string { return "exercise_animations" }

// knowledgeExerciseStill is one generated exercise still in the knowledge
// database, keyed by the catalog exercise it illustrates. The table is
// image-owned: the backend reads it and never migrates it.
type knowledgeExerciseStill struct {
	ExerciseID string `gorm:"primaryKey;column:exercise_id"`
	Still      string `gorm:"column:still"`
}

// TableName pins the image-owned table name.
func (knowledgeExerciseStill) TableName() string { return "exercise_stills" }

// knowledgeExerciseModel is one renderable model in the knowledge
// database, keyed by its identifier. The table is image-owned: the
// backend reads it and never migrates it.
type knowledgeExerciseModel struct {
	ModelID     string `gorm:"primaryKey;column:model_id"`
	DisplayName string `gorm:"column:display_name"`
	Model       []byte `gorm:"column:model"`
	Licence     string `gorm:"column:licence"`
	Source      string `gorm:"column:source"`
	MD5         string `gorm:"column:md5"`
}

// TableName pins the image-owned table name.
func (knowledgeExerciseModel) TableName() string { return "exercise_models" }

// animatedExerciseStill is one animated exercise with its still, as the
// app thumbnails consume it.
type animatedExerciseStill struct {
	ExerciseID string `gorm:"column:exercise_id" json:"exercise_id"`
	Still      string `gorm:"column:still" json:"still"`
}

// exerciseModelMeta is one model without its bytes, as the model list
// consumes it.
type exerciseModelMeta struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Licence     string `json:"licence"`
	Source      string `json:"source"`
	MD5         string `json:"md5"`
}

func initExerciseAnimations(app *fiber.App) {
	app.Get("/exercises/:id/animation", middleware.Authorized(), getExerciseAnimation)
	app.Get("/exercises/animated", middleware.Authorized(), getAnimatedExercises)
	app.Get("/exercises/models", middleware.Authorized(), getExerciseModels)
	app.Get("/exercises/models/:id", middleware.Authorized(), getExerciseModel)
}

func getExerciseAnimation(c *fiber.Ctx) error {
	var row knowledgeExerciseAnimation
	err := database.Knowledge.Where("exercise_id = ?", c.Params("id")).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "animation not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send([]byte(row.Animation))
}

func getAnimatedExercises(c *fiber.Ctx) error {
	var rows []animatedExerciseStill
	err := database.Knowledge.Table("exercise_animations").
		Select("exercise_animations.exercise_id, exercise_stills.still").
		Joins("JOIN exercise_stills ON exercise_stills.exercise_id = exercise_animations.exercise_id").
		Order("exercise_animations.exercise_id").
		Scan(&rows).Error
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if rows == nil {
		rows = []animatedExerciseStill{}
	}
	return c.JSON(rows)
}

func getExerciseModels(c *fiber.Ctx) error {
	var rows []knowledgeExerciseModel
	if err := database.Knowledge.Order("model_id").Find(&rows).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	metas := make([]exerciseModelMeta, 0, len(rows))
	for _, row := range rows {
		metas = append(metas, exerciseModelMeta{
			ID:          row.ModelID,
			DisplayName: row.DisplayName,
			Licence:     row.Licence,
			Source:      row.Source,
			MD5:         row.MD5,
		})
	}
	return c.JSON(metas)
}

func getExerciseModel(c *fiber.Ctx) error {
	var row knowledgeExerciseModel
	err := database.Knowledge.Where("model_id = ?", c.Params("id")).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "model not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	c.Set(fiber.HeaderContentType, "model/gltf-binary")
	c.Set(fiber.HeaderContentLength, strconv.Itoa(len(row.Model)))
	c.Set(fiber.HeaderETag, `"`+row.MD5+`"`)
	return c.Send(row.Model)
}
