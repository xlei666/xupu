package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

// ============================================
// 系统配置
// ============================================

// SysConfig 系统配置
type SysConfig struct {
	Key         string    `json:"key" gorm:"primaryKey"`
	Value       string    `json:"value" gorm:"type:text"`
	Type        string    `json:"type" gorm:"size:20"` // string, int, bool, json
	Description string    `json:"description" gorm:"size:255"`
	Group       string    `json:"group" gorm:"size:50;index"`     // llm, system, feature
	IsSecret    bool      `json:"is_secret" gorm:"default:false"` // masking in UI
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================
// 提示词管理
// ============================================

// PromptTemplate 提示词模板
type PromptTemplate struct {
	Key         string    `json:"key" gorm:"primaryKey"`
	Content     string    `json:"content" gorm:"type:text"`
	Description string    `json:"description" gorm:"size:255"`
	Variables   JSON      `json:"variables" gorm:"type:json"`    // []string
	ModelConfig JSON      `json:"model_config" gorm:"type:json"` // {model, temperature}
	Version     int       `json:"version" gorm:"default:1"`
	Tags        JSON      `json:"tags" gorm:"type:json"` // []string
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================
// 叙事结构模板
// ============================================

// NarrativeTemplate 叙事模板
type NarrativeTemplate struct {
	ID          string    `json:"id" gorm:"primaryKey"` // e.g. "infinite_flow"
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Structure   JSON      `json:"structure" gorm:"type:json"`    // Definition of acts/stages
	PromptRules JSON      `json:"prompt_rules" gorm:"type:json"` // How to prompt for this structure
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================
// 通用 JSON 类型
// ============================================

// JSON is a wrapper for handling JSON in GORM
type JSON []byte

// Value marshals the JSON value
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan unmarshals the JSON value
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return errors.New("type assertion to []byte or string failed")
		}
		*j = []byte(str)
		return nil
	}
	*j = s
	return nil
}

// MarshalJSON returns *j as the JSON encoding of j.
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data.
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
