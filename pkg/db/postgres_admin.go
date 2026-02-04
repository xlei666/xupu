package db

import (
	"time"

	"github.com/xlei/xupu/internal/models"
)

// ============================================
// Admin / Config Implementation
// ============================================

func (p *PostgresDatabase) GetSysConfigs() ([]models.SysConfig, error) {
	var configs []models.SysConfig
	err := p.db.Find(&configs).Error
	return configs, err
}

func (p *PostgresDatabase) GetSysConfig(key string) (*models.SysConfig, error) {
	var config models.SysConfig
	err := p.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (p *PostgresDatabase) SaveSysConfig(config *models.SysConfig) error {
	config.UpdatedAt = time.Now()
	// Upsert
	return p.db.Save(config).Error
}

func (p *PostgresDatabase) GetPromptTemplates() ([]models.PromptTemplate, error) {
	var prompts []models.PromptTemplate
	err := p.db.Find(&prompts).Error
	return prompts, err
}

func (p *PostgresDatabase) GetPromptTemplate(key string) (*models.PromptTemplate, error) {
	var prompt models.PromptTemplate
	err := p.db.Where("key = ?", key).First(&prompt).Error
	if err != nil {
		return nil, err
	}
	return &prompt, nil
}

func (p *PostgresDatabase) SavePromptTemplate(prompt *models.PromptTemplate) error {
	prompt.UpdatedAt = time.Now()
	return p.db.Save(prompt).Error
}

func (p *PostgresDatabase) GetNarrativeTemplates() ([]models.NarrativeTemplate, error) {
	var templates []models.NarrativeTemplate
	err := p.db.Find(&templates).Error
	return templates, err
}

func (p *PostgresDatabase) GetNarrativeTemplate(id string) (*models.NarrativeTemplate, error) {
	var template models.NarrativeTemplate
	err := p.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (p *PostgresDatabase) SaveNarrativeTemplate(template *models.NarrativeTemplate) error {
	template.UpdatedAt = time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	return p.db.Save(template).Error
}
