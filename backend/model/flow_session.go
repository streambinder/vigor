package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
)

var FlowSessionSchema JSONSchemaFormat

// flowSessionLLMOutput is the LLM-facing schema: Poses as a typed slice so the encoder can
// generate proper JSON schema for it (datatypes.JSON is just []byte and would be skipped).
type flowSessionLLMOutput struct {
	Name        string     `json:"name" prompt:"Session name"`
	Description string     `json:"description" prompt:"Brief description of the flow"`
	FactIndices []int      `json:"fact_indices" prompt:"Indices of [FACTS] used (0-based)"`
	Poses       []FlowPose `json:"poses" prompt:"Flat list of poses in sequence"`
}

func init() {
	FlowSessionSchema = JSONSchemaFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:        "flow_session_schema",
			Strict:      true,
			Description: "AI-generated personalized flow/yoga session",
			Schema:      encoder.JSONSchema(flowSessionLLMOutput{}),
		},
	}
}

type FlowSession struct {
	ID          uuid.UUID                              `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`
	Name        string                                 `gorm:"not null" json:"name" prompt:"Session name"`
	Description string                                 `gorm:"not null" json:"description" prompt:"Brief description of the flow"`
	Duration    int                                    `json:"duration" prompt:"-"`
	Muscles     pq.StringArray                         `gorm:"type:text[]" json:"muscles" prompt:"-"`
	Request     string                                 `json:"request" prompt:"-"`
	References  datatypes.JSONType[[]TrainingReference] `gorm:"type:jsonb" json:"references" prompt:"-"`
	FactIndices []int                                  `gorm:"-" json:"fact_indices" prompt:"Indices of [FACTS] used (0-based)"`
	Poses       datatypes.JSON                         `gorm:"type:jsonb" json:"poses" dart:"List<FlowPose>" prompt:"-"`
	Prompt      datatypes.JSONType[TrainingPrompt]     `gorm:"type:jsonb" json:"prompt" prompt:"-"`

	CompletedAt *time.Time `gorm:"type:timestamptz" json:"completed_at" prompt:"-"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at" prompt:"-"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at" prompt:"-"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id" prompt:"-"`
	User        User       `gorm:"constraint:OnDelete:CASCADE;" json:"-" prompt:"-"`
}

type FlowPose struct {
	ExerciseID string          `json:"exercise_id" prompt:"Exercise ID from [EXERCISES] list"`
	Name       string          `json:"name" prompt:"-"`
	Duration   int             `json:"duration" prompt:"Hold duration in seconds (10-60)"`
	Rest       int             `json:"rest" prompt:"Rest after pose in seconds (5-15)"`
	Detail     json.RawMessage `json:"detail" dart:"Map<String, dynamic>" prompt:"-"`
}

// GetPoses unmarshals the JSONB poses field into a typed slice.
func (s *FlowSession) GetPoses() ([]FlowPose, error) {
	var poses []FlowPose
	if err := json.Unmarshal(s.Poses, &poses); err != nil {
		return nil, err
	}
	return poses, nil
}

// SetPoses marshals a typed slice into the JSONB poses field.
func (s *FlowSession) SetPoses(poses []FlowPose) error {
	b, err := json.Marshal(poses)
	if err != nil {
		return err
	}
	s.Poses = b
	return nil
}
