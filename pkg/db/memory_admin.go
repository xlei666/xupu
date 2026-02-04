package db

import (
	"errors"

	"github.com/xlei/xupu/internal/models"
)

// ============================================
// Admin / Config Implementation (Stub for Memory DB)
// ============================================

func (d *MemoryDatabase) GetSysConfigs() ([]models.SysConfig, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) GetSysConfig(key string) (*models.SysConfig, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) SaveSysConfig(config *models.SysConfig) error {
	return errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) GetPromptTemplates() ([]models.PromptTemplate, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) GetPromptTemplate(key string) (*models.PromptTemplate, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) SavePromptTemplate(prompt *models.PromptTemplate) error {
	return errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) GetNarrativeTemplates() ([]models.NarrativeTemplate, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) GetNarrativeTemplate(id string) (*models.NarrativeTemplate, error) {
	return nil, errors.New("not implemented in memory db")
}

func (d *MemoryDatabase) SaveNarrativeTemplate(template *models.NarrativeTemplate) error {
	return errors.New("not implemented in memory db")
}
