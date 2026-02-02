package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// 项目相关
// ============================================

// Project 项目
type Project struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	UserID      string     `json:"user_id"`
	Mode        OrchestrationMode `json:"mode"`
	Status      ProjectStatus `json:"status"`
	Progress    float64    `json:"progress"` // 0-100
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联
	WorldID     string `json:"world_id"`
	NarrativeID string `json:"narrative_id"`
}

// OrchestrationMode 编排模式
type OrchestrationMode string

const (
	ModePlanning    OrchestrationMode = "planning"     // 规划生成
	ModeIntervention OrchestrationMode = "intervention" // 干预生成
	ModeRandom      OrchestrationMode = "random"       // 随机生成
	ModeStoryCore   OrchestrationMode = "story_core"   // 故事核
	ModeShort       OrchestrationMode = "short"        // 短篇模式
	ModeScript      OrchestrationMode = "script"       // 剧本模式
	ModeAssisted    OrchestrationMode = "assisted"    // 辅助创作
	ModeAutomatic   OrchestrationMode = "automatic"   // 自动创作
)

// ProjectStatus 项目状态
type ProjectStatus string

const (
	StatusDraft      ProjectStatus = "draft"       // 草稿
	StatusBuilding   ProjectStatus = "building"    // 构建中
	StatusGenerating ProjectStatus = "generating"  // 生成中
	StatusCompleted  ProjectStatus = "completed"   // 已完成
	StatusPaused     ProjectStatus = "paused"      // 已暂停
	StatusFailed     ProjectStatus = "failed"      // 失败
)

// ============================================
// 世界设定相关
// ============================================

// WorldSetting 世界设定
type WorldSetting struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Type        WorldType `json:"type"`
	Scale       WorldScale `json:"scale"`
	Style       string    `json:"style"` // 风格倾向
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 分层设定（存储为JSON）
	Philosophy   Philosophy   `json:"philosophy" gorm:"type:json"`
	Worldview   Worldview     `json:"worldview" gorm:"type:json"`
	Laws        Laws          `json:"laws" gorm:"type:json"`
	Geography   Geography     `json:"geography" gorm:"type:json"`
	Civilization Civilization `json:"civilization" gorm:"type:json"`
	Society     Society       `json:"society" gorm:"type:json"`
	History     History       `json:"history" gorm:"type:json"`

	// 故事土壤（叙事器最需要）
	StorySoil StorySoil `json:"story_soil" gorm:"type:json"`

	// 设定约束（写作器需要）
	SettingConstraints SettingConstraints `json:"setting_constraints" gorm:"type:json"`

	// 一致性检查报告（阶段7生成）
	ConsistencyReport *ConsistencyReport `json:"consistency_report,omitempty" gorm:"type:json"`
}

// WorldType 世界类型
type WorldType string

const (
	WorldFantasy     WorldType = "fantasy"      // 奇幻
	WorldScifi       WorldType = "scifi"        // 科幻
	WorldHistorical  WorldType = "historical"   // 历史
	WorldUrban       WorldType = "urban"        // 都市
	WorldWuxia       WorldType = "wuxia"        // 武侠
	WorldXianxia     WorldType = "xianxia"      // 仙侠
	WorldMixed       WorldType = "mixed"        // 混合
)

// WorldScale 世界规模
type WorldScale string

const (
	ScaleVillage   WorldScale = "village"   // 村庄级
	ScaleCity       WorldScale = "city"       // 城市级
	ScaleNation     WorldScale = "nation"     // 国家级
	ScaleContinent  WorldScale = "continent"  // 大陆级
	ScalePlanet     WorldScale = "planet"     // 星球级
	ScaleUniverse   WorldScale = "universe"   // 宇宙级
)

// ============================================
// 哲学思考层
// ============================================

// Philosophy 哲学思考
type Philosophy struct {
	CoreQuestion string      `json:"core_question"` // 探讨的根本问题
	Derivation  string      `json:"derivation"`     // 推导逻辑
	ValueSystem ValueSystem `json:"value_system"`
	Themes      []Theme     `json:"themes"`
}

