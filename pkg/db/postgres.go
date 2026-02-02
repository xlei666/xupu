// Package db PostgreSQL持久化实现
package db

import (
	"fmt"
	"time"

	"github.com/xlei/xupu/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultPostgresConfig 默认PostgreSQL配置
func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:    "localhost",
		Port:    5432,
		User:    "xupu",
		Password: "xupu123",
		DBName:  "xupu",
		SSLMode: "disable",
	}
}

// PostgresDatabase PostgreSQL数据库实现
type PostgresDatabase struct {
	db *gorm.DB
}

// NewPostgres 创建PostgreSQL数据库连接
func NewPostgres(cfg *PostgresConfig) (*PostgresDatabase, error) {
	if cfg == nil {
		cfg = DefaultPostgresConfig()
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &PostgresDatabase{db: db}, nil
}

// Migrate 执行数据库迁移
func (p *PostgresDatabase) Migrate() error {
	return p.db.AutoMigrate(
		&models.WorldSetting{},
		&models.Character{},
		&models.Project{},
		&models.NarrativeBlueprint{},
		&models.SceneOutput{},
		&models.User{},
	)
}

// Close 关闭数据库连接
func (p *PostgresDatabase) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取GORM实例
func (p *PostgresDatabase) GetDB() *gorm.DB {
	return p.db
}

// Save 空实现（PostgreSQL自动提交事务）
func (p *PostgresDatabase) Save() error {
	return nil
}

// ============================================
// WorldSetting CRUD 操作
// ============================================

// SaveWorld 保存世界设定
func (p *PostgresDatabase) SaveWorld(world *models.WorldSetting) error {
	world.UpdatedAt = time.Now()
	if world.CreatedAt.IsZero() {
		world.CreatedAt = time.Now()
	}
	return p.db.Save(world).Error
}

// GetWorld 获取世界设定
func (p *PostgresDatabase) GetWorld(id string) (*models.WorldSetting, error) {
	var world models.WorldSetting
	err := p.db.First(&world, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &world, nil
}

// ListWorlds 列出所有世界设定
func (p *PostgresDatabase) ListWorlds() []*models.WorldSetting {
	var worlds []*models.WorldSetting
	p.db.Order("created_at DESC").Find(&worlds)
	return worlds
}

// DeleteWorld 删除世界设定
func (p *PostgresDatabase) DeleteWorld(id string) error {
	return p.db.Delete(&models.WorldSetting{}, "id = ?", id).Error
}

// UpdateWorldStage 更新世界的特定阶段
func (p *PostgresDatabase) UpdateWorldStage(id string, stage string, data interface{}) error {
	updates := map[string]interface{}{
		fmt.Sprintf("%s", stage): data,
		"updated_at": time.Now(),
	}
	return p.db.Model(&models.WorldSetting{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ============================================
// Character CRUD 操作
// ============================================

// SaveCharacter 保存角色
func (p *PostgresDatabase) SaveCharacter(character *models.Character) error {
	character.UpdatedAt = time.Now()
	if character.CreatedAt.IsZero() {
		character.CreatedAt = time.Now()
	}
	return p.db.Save(character).Error
}

// GetCharacter 获取角色
func (p *PostgresDatabase) GetCharacter(id string) (*models.Character, error) {
	var character models.Character
	err := p.db.First(&character, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &character, nil
}

// ListCharacters 列出所有角色
func (p *PostgresDatabase) ListCharacters() []*models.Character {
	var characters []*models.Character
	p.db.Order("created_at DESC").Find(&characters)
	return characters
}

// ListCharactersByWorld 列出指定世界的角色
func (p *PostgresDatabase) ListCharactersByWorld(worldID string) []*models.Character {
	var characters []*models.Character
	p.db.Where("world_id = ?", worldID).
		Order("created_at DESC").
		Find(&characters)
	return characters
}

// DeleteCharacter 删除角色
func (p *PostgresDatabase) DeleteCharacter(id string) error {
	return p.db.Delete(&models.Character{}, "id = ?", id).Error
}

// UpdateCharacterState 更新角色动态状态
func (p *PostgresDatabase) UpdateCharacterState(id string, state models.DynamicState) error {
	return p.db.Model(&models.Character{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"dynamic_state": state,
			"updated_at":    time.Now(),
		}).Error
}

// ============================================
// Project CRUD 操作
// ============================================

// SaveProject 保存项目
func (p *PostgresDatabase) SaveProject(project *models.Project) error {
	project.UpdatedAt = time.Now()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}
	return p.db.Save(project).Error
}

// GetProject 获取项目
func (p *PostgresDatabase) GetProject(id string) (*models.Project, error) {
	var project models.Project
	err := p.db.First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects 列出所有项目
func (p *PostgresDatabase) ListProjects() []*models.Project {
	var projects []*models.Project
	p.db.Order("created_at DESC").Find(&projects)
	return projects
}

// ListProjectsByUser 列出指定用户的项目
func (p *PostgresDatabase) ListProjectsByUser(userID string) []*models.Project {
	var projects []*models.Project
	p.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&projects)
	return projects
}

// DeleteProject 删除项目
func (p *PostgresDatabase) DeleteProject(id string) error {
	// 级联删除关联的蓝图和场景
	p.db.Where("project_id = ?", id).Delete(&models.NarrativeBlueprint{})
	p.db.Where("blueprint_id IN (SELECT id FROM narrative_blueprints WHERE project_id = ?)", id).
		Delete(&models.SceneOutput{})
	return p.db.Delete(&models.Project{}, "id = ?", id).Error
}

// UpdateProjectStatus 更新项目状态
func (p *PostgresDatabase) UpdateProjectStatus(id string, status models.ProjectStatus, progress float64) error {
	return p.db.Model(&models.Project{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"progress":   progress,
			"updated_at": time.Now(),
		}).Error
}

// ============================================
// NarrativeBlueprint CRUD 操作
// ============================================

// SaveBlueprint 保存叙事蓝图
func (p *PostgresDatabase) SaveBlueprint(blueprint *models.NarrativeBlueprint) error {
	blueprint.UpdatedAt = time.Now()
	if blueprint.CreatedAt.IsZero() {
		blueprint.CreatedAt = time.Now()
	}
	return p.db.Save(blueprint).Error
}

// GetBlueprint 获取叙事蓝图
func (p *PostgresDatabase) GetBlueprint(id string) (*models.NarrativeBlueprint, error) {
	var blueprint models.NarrativeBlueprint
	err := p.db.First(&blueprint, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &blueprint, nil
}

// ListBlueprints 列出所有叙事蓝图
func (p *PostgresDatabase) ListBlueprints() []*models.NarrativeBlueprint {
	var blueprints []*models.NarrativeBlueprint
	p.db.Order("created_at DESC").Find(&blueprints)
	return blueprints
}

// DeleteBlueprint 删除叙事蓝图
func (p *PostgresDatabase) DeleteBlueprint(id string) error {
	// 级联删除场景
	p.db.Where("blueprint_id = ?", id).Delete(&models.SceneOutput{})
	return p.db.Delete(&models.NarrativeBlueprint{}, "id = ?", id).Error
}

// SaveNarrativeBlueprint 保存叙事蓝图（别名方法）
func (p *PostgresDatabase) SaveNarrativeBlueprint(blueprint *models.NarrativeBlueprint) error {
	return p.SaveBlueprint(blueprint)
}

// GetNarrativeBlueprint 获取叙事蓝图（别名方法）
func (p *PostgresDatabase) GetNarrativeBlueprint(id string) (*models.NarrativeBlueprint, error) {
	return p.GetBlueprint(id)
}

// ============================================
// SceneOutput CRUD 操作
// ============================================

// SaveScene 保存场景输出
func (p *PostgresDatabase) SaveScene(scene *models.SceneOutput) error {
	if scene.CreatedAt.IsZero() {
		scene.CreatedAt = time.Now()
	}
	return p.db.Save(scene).Error
}

// GetScene 获取场景输出
func (p *PostgresDatabase) GetScene(id string) (*models.SceneOutput, error) {
	var scene models.SceneOutput
	err := p.db.First(&scene, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

// ListScenesByBlueprint 列出指定蓝图的所有场景
func (p *PostgresDatabase) ListScenesByBlueprint(blueprintID string) []*models.SceneOutput {
	var scenes []*models.SceneOutput
	p.db.Where("blueprint_id = ?", blueprintID).
		Order("chapter, scene").
		Find(&scenes)
	return scenes
}

// ListScenesByChapter 列出指定章节的所有场景
func (p *PostgresDatabase) ListScenesByChapter(blueprintID string, chapter int) []*models.SceneOutput {
	var scenes []*models.SceneOutput
	p.db.Where("blueprint_id = ? AND chapter = ?", blueprintID, chapter).
		Order("scene").
		Find(&scenes)
	return scenes
}

// GetSceneByBlueprintAndChapter 获取指定蓝图和章节的场景
func (p *PostgresDatabase) GetSceneByBlueprintAndChapter(blueprintID string, chapter, sceneNum int) (*models.SceneOutput, error) {
	var scene models.SceneOutput
	err := p.db.Where("blueprint_id = ? AND chapter = ? AND scene = ?",
		blueprintID, chapter, sceneNum).
		First(&scene).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

// ============================================
// User CRUD 操作
// ============================================

// SaveUser 保存用户
func (p *PostgresDatabase) SaveUser(user *models.User) error {
	user.UpdatedAt = time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	return p.db.Save(user).Error
}

// GetUser 获取用户
func (p *PostgresDatabase) GetUser(id string) (*models.User, error) {
	var user models.User
	err := p.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (p *PostgresDatabase) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := p.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByAPIKey 根据API密钥获取用户
func (p *PostgresDatabase) GetUserByAPIKey(apiKey string) (*models.User, error) {
	var user models.User
	err := p.db.First(&user, "api_key = ?", apiKey).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ============================================
// NarrativeNode 相关方法
// ============================================

// SaveNarrativeNode 保存叙事节点
func (p *PostgresDatabase) SaveNarrativeNode(node *models.NarrativeNode) error {
	return p.db.Save(node).Error
}

// GetNarrativeNode 获取叙事节点
func (p *PostgresDatabase) GetNarrativeNode(id string) (*models.NarrativeNode, error) {
	var node models.NarrativeNode
	err := p.db.First(&node, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ListNarrativeNodes 列出所有叙事节点
func (p *PostgresDatabase) ListNarrativeNodes() []*models.NarrativeNode {
	var nodes []*models.NarrativeNode
	p.db.Order("created_at DESC").Find(&nodes)
	return nodes
}

// ListNarrativeNodesByProject 列出指定项目的叙事节点
func (p *PostgresDatabase) ListNarrativeNodesByProject(projectID string) []*models.NarrativeNode {
	var nodes []*models.NarrativeNode
	p.db.Where("project_id = ?", projectID).Order("node_order ASC, created_at ASC").Find(&nodes)
	return nodes
}

// ListNarrativeNodesByParent 列出指定父节点的子节点
func (p *PostgresDatabase) ListNarrativeNodesByParent(parentID string) []*models.NarrativeNode {
	var nodes []*models.NarrativeNode
	p.db.Where("parent_id = ?", parentID).Order("node_order ASC").Find(&nodes)
	return nodes
}

// DeleteNarrativeNode 删除叙事节点
func (p *PostgresDatabase) DeleteNarrativeNode(id string) error {
	return p.db.Delete(&models.NarrativeNode{}, "id = ?", id).Error
}

// ============================================
// NodeChapterMapping 相关方法
// ============================================

// SaveNodeChapterMapping 保存节点到章节的映射
func (p *PostgresDatabase) SaveNodeChapterMapping(mapping *models.NodeChapterMapping) error {
	return p.db.Save(mapping).Error
}

// ListNodeChapterMappingsByProject 列出指定项目的所有映射
func (p *PostgresDatabase) ListNodeChapterMappingsByProject(projectID string) []*models.NodeChapterMapping {
	var mappings []*models.NodeChapterMapping
	p.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&mappings)
	return mappings
}

// ListNodeChapterMappingsByChapter 列出指定章节的所有映射
func (p *PostgresDatabase) ListNodeChapterMappingsByChapter(chapterID string) []*models.NodeChapterMapping {
	var mappings []*models.NodeChapterMapping
	p.db.Where("chapter_id = ?", chapterID).Order("sequence ASC").Find(&mappings)
	return mappings
}

// ============================================
// Chapter 相关方法
// ============================================

// SaveChapter 保存章节
func (p *PostgresDatabase) SaveChapter(chapter *models.Chapter) error {
	return p.db.Save(chapter).Error
}

// GetChapter 获取章节
func (p *PostgresDatabase) GetChapter(id string) (*models.Chapter, error) {
	var chapter models.Chapter
	err := p.db.First(&chapter, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

// ListChaptersByProject 列出指定项目的所有章节
func (p *PostgresDatabase) ListChaptersByProject(projectID string) []*models.Chapter {
	var chapters []*models.Chapter
	p.db.Where("project_id = ?", projectID).Order("chapter_num ASC").Find(&chapters)
	return chapters
}

// DeleteChapter 删除章节
func (p *PostgresDatabase) DeleteChapter(id string) error {
	return p.db.Delete(&models.Chapter{}, "id = ?", id).Error
}

// ============================================
// 统计和工具方法
// ============================================

// Stats 获取数据库统计信息
func (p *PostgresDatabase) Stats() map[string]int {
	var worlds, characters, projects, blueprints, scenes int64

	p.db.Model(&models.WorldSetting{}).Count(&worlds)
	p.db.Model(&models.Character{}).Count(&characters)
	p.db.Model(&models.Project{}).Count(&projects)
	p.db.Model(&models.NarrativeBlueprint{}).Count(&blueprints)
	p.db.Model(&models.SceneOutput{}).Count(&scenes)

	return map[string]int{
		"worlds":     int(worlds),
		"characters": int(characters),
		"projects":   int(projects),
		"blueprints": int(blueprints),
		"scenes":     int(scenes),
	}
}

// Clear 清空所有数据（慎用）
func (p *PostgresDatabase) Clear() error {
	return p.db.Where("1 = 1").Delete(&models.SceneOutput{}).
		Error
}

// SaveSynopsis 保存简介
func (p *PostgresDatabase) SaveSynopsis(synopsis *models.Synopsis) error {
	return p.db.Create(synopsis).Error
}

// GetSynopsis 获取简介
func (p *PostgresDatabase) GetSynopsis(id string) (*models.Synopsis, error) {
	var synopsis models.Synopsis
	err := p.db.Where("id = ?", id).First(&synopsis).Error
	if err != nil {
		return nil, err
	}
	return &synopsis, nil
}
