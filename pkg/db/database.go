// Package db 提供数据持久化功能
// 支持内存数据库和PostgreSQL
package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xlei/xupu/internal/models"
)

// MemoryDatabase 内存数据库实现（Database接口的具体实现）
type MemoryDatabase struct {
	mu sync.RWMutex

	// 数据存储
	worlds              map[string]*models.WorldSetting
	characters          map[string]*models.Character
	synopses            map[string]*models.Synopsis
	projects            map[string]*models.Project
	blueprints          map[string]*models.NarrativeBlueprint
	scenes              map[string]*models.SceneOutput
	users               map[string]*models.User
	narrativeNodes      map[string]*models.NarrativeNode
	nodeChapterMappings map[string]*models.NodeChapterMapping
	chapters            map[string]*models.Chapter

	// 配置
	dataDir  string
	autoSave bool
}

var (
	// defaultDB 默认数据库实例
	defaultDB Database
	once      sync.Once
)

// Get 获取数据库实例（单例）
func Get() Database {
	once.Do(func() {
		// 使用PostgreSQL
		var err error
		pgDB, err := NewPostgres(nil) // nil means use strict default config (which reads envs)
		if err != nil {
			// Fallback to memory or panic?
			// Panic is better to ensure we know it failed
			fmt.Printf("Initial DB connection failed: %v\n", err)
			panic("failed to connect to database")
		}

		// 自动迁移
		if err := pgDB.Migrate(); err != nil {
			fmt.Printf("DB Migration failed: %v\n", err)
		}

		defaultDB = pgDB
	})
	return defaultDB
}

// NewMemory 创建新的内存数据库实例
func NewMemory(dataDir string) Database {
	db := &MemoryDatabase{
		worlds:              make(map[string]*models.WorldSetting),
		characters:          make(map[string]*models.Character),
		projects:            make(map[string]*models.Project),
		blueprints:          make(map[string]*models.NarrativeBlueprint),
		scenes:              make(map[string]*models.SceneOutput),
		users:               make(map[string]*models.User),
		narrativeNodes:      make(map[string]*models.NarrativeNode),
		nodeChapterMappings: make(map[string]*models.NodeChapterMapping),
		chapters:            make(map[string]*models.Chapter),
		dataDir:             dataDir,
		autoSave:            true,
	}

	// 确保数据目录存在
	os.MkdirAll(dataDir, 0755)

	// 尝试加载数据
	db.load()

	return db
}

// New 创建新的数据库实例（兼容旧代码）
func New(dataDir string) Database {
	return NewMemory(dataDir)
}

// SetAutoSave 设置是否自动保存
func (d *MemoryDatabase) SetAutoSave(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.autoSave = enabled
}

// Save 保存数据到磁盘
func (d *MemoryDatabase) Save() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.save()
}

// save 内部保存方法（调用时需已持有锁）
func (d *MemoryDatabase) save() error {
	if d.dataDir == "" {
		return nil
	}

	// 保存世界设定
	if err := d.saveTable("worlds.json", d.worlds); err != nil {
		return fmt.Errorf("保存worlds失败: %w", err)
	}

	// 保存角色
	if err := d.saveTable("characters.json", d.characters); err != nil {
		return fmt.Errorf("保存characters失败: %w", err)
	}

	// 保存项目
	if err := d.saveTable("projects.json", d.projects); err != nil {
		return fmt.Errorf("保存projects失败: %w", err)
	}

	// 保存叙事蓝图
	if err := d.saveTable("blueprints.json", d.blueprints); err != nil {
		return fmt.Errorf("保存blueprints失败: %w", err)
	}

	// 保存场景
	if err := d.saveTable("scenes.json", d.scenes); err != nil {
		return fmt.Errorf("保存scenes失败: %w", err)
	}

	// 保存叙事节点
	if err := d.saveTable("narrative_nodes.json", d.narrativeNodes); err != nil {
		return fmt.Errorf("保存narrative_nodes失败: %w", err)
	}

	// 保存节点章节映射
	if err := d.saveTable("node_chapter_mappings.json", d.nodeChapterMappings); err != nil {
		return fmt.Errorf("保存node_chapter_mappings失败: %w", err)
	}

	// 保存章节
	if err := d.saveTable("chapters.json", d.chapters); err != nil {
		return fmt.Errorf("保存chapters失败: %w", err)
	}

	return nil
}

