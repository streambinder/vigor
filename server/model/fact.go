package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// FactArea defines the type of fitness knowledge.
type FactArea string

const (
	AreaTechnique    FactArea = "technique"    // Exercise form, movement execution, cueing, movement patterns
	AreaProgramming  FactArea = "programming"  // Training design, splits, periodization, progression, volume/intensity
	AreaPerformance  FactArea = "performance"  // Strength, power, endurance, speed, agility development
	AreaRecovery     FactArea = "recovery"     // Sleep, rest, recovery modalities, stress management, deload
	AreaInjury       FactArea = "injury"       // Prevention, rehabilitation, pain management
	AreaMobility     FactArea = "mobility"     // Flexibility, range of motion, joint health, movement quality
	AreaNutrition    FactArea = "nutrition"    // Macros, micros, hydration, supplements, timing, body composition
	AreaPhysiology   FactArea = "physiology"   // Body systems, adaptation, energy systems, muscle function
	AreaBiomechanics FactArea = "biomechanics" // Movement mechanics, force production, joint function, leverage
	AreaPsychology   FactArea = "psychology"   // Motivation, adherence, mindset, goal setting, mental training
	AreaAssessment   FactArea = "assessment"   // Testing, measurement, tracking, biofeedback, standards
	AreaEquipment    FactArea = "equipment"    // Proper use, selection, and application of training tools
)

// Fact stores fitness and training science knowledge for RAG retrieval.
type Fact struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Reference string         `gorm:"type:varchar(255);not null;index:idx_knowledge_reference" json:"reference"`
	Area      string         `gorm:"type:varchar(50);not null;index:idx_knowledge_area" json:"area"`
	Tags      pq.StringArray `gorm:"type:text[];index:idx_knowledge_tags,type:gin" json:"tags"`
	Content   string         `gorm:"type:text;not null" json:"content"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type FactEmbedding struct {
	ID        uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null" json:"text"` // The text that was embedded
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`      // all-MiniLM-L6-v2 embedding dimension is 384

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	FactID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_fact_embedding" json:"fact_id"`
	Fact   Fact      `gorm:"foreignKey:FactID;references:ID" json:"fact"`
}