// ValueSystem 价值体系
type ValueSystem struct {
	HighestGood   string `json:"highest_good"`    // 最高善
	UltimateEvil  string `json:"ultimate_evil"`    // 最大恶
	MoralDilemmas []Dilemma `json:"moral_dilemmas"`
}

// Dilemma 道德困境
type Dilemma struct {
	Dilemma     string `json:"dilemma"`
	Description string `json:"description"`
}

// Theme 主题
type Theme struct {
	Name             string `json:"name"`
	ExplorationAngle string `json:"exploration_angle"`
}

// ============================================
// 世界观层
// ============================================

// Worldview 世界观
type Worldview struct {
	Derivation string    `json:"derivation"` // 推导逻辑
	Cosmology  Cosmology `json:"cosmology"`
	Metaphysics Metaphysics `json:"metaphysics"`
}

// Cosmology 宇宙论
type Cosmology struct {
	Origin   string `json:"origin"`   // 世界起源
	Structure string `json:"structure"` // 世界结构
	Eschatology string `json:"eschatology"` // 终极命运
}

// Metaphysics 形而上学
type Metaphysics struct {
	SoulExists  bool   `json:"soul_exists"`
	SoulNature  string `json:"soul_nature,omitempty"`
	Afterlife   string `json:"afterlife,omitempty"`
	FateExists  bool   `json:"fate_exists"`
	FateRelShip string `json:"fate_relationship,omitempty"` // 命运与自由意志的关系
}

// ============================================
// 法则设定层
// ============================================

// Laws 法则设定
type Laws struct {
	Physics    Physics    `json:"physics"`
	Supernatural *Supernatural `json:"supernatural,omitempty"` // 可选
}

// Physics 物理法则
type Physics struct {
	Gravity        string `json:"gravity"`
	TimeFlow       string `json:"time_flow"`
	EnergyConservation string `json:"energy_conservation"`
	Causality      string `json:"causality"`
	DeathNature    string `json:"death_nature"`
}

// Supernatural 超自然体系
type Supernatural struct {
	Exists   bool              `json:"exists"`
	Type     string            `json:"type"` // magic, cultivation, superpower, etc.
	Settings *SupernaturalSettings `json:"settings,omitempty"`
}

// SupernaturalSettings 超自然设定
type SupernaturalSettings struct {
	// 魔法体系
	MagicSystem *MagicSystem `json:"magic_system,omitempty"`
	// 修真体系
	CultivationSystem *CultivationSystem `json:"cultivation_system,omitempty"`
	// 异能体系
	SuperpowerSystem *SuperpowerSystem `json:"superpower_system,omitempty"`
}

// MagicSystem 魔法体系
type MagicSystem struct {
	Source     string   `json:"source"`     // 魔法来源
	Cost       string   `json:"cost"`       // 使用代价
	Limitation []string `json:"limitation"` // 绝对限制
}

// CultivationSystem 修真体系
type CultivationSystem struct {
	Realms       []string `json:"realms"`        // 境界划分
	ResourceSystem string `json:"resource_system"` // 资源体系
	Bottleneck   string   `json:"bottleneck"`    // 瓶颈
}

// SuperpowerSystem 异能体系
type SuperpowerSystem struct {
	Origin   string   `json:"origin"`   // 能力起源
	Type     string   `json:"type"`     // 能力类型
	Limit    []string `json:"limit"`    // 限制条件
}

// ============================================
// 地理环境层
// ============================================

// Geography 地理环境
type Geography struct {
	Regions   []Region  `json:"regions"`
	Resources *Resources `json:"resources,omitempty"`
	Climate   *Climate   `json:"climate,omitempty"`
}

// Region 区域
type Region struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // mountain, plain, river, ocean, forest, desert
	Description string `json:"description"`
	Resources   []string `json:"resources"`
	Risks       []string `json:"risks"` // 自然灾害
}