// saveTable 保存单个数据表
func (d *MemoryDatabase) saveTable(filename string, data interface{}) error {
	path := filepath.Join(d.dataDir, filename)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// load 从磁盘加载数据
func (d *MemoryDatabase) load() error {
	d.loadTable("worlds.json", &d.worlds)
	d.loadTable("characters.json", &d.characters)
	d.loadTable("projects.json", &d.projects)
	d.loadTable("blueprints.json", &d.blueprints)
	d.loadTable("scenes.json", &d.scenes)
	d.loadTable("narrative_nodes.json", &d.narrativeNodes)
	d.loadTable("node_chapter_mappings.json", &d.nodeChapterMappings)
	d.loadTable("chapters.json", &d.chapters)
	return nil
}

// loadTable 加载单个数据表
func (d *MemoryDatabase) loadTable(filename string, target interface{}) error {
	path := filepath.Join(d.dataDir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在是正常情况
		}
		return err
	}

	return json.Unmarshal(data, target)
}

// ============================================
// WorldSetting CRUD 操作
// ============================================

// SaveWorld 保存世界设定
func (d *MemoryDatabase) SaveWorld(world *models.WorldSetting) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 设置更新时间
	world.UpdatedAt = time.Now()
	if world.CreatedAt.IsZero() {
		world.CreatedAt = time.Now()
	}

	d.worlds[world.ID] = world

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetWorld 获取世界设定
func (d *MemoryDatabase) GetWorld(id string) (*models.WorldSetting, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	world, ok := d.worlds[id]
	if !ok {
		return nil, ErrNotFound
	}
	return world, nil
}

// ListWorlds 列出所有世界设定
func (d *MemoryDatabase) ListWorlds() []*models.WorldSetting {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.WorldSetting, 0, len(d.worlds))
	for _, world := range d.worlds {
		result = append(result, world)
	}
	return result
}

// DeleteWorld 删除世界设定
func (d *MemoryDatabase) DeleteWorld(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.worlds[id]; !ok {
		return ErrNotFound
	}

	delete(d.worlds, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// UpdateWorldStage 更新世界的特定阶段
func (d *MemoryDatabase) UpdateWorldStage(id string, stage string, data interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	world, ok := d.worlds[id]
	if !ok {
		return ErrNotFound
	}

	// 根据阶段更新对应字段（支持指针和值类型）
	switch stage {
	case "philosophy":
		switch v := data.(type) {
		case models.Philosophy:
			world.Philosophy = v
		case *models.Philosophy:
			world.Philosophy = *v
		}
	case "worldview":
		switch v := data.(type) {
		case models.Worldview:
			world.Worldview = v
		case *models.Worldview:
			world.Worldview = *v
		}
	case "laws":
		switch v := data.(type) {
		case models.Laws:
			world.Laws = v
		case *models.Laws:
			world.Laws = *v
		}
	case "geography":
		switch v := data.(type) {
		case models.Geography:
			world.Geography = v
		case *models.Geography:
			world.Geography = *v
		}
	case "civilization":
		switch v := data.(type) {
		case models.Civilization:
			world.Civilization = v
		case *models.Civilization:
			world.Civilization = *v
		}
	case "society":
		switch v := data.(type) {
		case models.Society:
			world.Society = v
		case *models.Society:
			world.Society = *v
		}
	case "history":
		switch v := data.(type) {
		case models.History:
			world.History = v
		case *models.History:
			world.History = *v
		}
	case "story_soil":
		switch v := data.(type) {
		case models.StorySoil:
			world.StorySoil = v
		case *models.StorySoil:
			world.StorySoil = *v
		}
	default:
		return fmt.Errorf("未知阶段: %s", stage)
	}

	world.UpdatedAt = time.Now()

	if d.autoSave {
		return d.save()
	}
	return nil
}

// ============================================
// Character CRUD 操作
// ============================================

// SaveCharacter 保存角色
func (d *MemoryDatabase) SaveCharacter(character *models.Character) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	character.UpdatedAt = time.Now()
	if character.CreatedAt.IsZero() {
		character.CreatedAt = time.Now()
	}

	d.characters[character.ID] = character

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetCharacter 获取角色
func (d *MemoryDatabase) GetCharacter(id string) (*models.Character, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	character, ok := d.characters[id]
	if !ok {
		return nil, ErrNotFound
	}
	return character, nil
}

// ListCharacters 列出所有角色
func (d *MemoryDatabase) ListCharacters() []*models.Character {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.Character, 0, len(d.characters))
	for _, char := range d.characters {
		result = append(result, char)
	}
	return result
}

// ListCharactersByWorld 列出指定世界的角色
func (d *MemoryDatabase) ListCharactersByWorld(worldID string) []*models.Character {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.Character, 0)
	for _, char := range d.characters {
		if char.WorldID == worldID {
			result = append(result, char)
		}
	}
	return result
}

