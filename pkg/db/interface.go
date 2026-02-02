// Package db 数据库接口和工厂
package db

import (
	"github.com/xlei/xupu/internal/models"
)

// Database 数据库接口
type Database interface {
	// 通用方法
	Save() error

	// WorldSetting
	SaveWorld(world *models.WorldSetting) error
	GetWorld(id string) (*models.WorldSetting, error)
	ListWorlds() []*models.WorldSetting
	DeleteWorld(id string) error
	UpdateWorldStage(id string, stage string, data interface{}) error

	// Character
	SaveCharacter(character *models.Character) error
	GetCharacter(id string) (*models.Character, error)
	ListCharacters() []*models.Character
	ListCharactersByWorld(worldID string) []*models.Character
	DeleteCharacter(id string) error
	UpdateCharacterState(id string, state models.DynamicState) error

	// Synopsis
	SaveSynopsis(synopsis *models.Synopsis) error
	GetSynopsis(id string) (*models.Synopsis, error)

	// Project
	SaveProject(project *models.Project) error
	GetProject(id string) (*models.Project, error)
	ListProjects() []*models.Project
	ListProjectsByUser(userID string) []*models.Project
	DeleteProject(id string) error
	UpdateProjectStatus(id string, status models.ProjectStatus, progress float64) error

	// NarrativeBlueprint
	SaveBlueprint(blueprint *models.NarrativeBlueprint) error
	GetBlueprint(id string) (*models.NarrativeBlueprint, error)
	ListBlueprints() []*models.NarrativeBlueprint
	DeleteBlueprint(id string) error
	SaveNarrativeBlueprint(blueprint *models.NarrativeBlueprint) error
	GetNarrativeBlueprint(id string) (*models.NarrativeBlueprint, error)

	// SceneOutput
	SaveScene(scene *models.SceneOutput) error
	GetScene(id string) (*models.SceneOutput, error)
	ListScenesByBlueprint(blueprintID string) []*models.SceneOutput
	ListScenesByChapter(blueprintID string, chapter int) []*models.SceneOutput
	GetSceneByBlueprintAndChapter(blueprintID string, chapter, sceneNum int) (*models.SceneOutput, error)

	// NarrativeNode
	SaveNarrativeNode(node *models.NarrativeNode) error
	GetNarrativeNode(id string) (*models.NarrativeNode, error)
	ListNarrativeNodes() []*models.NarrativeNode
	ListNarrativeNodesByProject(projectID string) []*models.NarrativeNode
	ListNarrativeNodesByParent(parentID string) []*models.NarrativeNode
	DeleteNarrativeNode(id string) error

	// NodeChapterMapping
	SaveNodeChapterMapping(mapping *models.NodeChapterMapping) error
	ListNodeChapterMappingsByProject(projectID string) []*models.NodeChapterMapping
	ListNodeChapterMappingsByChapter(chapterID string) []*models.NodeChapterMapping

	// Chapter
	SaveChapter(chapter *models.Chapter) error
	GetChapter(id string) (*models.Chapter, error)
	ListChaptersByProject(projectID string) []*models.Chapter
	DeleteChapter(id string) error

	// User
	SaveUser(user *models.User) error
	GetUser(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByAPIKey(apiKey string) (*models.User, error)

	// Utilities
	Stats() map[string]int
	Clear() error
}

// DBType 数据库类型
type DBType string

const (
	DBTypeMemory   DBType = "memory"   // 内存数据库
	DBTypePostgres DBType = "postgres" // PostgreSQL
)

// Config 数据库配置
type Config struct {
	Type     DBType
	Postgres *PostgresConfig
	DataDir  string // 用于内存数据库的数据目录
}

// Init 根据配置初始化数据库
func Init(cfg *Config) (Database, error) {
	if cfg == nil {
		cfg = &Config{
			Type:    DBTypeMemory,
			DataDir: "data",
		}
	}

	switch cfg.Type {
	case DBTypePostgres:
		if cfg.Postgres == nil {
			cfg.Postgres = DefaultPostgresConfig()
		}
		return NewPostgres(cfg.Postgres)
	case DBTypeMemory:
		fallthrough
	default:
		return New(cfg.DataDir), nil
	}
}

// SetGlobal 设置全局数据库实例
func SetGlobal(db Database) {
	defaultDB = db
}