// Resources 资源
type Resources struct {
	Basic    []string `json:"basic"`    // 基础资源
	Strategic []string `json:"strategic"` // 战略资源
	Rare     []string `json:"rare"`     // 稀有资源
}

// Climate 气候
type Climate struct {
	Type     string `json:"type"`
	Seasons  bool   `json:"seasons"`
	Features []string `json:"features"`
}

// ============================================
// 文明社会层
// ============================================

// Civilization 文明
type Civilization struct {
	Races     []Race       `json:"races"`
	Languages []Language   `json:"languages"`
	Religions []Religion   `json:"religions"`
	Values    *ValueSystem `json:"values,omitempty"` // 价值观
}

// Race 种族
type Race struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Traits      []string `json:"traits"`
	Abilities   []string `json:"abilities"`
	Relations   map[string]string `json:"relations"` // 与其他种族的关系
}

// Language 语言
type Language struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // natural, artificial, divine, ancient
	Speakers    string `json:"speakers"` // 使用者
	Features    []string `json:"features"`
}

// Religion 宗教
type Religion struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Cosmology   string       `json:"cosmology"`
	Ethics      []string     `json:"ethics"`
	Practices   []string     `json:"practices"`
	Organization *ReligionOrganization `json:"organization,omitempty"`
}

// ReligionOrganization 宗教组织
type ReligionOrganization struct {
	Type     string `json:"type"` // hierarchy, decentralized, cult
	Leader   string `json:"leader"`
	Factions []string `json:"factions"`
}

// ============================================
// 社会结构层
// ============================================

// Society 社会
type Society struct {
	Politics    Politics    `json:"politics"`
	Classes    []Class     `json:"classes"`
	Economy     Economy     `json:"economy"`
	Laws       []Law       `json:"laws"`
	Conflicts   []Conflict  `json:"conflicts"`
}

// Politics 政治结构
type Politics struct {
	Type              string `json:"type"` // monarchy, republic, theocracy, military, tribal
	PowerStructure    *PowerStructure `json:"power_structure,omitempty"`
	LegitimacySource  string `json:"legitimacy_source"`
}

// PowerStructure 权力结构
type PowerStructure struct {
	Formal    []PowerLevel `json:"formal"`    // 明面权力
	Actual    []PowerHolder `json:"actual"`   // 实际掌权者
	ChecksAndBalances string `json:"checks_and_balances"`
}

// PowerLevel 权力层级
type PowerLevel struct {
	Level   string `json:"level"`
	Name    string `json:"name"`
	Powers  []string `json:"powers"`
}

// PowerHolder 实际掌权者
type PowerHolder struct {
	Entity        string `json:"entity"` // 实体名称
	PowerSource   string `json:"power_source"`
	Relationship  string `json:"relationship"`
}

// Class 阶级
type Class struct {
	Name        string `json:"name"`
	Rank        int    `json:"rank"` // 社会等级 0-100
	Rights      []string `json:"rights"`
	Obligations []string `json:"obligations"`
}

// Economy 经济
type Economy struct {
	Type         string   `json:"type"` // natural, commodity, capitalist, planned
	TradeNetwork string   `json:"trade_network"`
	Currency     []string `json:"currency"`
}

// Law 法律
type Law struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Penalty     string `json:"penalty"`
}

// Conflict 冲突
type Conflict struct {
	Type        string   `json:"type"` // economic, political, social, cultural
	Description string   `json:"description"`
	Parties     []string `json:"parties"`
	Tension     int      `json:"tension"` // 0-100
	Triggers    []string `json:"triggers"` // 触发条件
}

// ============================================
// 历史层
// ============================================

// History 历史
type History struct {
	Origin   string   `json:"origin"`
	Eras     []Era    `json:"eras"`
	Events   []Event  `json:"events"`
	Traumas  []string `json:"traumas"` // 集体创伤
	Legacies []string `json:"legacies"` // 历史遗产
}

// Era 时代
type Era struct {
	Name        string `json:"name"`
	Period      string `json:"period"` // 起止时间
	Type        string `json:"type"` // origin, expansion, peak, decline
	Description string `json:"description"`
}