// DeleteCharacter 删除角色
func (d *MemoryDatabase) DeleteCharacter(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.characters[id]; !ok {
		return ErrNotFound
	}

	delete(d.characters, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// UpdateCharacterState 更新角色动态状态
func (d *MemoryDatabase) UpdateCharacterState(id string, state models.DynamicState) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	character, ok := d.characters[id]
	if !ok {
		return ErrNotFound
	}

	character.DynamicState = state
	character.UpdatedAt = time.Now()

	if d.autoSave {
		return d.save()
	}
	return nil
}

// ============================================
// Project CRUD 操作
// ============================================

// SaveProject 保存项目
func (d *MemoryDatabase) SaveProject(project *models.Project) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	project.UpdatedAt = time.Now()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}

	d.projects[project.ID] = project

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetProject 获取项目
func (d *MemoryDatabase) GetProject(id string) (*models.Project, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	project, ok := d.projects[id]
	if !ok {
		return nil, ErrNotFound
	}
	return project, nil
}

// ListProjects 列出所有项目
func (d *MemoryDatabase) ListProjects() []*models.Project {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.Project, 0, len(d.projects))
	for _, proj := range d.projects {
		result = append(result, proj)
	}
	return result
}

// ListProjectsByUser 列出指定用户的项目
func (d *MemoryDatabase) ListProjectsByUser(userID string) []*models.Project {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.Project, 0)
	for _, proj := range d.projects {
		if proj.UserID == userID {
			result = append(result, proj)
		}
	}
	return result
}

// DeleteProject 删除项目
func (d *MemoryDatabase) DeleteProject(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.projects[id]; !ok {
		return ErrNotFound
	}

	delete(d.projects, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// UpdateProjectStatus 更新项目状态
func (d *MemoryDatabase) UpdateProjectStatus(id string, status models.ProjectStatus, progress float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	project, ok := d.projects[id]
	if !ok {
		return ErrNotFound
	}

	project.Status = status
	project.Progress = progress
	project.UpdatedAt = time.Now()

	if d.autoSave {
		return d.save()
	}
	return nil
}

// ============================================
// NarrativeBlueprint CRUD 操作
// ============================================

// SaveBlueprint 保存叙事蓝图
func (d *MemoryDatabase) SaveBlueprint(blueprint *models.NarrativeBlueprint) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	blueprint.UpdatedAt = time.Now()
	if blueprint.CreatedAt.IsZero() {
		blueprint.CreatedAt = time.Now()
	}

	d.blueprints[blueprint.ID] = blueprint

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetBlueprint 获取叙事蓝图
func (d *MemoryDatabase) GetBlueprint(id string) (*models.NarrativeBlueprint, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	blueprint, ok := d.blueprints[id]
	if !ok {
		return nil, ErrNotFound
	}
	return blueprint, nil
}

// ListBlueprints 列出所有叙事蓝图
func (d *MemoryDatabase) ListBlueprints() []*models.NarrativeBlueprint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NarrativeBlueprint, 0, len(d.blueprints))
	for _, bp := range d.blueprints {
		result = append(result, bp)
	}
	return result
}

