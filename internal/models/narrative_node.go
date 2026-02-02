package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// 叙事节点相关
// ============================================

// NarrativeNode 叙事节点
type NarrativeNode struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	ProjectID   string         `json:"project_id" gorm:"not null;index"`
	WorldID     string         `json:"world_id" gorm:"not null;index"`
	ChapterID   *string        `json:"chapter_id,omitempty" gorm:"index"`

	// 层级结构
	ParentID    *string        `json:"parent_id,omitempty" gorm:"index"`
	NodeLevel   int            `json:"node_level" gorm:"default:0"`
	NodeOrder   int            `json:"node_order" gorm:"default:0"`

	// 节点类型和状态
	NodeType    NodeType       `json:"node_type" gorm:"size:20;not null"`
	NodeStatus  NodeStatus     `json:"node_status" gorm:"size:20;default:'draft'"`

	// 节点内容
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Content     string         `json:"content" gorm:"type:text"`

	// AI生成的元数据
	Metadata    NodeMetadata   `json:"metadata" gorm:"type:jsonb"`

	// 分支选项（AI生成的多个选项）
	Branches         []NodeBranch `json:"branches" gorm:"type:jsonb"`
	SelectedBranchID *string      `json:"selected_branch_id,omitempty"`

	// 世界设定快照
	WorldStages WorldStagesSnapshot `json:"world_stages" gorm:"type:jsonb"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeChapter NodeType = "chapter" // 章节点
	NodeTypeScene   NodeType = "scene"   // 场景节点
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusDraft     NodeStatus = "draft"      // 草稿
	NodeStatusGenerated NodeStatus = "generated"  // 已生成内容
	NodeStatusMerged    NodeStatus = "merged"     // 已合并到章节
	NodeStatusBranching NodeStatus = "branching"  // 分支生成中
)

// NodeMetadata 节点元数据
type NodeMetadata struct {
	Characters      []string   `json:"characters"`       // 涉及角色
	Locations       []string   `json:"locations"`        // 地点
	POVCharacter    string     `json:"pov_character"`    // 视角角色
	Mood            string     `json:"mood"`             // 氛围
	WordCount       int        `json:"word_count"`       // 字数
	PlotAdvancement string     `json:"plot_advancement"` // 情节推进
	ArcProgress     string     `json:"arc_progress"`     // 角色弧光进展
	AIModel         string     `json:"ai_model"`         // 使用的AI模型
	GeneratedAt     *time.Time `json:"generated_at,omitempty"`
}

// NodeBranch 节点分支（AI生成的选项）
type NodeBranch struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	ContentPreview  string        `json:"content_preview"` // 内容预览（200字）
	FullContent     string        `json:"full_content"`    // 完整内容
	BranchType      BranchType    `json:"branch_type"`      // 分支类型
	Rationale       string        `json:"rationale"`        // AI的解释：为什么生成这个分支
	ExpectedOutcome BranchOutcome `json:"expected_outcome"` // 预期结果
	CreatedAt       time.Time     `json:"created_at"`
}

// BranchType 分支类型
type BranchType string

const (
	BranchTypeContinuation           BranchType = "continuation"              // 直接延续
	BranchTypePlotTwist             BranchType = "plot_twist"                // 情节转折
	BranchTypeCharacterDevelopment  BranchType = "character_development"     // 角色发展
	BranchTypeConflictEscalation    BranchType = "conflict_escalation"      // 冲突升级
	BranchTypeConflictResolution    BranchType = "conflict_resolution"       // 冲突解决
	BranchTypeForeshadow            BranchType = "foreshadow"                // 伏笔铺垫
)

// BranchOutcome 分支预期结果
type BranchOutcome struct {
	PlotProgression   string   `json:"plot_progression"`   // 情节如何推进
	CharacterChanges  []string `json:"character_changes"`  // 角色变化
	NewConflicts      []string `json:"new_conflicts"`      // 新的冲突
	ResolvedConflicts []string `json:"resolved_conflicts"` // 解决的冲突
	ForeshadowHints   []string `json:"foreshadow_hints"`  // 伏笔提示
}

// WorldStagesSnapshot 世界设定快照（7个阶段）
type WorldStagesSnapshot struct {
	Philosophy   *Philosophy    `json:"philosophy,omitempty"`
	Worldview    *Worldview     `json:"worldview,omitempty"`
	Laws         *Laws          `json:"laws,omitempty"`
	Geography    *Geography     `json:"geography,omitempty"`
	Civilization *Civilization  `json:"civilization,omitempty"`
	Society      *Society        `json:"society,omitempty"`
	History      *History        `json:"history,omitempty"`
	StorySoil    *StorySoil      `json:"story_soil,omitempty"`
}

// BeforeCreate GORM hook
func (n *NarrativeNode) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// NodeChapterMapping 节点到章节的映射关系
type NodeChapterMapping struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	ProjectID     string        `json:"project_id" gorm:"not null;index"`
	ChapterID     string        `json:"chapter_id" gorm:"not null;index"`
	NodeID        string        `json:"node_id" gorm:"not null;index"`

	// 映射类型
	MappingType   MappingType   `json:"mapping_type" gorm:"size:20;not null"`

	// 顺序信息
	Sequence      int           `json:"sequence" gorm:"default:0"` // 在章节中的顺序

	// 合并方式
	MergeStrategy MergeStrategy `json:"merge_strategy"` // 合并策略

	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// MappingType 映射类型
type MappingType string

const (
	MappingTypeDirect  MappingType = "direct"  // 直接映射：一个节点=一个章节
	MappingTypeMerge   MappingType = "merge"   // 合并映射：多个节点=一个章节
	MappingTypeSplit   MappingType = "split"   // 拆分映射：一个节点拆分到多个章节
	MappingTypePartial MappingType = "partial" // 部分映射：节点内容的一部分
)

// MergeStrategy 合并策略
type MergeStrategy string

const (
	MergeStrategyAppend  MergeStrategy = "append"  // 追加：节点内容追加到章节末尾
	MergeStrategyPrepend MergeStrategy = "prepend" // 前置：节点内容插入到章节开头
	MergeStrategyReplace MergeStrategy = "replace" // 替换：节点内容替换章节
	MergeStrategyInsert  MergeStrategy = "insert"  // 插入：在指定位置插入
)

// BeforeCreate GORM hook
func (m *NodeChapterMapping) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