// Event 历史事件
type Event struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Time        string `json:"time"`
	Description string `json:"description"`
	Causes      []string `json:"causes"`     // 深层原因
	Consequences []string `json:"consequences"` // 后果
	Impact      string `json:"impact"`
}

// ============================================
// 故事土壤
// ============================================

// StorySoil 故事土壤
type StorySoil struct {
	SocialConflicts     []Conflict            `json:"social_conflicts"`
	PowerStructures     []PowerStructure       `json:"power_structures"`
	HistoricalContext   HistoricalContext     `json:"historical_context"`
	PotentialPlotHooks  []PlotHook           `json:"potential_plot_hooks"`
	CulturalDetails     CulturalDetails       `json:"cultural_details"`
}

// HistoricalContext 历史语境
type HistoricalContext struct {
	RecentEvents []RecentEvent `json:"recent_events"`
	CollectiveMemory string      `json:"collective_memory"`
	UnresolvedIssues []string   `json:"unresolved_issues"`
}

// RecentEvent 近期事件
type RecentEvent struct {
	Event   string `json:"event"`
	Impact  string `json:"impact"`
	YearsAgo int    `json:"years_ago"`
}

// PlotHook 情节钩子
type PlotHook struct {
	Type           string `json:"type"` // power_vacuum, conflict, mystery
	Description    string `json:"description"`
	StoryPotential string `json:"story_potential"`
	Triggers       []string `json:"triggers"`
}

// CulturalDetails 文化细节
type CulturalDetails struct {
	Customs  []string `json:"customs"`
	Taboos   []string `json:"taboos"`
	Slang    []string `json:"slang"`
	Holidays []string `json:"holidays"`
	Arts     []string `json:"arts"`
}

// SettingConstraints 设定约束
type SettingConstraints struct {
	MagicSystem  *MagicSystem  `json:"magic_system,omitempty"`
	TechnologyLevel string `json:"technology_level"`
	GeographySummary string `json:"geography_summary"`
}

// ============================================
// 角色相关
// ============================================

// Character 角色
type Character struct {
	ID             string           `json:"id" gorm:"primaryKey"`
	WorldID        string           `json:"world_id"`
	Name           string           `json:"name"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`

	// 静态档案（世界设定器生成）
	StaticProfile  StaticProfile    `json:"static_profile" gorm:"type:json"`

	// 叙事档案（叙事器生成）
	NarrativeProfile NarrativeProfile `json:"narrative_profile" gorm:"type:json"`

	// 动态状态（写作器维护）
	DynamicState   DynamicState     `json:"dynamic_state" gorm:"type:json"`
}

// StaticProfile 静态档案
type StaticProfile struct {
	Background   string   `json:"background"`
	Race         string   `json:"race"`
	Age          int      `json:"age"`
	Gender       string   `json:"gender"`
	Appearance   string   `json:"appearance"`
	Abilities    []string `json:"abilities"`
	SocialStatus string   `json:"social_status"`
	Occupation   string   `json:"occupation"`
}

// NarrativeProfile 叙事档案
type NarrativeProfile struct {
	Personality    []Trait  `json:"personality"`
	Motivation     Motivation `json:"motivation"`
	Flaw           string    `json:"flaw"` // 致命缺陷
	Fear           string    `json:"fear"` // 核心恐惧
	BeliefSystem   BeliefSystem `json:"belief_system"`
	ArcPlan        *ArcPlan  `json:"arc_plan,omitempty"`
	Relationships  map[string]*Relationship `json:"relationships"`
}

// Trait 特质
type Trait struct {
	Name        string `json:"name"`
	Category    string `json:"category"` // positive, negative, neutral
	Intensity   int    `json:"intensity"` // 0-100
}

// Motivation 动机
type Motivation struct {
	CoreNeed      string `json:"core_need"`      // 核心需求
	ExternalGoal  string `json:"external_goal"`  // 外在目标
	InnerConflict  string `json:"inner_conflict"` // 内在冲突
}