// DeleteBlueprint 删除叙事蓝图
func (d *MemoryDatabase) DeleteBlueprint(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.blueprints[id]; !ok {
		return ErrNotFound
	}

	delete(d.blueprints, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// SaveNarrativeBlueprint 保存叙事蓝图（别名方法）
func (d *MemoryDatabase) SaveNarrativeBlueprint(blueprint *models.NarrativeBlueprint) error {
	return d.SaveBlueprint(blueprint)
}

// GetNarrativeBlueprint 获取叙事蓝图（别名方法）
func (d *MemoryDatabase) GetNarrativeBlueprint(id string) (*models.NarrativeBlueprint, error) {
	return d.GetBlueprint(id)
}

// ============================================
// SceneOutput CRUD 操作
// ============================================

// SaveScene 保存场景输出
func (d *MemoryDatabase) SaveScene(scene *models.SceneOutput) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if scene.CreatedAt.IsZero() {
		scene.CreatedAt = time.Now()
	}

	d.scenes[scene.ID] = scene

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetScene 获取场景输出
func (d *MemoryDatabase) GetScene(id string) (*models.SceneOutput, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	scene, ok := d.scenes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return scene, nil
}

// ListScenesByBlueprint 列出指定蓝图的所有场景
func (d *MemoryDatabase) ListScenesByBlueprint(blueprintID string) []*models.SceneOutput {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.SceneOutput, 0)
	for _, scene := range d.scenes {
		if scene.BlueprintID == blueprintID {
			result = append(result, scene)
		}
	}
	return result
}

// ListScenesByChapter 列出指定章节的所有场景
func (d *MemoryDatabase) ListScenesByChapter(blueprintID string, chapter int) []*models.SceneOutput {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.SceneOutput, 0)
	for _, scene := range d.scenes {
		if scene.BlueprintID == blueprintID && scene.Chapter == chapter {
			result = append(result, scene)
		}
	}
	return result
}

// GetSceneByBlueprintAndChapter 获取指定蓝图和章节的场景
func (d *MemoryDatabase) GetSceneByBlueprintAndChapter(blueprintID string, chapter, sceneNum int) (*models.SceneOutput, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, s := range d.scenes {
		if s.BlueprintID == blueprintID && s.Chapter == chapter && s.Scene == sceneNum {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

// ============================================
// 错误定义
// ============================================

// ErrNotFound 记录不存在错误
var ErrNotFound = fmt.Errorf("record not found")

// IsNotFound 判断是否为记录不存在错误
func IsNotFound(err error) bool {
	return err == ErrNotFound || strings.Contains(err.Error(), "not found")
}

// ============================================
// 工具方法
// ============================================

// GenerateID 生成唯一ID
func GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", prefix, timestamp)
}

// Clear 清空所有数据（慎用）
func (d *MemoryDatabase) Clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.worlds = make(map[string]*models.WorldSetting)
	d.characters = make(map[string]*models.Character)
	d.projects = make(map[string]*models.Project)
	d.blueprints = make(map[string]*models.NarrativeBlueprint)
	d.scenes = make(map[string]*models.SceneOutput)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// Stats 获取数据库统计信息
func (d *MemoryDatabase) Stats() map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]int{
		"worlds":     len(d.worlds),
		"characters": len(d.characters),
		"projects":   len(d.projects),
		"blueprints": len(d.blueprints),
		"scenes":     len(d.scenes),
		"users":      len(d.users),
	}
}

// ============================================
// User CRUD 操作
// ============================================

// SaveUser 保存用户
func (d *MemoryDatabase) SaveUser(user *models.User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	user.UpdatedAt = time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	d.users[user.ID] = user

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetUser 获取用户
func (d *MemoryDatabase) GetUser(id string) (*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	user, ok := d.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (d *MemoryDatabase) GetUserByEmail(email string) (*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, user := range d.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

// GetUserByAPIKey 根据API密钥获取用户
func (d *MemoryDatabase) GetUserByAPIKey(apiKey string) (*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, user := range d.users {
		if user.APIKey != nil && *user.APIKey == apiKey {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

// ============================================
// NarrativeNode CRUD 操作
// ============================================

// SaveNarrativeNode 保存叙事节点
func (d *MemoryDatabase) SaveNarrativeNode(node *models.NarrativeNode) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	node.UpdatedAt = time.Now()
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now()
	}

	d.narrativeNodes[node.ID] = node

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetNarrativeNode 获取叙事节点
func (d *MemoryDatabase) GetNarrativeNode(id string) (*models.NarrativeNode, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	node, ok := d.narrativeNodes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return node, nil
}

// ListNarrativeNodes 列出所有叙事节点
func (d *MemoryDatabase) ListNarrativeNodes() []*models.NarrativeNode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NarrativeNode, 0, len(d.narrativeNodes))
	for _, node := range d.narrativeNodes {
		result = append(result, node)
	}
	return result
}

// ListNarrativeNodesByProject 列出指定项目的叙事节点
func (d *MemoryDatabase) ListNarrativeNodesByProject(projectID string) []*models.NarrativeNode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NarrativeNode, 0)
	for _, node := range d.narrativeNodes {
		if node.ProjectID == projectID {
			result = append(result, node)
		}
	}
	return result
}

// ListNarrativeNodesByParent 列出指定父节点的子节点
func (d *MemoryDatabase) ListNarrativeNodesByParent(parentID string) []*models.NarrativeNode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NarrativeNode, 0)
	for _, node := range d.narrativeNodes {
		if node.ParentID != nil && *node.ParentID == parentID {
			result = append(result, node)
		}
	}
	return result
}