// BeliefSystem 信念系统
type BeliefSystem struct {
	WorldView   string   `json:"world_view"`   // 世界观
	HumanNature string   `json:"human_nature"` // 人性观
	Morality    []string `json:"morality"`    // 道德观念
}

// ArcPlan 弧光规划
type ArcPlan struct {
	ArcType      string       `json:"arc_type"` // growth, negative, flat, complex
	StartState   CharacterState `json:"start_state"`
	EndState     CharacterState `json:"end_state"`
	TurningPoints []TurningPoint `json:"turning_points"`
	CurrentProgress int       `json:"current_progress"` // 0-100
}

// CharacterState 角色状态
type CharacterState struct {
	Personality  []string `json:"personality"`
	Motivation   string   `json:"motivation"`
	Emotion      string   `json:"emotion"`
}

// TurningPoint 转折点
type TurningPoint struct {
	Chapter      int    `json:"chapter"`
	Event       string `json:"event"`
	Change      string `json:"change"`
}

// Relationship 关系
type Relationship struct {
	CharacterID  string  `json:"character_id"`
	Emotion      int     `json:"emotion"`      // -100 到 100
	Power        string  `json:"power"`        // superior, equal, inferior
	Secrets      []string `json:"secrets"`      // 保密内容
	Attitude     string  `json:"attitude"`     // 表面态度
	TrustLevel   int     `json:"trust_level"` // 0-100
}

// DynamicState 动态状态
type DynamicState struct {
	Location     string       `json:"location"`
	Timestamp    time.Time    `json:"timestamp"`
	Physiology   Physiology  `json:"physiology"`
	Emotion      Emotion      `json:"emotion"`
	Psychology   Psychology   `json:"psychology"`
	Knowledge    Knowledge    `json:"knowledge"`
	ArcProgress  int          `json:"arc_progress"` // 0-100
}

// Physiology 生理状态
type Physiology struct {
	Health    string `json:"health"`    // healthy, injured, sick, critical
	Energy    string `json:"energy"`    // energized, normal, tired, exhausted
	Condition  []string `json:"condition"` // 其他状态
}

// Emotion 情绪状态
type Emotion struct {
	Current   string `json:"current"`    // 当前主情绪
	Intensity int    `json:"intensity"`  // 0-100
	Trigger   string `json:"trigger"`    // 触发源
	Mood      string `json:"mood"`       // 心境
}

// Psychology 心理状态
type Psychology struct {
	Focus      string `json:"focus"`      // high, medium, low
	Stress     int    `json:"stress"`     // 0-100
	Stability  string `json:"stability"`  // stable, shaken, breaking
}

// Knowledge 知识状态
type Knowledge struct {
	Known    []string `json:"known"`     // 已知信息
	Unknown  []string `json:"unknown"`   // 未知信息
	Mistaken []string `json:"mistaken"`  // 错误信息
}

// ============================================
// 叙事蓝图相关
// ============================================

// NarrativeBlueprint 叙事蓝图
type NarrativeBlueprint struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	WorldID       string         `json:"world_id"`
	ProjectID     string         `json:"project_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	// 核心内容
	StoryOutline  StoryOutline   `json:"story_outline" gorm:"type:json"`
	ChapterPlans  []ChapterPlan   `json:"chapter_plans"`
	Scenes        []SceneInstruction `json:"scenes"`
	CharacterArcs map[string]*ArcPlan `json:"character_arcs" gorm:"type:json"`
	ThemePlan     ThemePlan      `json:"theme_plan" gorm:"type:json"`
}

// StoryOutline 故事大纲
type StoryOutline struct {
	StructureType string    `json:"structure_type"` // three_act, heros_journey, kishotenketsu
	Act1          Act1      `json:"act1"`
	Act2          Act2      `json:"act2"`
	Act3          Act3      `json:"act3"`
}

// Act1 第一幕
type Act1 struct {
	Setup          string `json:"setup"`
	IncitingIncident string `json:"inciting_incident"`
	PlotPoint1     string `json:"plot_point1"`
}

// Act2 第二幕
type Act2 struct {
	RisingAction []string `json:"rising_action"`
	Midpoint      string   `json:"midpoint"`
	AllIsLost     string   `json:"all_is_lost"`
	PlotPoint2    string   `json:"plot_point2"`
}

// Act3 第三幕
type Act3 struct {
	Climax     string `json:"climax"`
	Resolution string `json:"resolution"`
}

// ChapterPlan 章节规划
type ChapterPlan struct {
	Chapter       int      `json:"chapter"`
	Title        string   `json:"title"`
	Purpose       string   `json:"purpose"`
	KeyScenes     []string `json:"key_scenes"`
	PlotAdvancement string  `json:"plot_advancement"`
	ArcProgress   string   `json:"arc_progress"`
	EndingHook    string   `json:"ending_hook"`
	WordCount     int      `json:"word_count"`
	Status        string   `json:"status"` // pending, generating, completed
}

// SceneInstruction 场景指令
type SceneInstruction struct {
	Chapter        int      `json:"chapter"`
	Scene          int      `json:"scene"`
	Sequence       int      `json:"sequence"`
	Purpose        string   `json:"purpose"`
	Location       string   `json:"location"`
	Characters     []string `json:"characters"`
	POVCharacter   string   `json:"pov_character"` // 视角角色
	Action         string   `json:"action"`
	DialogueFocus  string   `json:"dialogue_focus"`
	ExpectedLength int     `json:"expected_length"` // 字数
	Mood           string   `json:"mood"` // 氛围要求
	Status         string   `json:"status"` // pending, generating, completed
}

// ThemePlan 主题规划
type ThemePlan struct {
	CoreTheme    string            `json:"core_theme"`
	Threading     []ThemeThreading `json:"threading"`
	Symbols      []Symbol         `json:"symbols"`
	Motifs       []string         `json:"motifs"`
}

// ThemeThreading 主题贯穿
type ThemeThreading struct {
	Chapter   int    `json:"chapter"`
	Expression string `json:"expression"`
	Depth     string `json:"depth"` // surface, deep, philosophical
}

// Symbol 象征
type Symbol struct {
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Appearances []int  `json:"appearances"` // 出现章节
}

// ============================================
// 场景输出相关
// ============================================

// SceneOutput 场景输出
type SceneOutput struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	BlueprintID   string         `json:"blueprint_id"`
	Chapter       int            `json:"chapter"`
	Scene         int            `json:"scene"`
	Content       string         `json:"content"` // 生成的文本
	WordCount     int            `json:"word_count"`
	CreatedAt     time.Time      `json:"created_at"`

	// 元数据
	POVCharacter   string         `json:"pov_character"`
	Tone            string         `json:"tone"`
	Style           string         `json:"style"`

	// 状态更新
	StateUpdates   StateUpdates   `json:"state_updates" gorm:"type:json"`
}

// StateUpdates 状态更新
type StateUpdates struct {
	Characters    []CharacterUpdate `json:"characters"`
	World         *WorldUpdate      `json:"world,omitempty"`
	PlotProgress  string            `json:"plot_progress"`
}

// CharacterUpdate 角色更新
type CharacterUpdate struct {
	ID            string        `json:"id"`
	Location      string        `json:"location"`
	EmotionChange  EmotionChange `json:"emotion_change"`
	KnowledgeGain  []string      `json:"knowledge_gain"`
	RelationshipChanges map[string]int `json:"relationship_changes"`
}

// EmotionChange 情绪变化
type EmotionChange struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Reason string `json:"reason"`
}

// WorldUpdate 世界状态更新
type WorldUpdate struct {
	TimeChange   string `json:"time_change"`
	EventAdded   string `json:"event_added"`
	StateChanged string `json:"state_changed"`
}

// ============================================
// 用户相关
// ============================================

// User 用户
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"size:100;uniqueIndex;not null"`
	PasswordHash string  `json:"-" gorm:"size:255;not null"` // 不在JSON中输出
	APIKey    *string   `json:"-" gorm:"size:255;uniqueIndex"` // 不在JSON中输出
	Phone        string    `json:"phone,omitempty" gorm:"size:20"`
	AvatarURL    string    `json:"avatar_url,omitempty" gorm:"size:500"`

	// 用户等级
	Tier       string     `json:"tier" gorm:"size:20;default:'free';not null"` // free, vip, svip, admin
	TierExpire *time.Time `json:"tier_expires_at,omitempty"`

	// 账户状态
	Status         string     `json:"status" gorm:"size:20;default:'active';not null"` // active, suspended, deleted
	EmailVerified  bool       `json:"email_verified" gorm:"default:false"`
	PhoneVerified  bool       `json:"phone_verified" gorm:"default:false"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 配置
	Settings UserSettings `json:"settings" gorm:"type:jsonb;serializer:json"`
}

// BeforeCreate GORM hook - 创建前生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = generateUUID()
	}
	return nil
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Tier == "svip" || u.Tier == "admin"
}

// IsVip 检查是否是VIP用户
func (u *User) IsVip() bool {
	return u.Tier == "vip" || u.Tier == "svip" || u.Tier == "admin"
}

// HasValidTier 检查用户等级是否有效
func (u *User) HasValidTier() bool {
	if u.Tier == "free" || u.Tier == "admin" {
		return true
	}
	if u.TierExpire == nil {
		return false
	}
	return u.TierExpire.After(time.Now())
}

// generateUUID 生成UUID字符串
func generateUUID() string {
	return uuid.New().String()
}

// UserSettings 用户设置
type UserSettings struct {
	DefaultLLM   string `json:"default_llm"`
	MaxTokens    int    `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	AutoSave     bool   `json:"auto_save"`
}