// DeleteNarrativeNode 删除叙事节点
func (d *MemoryDatabase) DeleteNarrativeNode(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.narrativeNodes[id]; !ok {
		return ErrNotFound
	}

	delete(d.narrativeNodes, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// ============================================
// NodeChapterMapping CRUD 操作
// ============================================

// SaveNodeChapterMapping 保存节点到章节的映射
func (d *MemoryDatabase) SaveNodeChapterMapping(mapping *models.NodeChapterMapping) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	mapping.UpdatedAt = time.Now()
	if mapping.CreatedAt.IsZero() {
		mapping.CreatedAt = time.Now()
	}

	d.nodeChapterMappings[mapping.ID] = mapping

	if d.autoSave {
		return d.save()
	}
	return nil
}

// ListNodeChapterMappingsByProject 列出指定项目的所有映射
func (d *MemoryDatabase) ListNodeChapterMappingsByProject(projectID string) []*models.NodeChapterMapping {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NodeChapterMapping, 0)
	for _, mapping := range d.nodeChapterMappings {
		if mapping.ProjectID == projectID {
			result = append(result, mapping)
		}
	}
	return result
}

// ListNodeChapterMappingsByChapter 列出指定章节的所有映射
func (d *MemoryDatabase) ListNodeChapterMappingsByChapter(chapterID string) []*models.NodeChapterMapping {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.NodeChapterMapping, 0)
	for _, mapping := range d.nodeChapterMappings {
		if mapping.ChapterID == chapterID {
			result = append(result, mapping)
		}
	}
	return result
}

// ============================================
// Chapter CRUD 操作
// ============================================

// SaveChapter 保存章节
func (d *MemoryDatabase) SaveChapter(chapter *models.Chapter) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	chapter.UpdatedAt = time.Now()
	if chapter.CreatedAt.IsZero() {
		chapter.CreatedAt = time.Now()
	}

	d.chapters[chapter.ID] = chapter

	if d.autoSave {
		return d.save()
	}
	return nil
}

// GetChapter 获取章节
func (d *MemoryDatabase) GetChapter(id string) (*models.Chapter, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	chapter, ok := d.chapters[id]
	if !ok {
		return nil, ErrNotFound
	}
	return chapter, nil
}

// GetChapterByNum 根据章节号获取章节
func (d *MemoryDatabase) GetChapterByNum(projectID string, chapterNum int) (*models.Chapter, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, ch := range d.chapters {
		if ch.ProjectID == projectID && ch.ChapterNum == chapterNum {
			return ch, nil
		}
	}
	return nil, ErrNotFound
}

// ListChaptersByProject 列出指定项目的所有章节
func (d *MemoryDatabase) ListChaptersByProject(projectID string) []*models.Chapter {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*models.Chapter, 0)
	for _, chapter := range d.chapters {
		if chapter.ProjectID == projectID {
			result = append(result, chapter)
		}
	}
	return result
}

// DeleteChapter 删除章节
func (d *MemoryDatabase) DeleteChapter(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.chapters[id]; !ok {
		return ErrNotFound
	}

	delete(d.chapters, id)

	if d.autoSave {
		return d.save()
	}
	return nil
}

// Now 返回当前时间
func Now() time.Time {
	return time.Now()
}

// SaveSynopsis 保存简介
func (d *MemoryDatabase) SaveSynopsis(synopsis *models.Synopsis) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.synopses[synopsis.ID] = synopsis
	return nil
}

// GetSynopsis 获取简介
func (d *MemoryDatabase) GetSynopsis(id string) (*models.Synopsis, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	synopsis, exists := d.synopses[id]
	if !exists {
		return nil, fmt.Errorf("synopsis not found")
	}
	return synopsis, nil
}