// AuthToken 认证Token模型
type AuthToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"-" gorm:"size:500;uniqueIndex;not null"` // 不在JSON中输出
	TokenType string    `json:"token_type" gorm:"size:20;not null"` // access, refresh, reset
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================
// API 请求/响应
// ============================================

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Mode        string `json:"mode" binding:"required"`
	Params      CreateParams `json:"params"`
}

// CreateParams 创建参数
type CreateParams struct {
	// 世界参数
	WorldType   string `json:"world_type"`
	WorldTheme  string `json:"world_theme"`
	WorldScale  string `json:"world_scale"`
	WorldStyle  string `json:"world_style"`

	// 故事参数
	StoryType   string `json:"story_type"`
	Protagonist  string `json:"protagonist"`
	Length      string `json:"length"` // short, medium, long

	// 特殊要求
	SpecialRequirements []string `json:"special_requirements"`
}

// GenerateChapterRequest 生成章节请求
type GenerateChapterRequest struct {
	BlueprintID string `json:"blueprint_id" binding:"required"`
	Chapter     int    `json:"chapter" binding:"required"`
	Regenerate   bool   `json:"regenerate"`
}

// InterveneRequest 干预请求
type InterveneRequest struct {
	ProjectID   string `json:"project_id" binding:"required"`
	Type        string `json:"type" binding:"required"` // world, narrative, chapter
	Content     string `json:"content" binding:"required"`
}

// ============================================
// 响应结构
// ============================================

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Progress  float64 `json:"progress"`
	WorldID   string `json:"world_id"`
	NarrativeID string `json:"narrative_id"`
	CreatedAt string `json:"created_at"`
}

// WorldResponse 世界设定响应
type WorldResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	StorySoil   StorySoil `json:"story_soil"`
	CreatedAt   string `json:"created_at"`
}

// NarrativeResponse 叙事蓝图响应
type NarrativeResponse struct {
	ID             string         `json:"id"`
	StoryOutline   StoryOutline   `json:"story_outline"`
	ChapterCount   int            `json:"chapter_count"`
	CharacterArcs  map[string]*ArcPlan `json:"character_arcs"`
}

// ChapterResponse 章节响应
type ChapterResponse struct {
	Chapter   int           `json:"chapter"`
	Title     string        `json:"title"`
	Scenes    []SceneOutput `json:"scenes"`
	WordCount int           `json:"word_count"`
	Status    string        `json:"status"`
}

// ============================================
// 统计相关
// ============================================

// TokenUsage Token使用统计
type TokenUsage struct {
	TotalTokens   int `json:"total_tokens"`
	InputTokens   int `json:"input_tokens"`
	OutputTokens  int `json:"output_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
}

// GenerationStats 生成统计
type GenerationStats struct {
	ProjectID     string     `json:"project_id"`
	TokensUsed    int        `json:"tokens_used"`
	Cost          float64    `json:"cost"`
	ChaptersGenerated int     `json:"chapters_generated"`
	WordsGenerated int       `json:"words_generated"`
	Duration      string     `json:"duration"`
}

// ============================================
// 一致性检查相关（世界设定器阶段7）
// ============================================

// ConsistencyReport 一致性检查报告
type ConsistencyReport struct {
	OverallScore   int                `json:"overall_score"`   // 0-100
	Issues         []ConsistencyIssue `json:"issues"`
	Strengths      []string           `json:"strengths"`
	Improvements   []string           `json:"improvements"`
	StoryPotential StoryPotential     `json:"story_potential"`
}

// ConsistencyIssue 一致性问题
type ConsistencyIssue struct {
	Aspect     string `json:"aspect"`     // 问题所属方面
	Issue      string `json:"issue"`      // 问题描述
	Severity   string `json:"severity"`   // low/medium/high
	Suggestion string `json:"suggestion"` // 修复建议
}

// StoryPotential 故事潜力评估
type StoryPotential struct {
	Score                 int      `json:"score"`                 // 0-100
	HighPotentialElements []string `json:"high_potential_elements"` // 高潜力元素
	UnderutilizedElements []string `json:"underutilized_elements"` // 未充分利用元素
}

// ============================================
// 章节相关
// ============================================

// Chapter 章节
type Chapter struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	ProjectID    string     `json:"project_id" gorm:"not null;index"`
	ChapterNum   int        `json:"chapter_num" gorm:"not null"`
	Title        string     `json:"title" gorm:"size:200;not null"`
	Content      string     `json:"content" gorm:"type:text"`
	WordCount    int        `json:"word_count" gorm:"default:0"`
	AIWordCount  int        `json:"ai_generated_word_count" gorm:"default:0"`
	Status       ChapterStatus `json:"status" gorm:"size:20;default:'draft'"`
	GeneratedAt  *time.Time `json:"generated_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ChapterStatus 章节状态
type ChapterStatus string

const (
	ChapterStatusDraft     ChapterStatus = "draft"     // 草稿
	ChapterStatusCompleted ChapterStatus = "completed" // 已完成
)

// BeforeCreate GORM hook - 创建前生成UUID
func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateUUID()
	}
	return nil
}

// Synopsis 作品简介/大纲
type Synopsis struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ProjectID   string    `json:"project_id" gorm:"not null;index"`
	WorldID     string    `json:"world_id" gorm:"not null;index"`

	// 核心故事要素
	OneLineSummary string `json:"one_line_summary"` // 一句话简介
	ShortSummary   string `json:"short_summary"`    // 简短简介（100-200字）
	DetailedSummary string `json:"detailed_summary"` // 详细简介

	// 故事大纲
	MainPlot      string   `json:"main_plot"`      // 主线情节
	SubPlots      []string `json:"sub_plots"`      // 支线情节
	CoreConflict  string   `json:"core_conflict"`  // 核心冲突
	Resolution    string   `json:"resolution"`     // 解决方案

	// 故事结构
	StructureType string   `json:"structure_type"` // 结构类型（三幕式、英雄之旅等）
	KeyEvents      []string `json:"key_events"`     // 关键事件列表
	Themes         []string `json:"themes"`         // 主题标签

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
