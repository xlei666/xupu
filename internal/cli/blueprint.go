// Package cli CLI命令实现 - 蓝图和导出
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/narrative"
)

// 确保db包被导入（用于类型）
var _ db.Database

// NewBlueprintCommand 创建蓝图命令组
func NewBlueprintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blueprint",
		Short: "叙事蓝图管理",
	}

	cmd.AddCommand(newBlueprintListCmd())
	cmd.AddCommand(newBlueprintCreateCmd())
	cmd.AddCommand(newBlueprintShowCmd())

	return cmd
}

// newBlueprintListCmd 列出所有蓝图
func newBlueprintListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有叙事蓝图",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			blueprints := database.ListBlueprints()

			PrintHeader("叙事蓝图列表")

			if len(blueprints) == 0 {
				PrintInfo("暂无叙事蓝图")
				return
			}

			rows := make([][]string, 0, len(blueprints))
			for _, b := range blueprints {
				rows = append(rows, []string{
					b.ID[:12],
					b.WorldID[:12],
					b.StoryOutline.StructureType,
					fmt.Sprintf("%d章", len(b.ChapterPlans)),
					fmt.Sprintf("%d场", len(b.Scenes)),
				})
			}

			PrintTable([]string{"ID", "世界ID", "结构", "章节", "场景"}, rows)
			fmt.Println()
			PrintInfo("共 %d 个蓝图", len(blueprints))
		},
	}
}

// newBlueprintCreateCmd 创建新蓝图
func newBlueprintCreateCmd() *cobra.Command {
	var (
		worldID     string
		storyType   string
		theme       string
		protagonist string
		length      string
		chapters    int
		structure   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建叙事蓝图",
		Run: func(cmd *cobra.Command, args []string) {
			if worldID == "" {
				PrintError("请指定世界ID (--world-id)")
				return
			}

			PrintHeader("创建叙事蓝图")

			database := GetDBOrExit()
			// 验证世界存在
			if _, err := database.GetWorld(worldID); err != nil {
				PrintError("世界不存在: %s", worldID)
				return
			}

			// 创建叙事引擎
			engine, err := narrative.New()
			if err != nil {
				PrintError("初始化叙事引擎失败: %v", err)
				return
			}

			// 构建参数
			params := narrative.CreateParams{
				WorldID:      worldID,
				StoryType:    storyType,
				Theme:        theme,
				Protagonist:  protagonist,
				Length:       length,
				ChapterCount: chapters,
				Structure:    parseNarrativeStructure(structure),
			}

			PrintInfo("正在生成叙事蓝图...")

			// 创建蓝图
			blueprint, err := engine.CreateBlueprint(params)
			if err != nil {
				PrintError("创建蓝图失败: %v", err)
				return
			}

			// 保存蓝图
			if err := database.SaveNarrativeBlueprint(blueprint); err != nil {
				PrintError("保存蓝图失败: %v", err)
				return
			}

			PrintSuccess("叙事蓝图创建成功!")
			fmt.Println()
			printBlueprintDetail(blueprint)
		},
	}

	cmd.Flags().StringVar(&worldID, "world-id", "", "世界ID")
	cmd.Flags().StringVar(&storyType, "story-type", "adventure", "故事类型")
	cmd.Flags().StringVar(&theme, "theme", "", "故事主题")
	cmd.Flags().StringVar(&protagonist, "protagonist", "", "主角设定")
	cmd.Flags().StringVar(&length, "length", "medium", "故事长度 (short/medium/long)")
	cmd.Flags().IntVar(&chapters, "chapters", 12, "章节数量")
	cmd.Flags().StringVar(&structure, "structure", "three_act", "叙事结构 (three_act/heros_journey/save_the_cat)")

	return cmd
}

// newBlueprintShowCmd 查看蓝图详情
func newBlueprintShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "查看蓝图详情",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			blueprint, err := database.GetNarrativeBlueprint(args[0])
			if err != nil {
				PrintError("蓝图不存在: %s", args[0])
				return
			}

			PrintHeader("叙事蓝图详情")
			printBlueprintDetail(blueprint)
		},
	}
}

// printBlueprintDetail 打印蓝图详情
func printBlueprintDetail(blueprint *models.NarrativeBlueprint) {
	fmt.Printf("ID:       %s\n", blueprint.ID)
	fmt.Printf("世界ID:   %s\n", blueprint.WorldID)
	fmt.Printf("结构类型: %s\n", blueprint.StoryOutline.StructureType)
	fmt.Printf("章节数量: %d\n", len(blueprint.ChapterPlans))
	fmt.Printf("场景数量: %d\n", len(blueprint.Scenes))
	fmt.Println()

	// 故事大纲
	PrintSection("故事大纲")
	if blueprint.StoryOutline.Act1.Setup != "" {
		fmt.Printf("【第一幕】\n")
		fmt.Printf("铺垫: %s\n", blueprint.StoryOutline.Act1.Setup)
		fmt.Printf("激励事件: %s\n", blueprint.StoryOutline.Act1.IncitingIncident)
		fmt.Printf("情节点1: %s\n", blueprint.StoryOutline.Act1.PlotPoint1)
	}
	fmt.Println()

	// 核心主题
	if blueprint.ThemePlan.CoreTheme != "" {
		PrintSection("核心主题")
		fmt.Println(blueprint.ThemePlan.CoreTheme)
		fmt.Println()
	}

	// 章节规划
	if len(blueprint.ChapterPlans) > 0 {
		PrintSection("章节规划")
		for _, ch := range blueprint.ChapterPlans {
			fmt.Printf("第%d章: %s\n", ch.Chapter, ch.Title)
			if ch.Purpose != "" {
				fmt.Printf("  目的: %s\n", ch.Purpose)
			}
		}
		fmt.Println()
	}
}

// NewExportCommand 创建导出命令组
func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "导出功能",
	}

	cmd.AddCommand(newExportProjectCmd())
	cmd.AddCommand(newExportWorldCmd())
	cmd.AddCommand(newExportBlueprintCmd())

	return cmd
}

// newExportProjectCmd 导出项目
func newExportProjectCmd() *cobra.Command {
	var format string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "project <id>",
		Short: "导出项目",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			project, err := database.GetProject(args[0])
			if err != nil {
				PrintError("项目不存在: %s", args[0])
				return
			}

			// 确定输出文件
			if outputFile == "" {
				outputFile = fmt.Sprintf("%s.%s", project.Name, format)
			}

			// 导出
			if format == "markdown" || format == "md" {
				exportProjectMarkdown(project, outputFile)
			} else {
				PrintError("不支持的格式: %s", format)
				return
			}

			PrintSuccess("已导出到: %s", outputFile)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "导出格式 (markdown/txt)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径")

	return cmd
}

// newExportWorldCmd 导出世界设定
func newExportWorldCmd() *cobra.Command {
	var format string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "world <id>",
		Short: "导出世界设定",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			world, err := database.GetWorld(args[0])
			if err != nil {
				PrintError("世界不存在: %s", args[0])
				return
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("world_%s.%s", world.ID[:8], format)
			}

			if format == "markdown" || format == "md" {
				exportWorldMarkdown(world, outputFile)
			} else {
				PrintError("不支持的格式: %s", format)
				return
			}

			PrintSuccess("已导出到: %s", outputFile)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "导出格式 (markdown/txt)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径")

	return cmd
}

// newExportBlueprintCmd 导出蓝图
func newExportBlueprintCmd() *cobra.Command {
	var format string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "blueprint <id>",
		Short: "导出叙事蓝图",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			blueprint, err := database.GetNarrativeBlueprint(args[0])
			if err != nil {
				PrintError("蓝图不存在: %s", args[0])
				return
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("blueprint_%s.%s", blueprint.ID[:8], format)
			}

			if format == "markdown" || format == "md" {
				exportBlueprintMarkdown(blueprint, outputFile)
			} else {
				PrintError("不支持的格式: %s", format)
				return
			}

			PrintSuccess("已导出到: %s", outputFile)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "导出格式 (markdown/txt)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径")

	return cmd
}

// NewGenerateCommand 创建生成命令
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "内容生成",
	}

	cmd.AddCommand(newGenerateChapterCmd())

	return cmd
}

func newGenerateChapterCmd() *cobra.Command {
	var blueprintID string
	var chapter int

	cmd := &cobra.Command{
		Use:   "chapter",
		Short: "生成章节内容",
		Run: func(cmd *cobra.Command, args []string) {
			if blueprintID == "" {
				PrintError("请指定蓝图ID (--blueprint-id)")
				return
			}

			PrintHeader("生成章节内容")
			PrintInfo("蓝图ID: %s", blueprintID)
			PrintInfo("章节: %d", chapter)

			PrintWarn("章节生成功能开发中")
		},
	}

	cmd.Flags().StringVar(&blueprintID, "blueprint-id", "", "蓝图ID")
	cmd.Flags().IntVarP(&chapter, "chapter", "c", 1, "章节号")

	return cmd
}

// ============================================
// 配置命令
// ============================================

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "初始化配置",
		Run: func(cmd *cobra.Command, args []string) {
			PrintInfo("初始化配置文件...")
			// TODO: 实现配置初始化
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "显示当前配置",
		Run: func(cmd *cobra.Command, args []string) {
			PrintHeader("当前配置")
			// TODO: 实现配置显示
		},
	})

	return cmd
}

// ============================================
// 导出辅助函数
// ============================================

func exportProjectMarkdown(project interface{}, outputFile string) {
	f, err := os.Create(outputFile)
	if err != nil {
		PrintError("创建文件失败: %v", err)
		return
	}
	defer f.Close()

	// TODO: 实现完整的导出逻辑
	f.WriteString(fmt.Sprintf("# %s\n\n", "项目导出"))
}

func exportWorldMarkdown(world *models.WorldSetting, outputFile string) {
	f, err := os.Create(outputFile)
	if err != nil {
		PrintError("创建文件失败: %v", err)
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("# 世界设定: %s\n\n", world.Name))
	f.WriteString(fmt.Sprintf("**类型**: %s\n", FormatWorldType(world.Type)))
	f.WriteString(fmt.Sprintf("**规模**: %s\n\n", FormatWorldScale(world.Scale)))
}

func exportBlueprintMarkdown(blueprint *models.NarrativeBlueprint, outputFile string) {
	f, err := os.Create(outputFile)
	if err != nil {
		PrintError("创建文件失败: %v", err)
		return
	}
	defer f.Close()

	// 标题和基本信息
	f.WriteString(fmt.Sprintf("# 叙事蓝图\n\n"))
	f.WriteString(fmt.Sprintf("**ID**: %s\n", blueprint.ID))
	f.WriteString(fmt.Sprintf("**世界ID**: %s\n", blueprint.WorldID))
	f.WriteString(fmt.Sprintf("**结构**: %s\n", blueprint.StoryOutline.StructureType))
	f.WriteString(fmt.Sprintf("**章节数量**: %d\n", len(blueprint.ChapterPlans)))
	f.WriteString(fmt.Sprintf("**场景数量**: %d\n\n", len(blueprint.Scenes)))

	// 核心主题
	if blueprint.ThemePlan.CoreTheme != "" {
		f.WriteString("## 核心主题\n\n")
		f.WriteString(fmt.Sprintf("%s\n\n", blueprint.ThemePlan.CoreTheme))
	}

	// 故事大纲
	f.WriteString("## 故事大纲\n\n")
	if blueprint.StoryOutline.Act1.Setup != "" {
		f.WriteString("### 第一幕\n\n")
		if blueprint.StoryOutline.Act1.Setup != "" {
			f.WriteString(fmt.Sprintf("**铺垫**: %s\n\n", blueprint.StoryOutline.Act1.Setup))
		}
		if blueprint.StoryOutline.Act1.IncitingIncident != "" {
			f.WriteString(fmt.Sprintf("**激励事件**: %s\n\n", blueprint.StoryOutline.Act1.IncitingIncident))
		}
		if blueprint.StoryOutline.Act1.PlotPoint1 != "" {
			f.WriteString(fmt.Sprintf("**情节点1**: %s\n\n", blueprint.StoryOutline.Act1.PlotPoint1))
		}
	}

	if blueprint.StoryOutline.Act2.Midpoint != "" {
		f.WriteString("### 第二幕\n\n")
		if len(blueprint.StoryOutline.Act2.RisingAction) > 0 {
			f.WriteString("**上升动作**:\n")
			for _, action := range blueprint.StoryOutline.Act2.RisingAction {
				f.WriteString(fmt.Sprintf("- %s\n", action))
			}
			f.WriteString("\n")
		}
		f.WriteString(fmt.Sprintf("**中点**: %s\n\n", blueprint.StoryOutline.Act2.Midpoint))
		f.WriteString(fmt.Sprintf("**一无所有**: %s\n\n", blueprint.StoryOutline.Act2.AllIsLost))
		f.WriteString(fmt.Sprintf("**情节点2**: %s\n\n", blueprint.StoryOutline.Act2.PlotPoint2))
	}

	if blueprint.StoryOutline.Act3.Climax != "" {
		f.WriteString("### 第三幕\n\n")
		f.WriteString(fmt.Sprintf("**高潮**: %s\n\n", blueprint.StoryOutline.Act3.Climax))
		f.WriteString(fmt.Sprintf("**结局**: %s\n\n", blueprint.StoryOutline.Act3.Resolution))
	}

	// 章节规划
	if len(blueprint.ChapterPlans) > 0 {
		f.WriteString("## 章节规划\n\n")
		for _, ch := range blueprint.ChapterPlans {
			f.WriteString(fmt.Sprintf("### %s\n\n", ch.Title))
			f.WriteString(fmt.Sprintf("**章节**: %d\n", ch.Chapter))
			f.WriteString(fmt.Sprintf("**目的**: %s\n", ch.Purpose))

			if len(ch.KeyScenes) > 0 {
				f.WriteString("\n**关键场景**:\n")
				for _, scene := range ch.KeyScenes {
					f.WriteString(fmt.Sprintf("- %s\n", scene))
				}
			}

			if ch.PlotAdvancement != "" {
				f.WriteString(fmt.Sprintf("\n**情节推进**: %s", ch.PlotAdvancement))
			}

			f.WriteString("\n\n")
		}
	}

	// 场景列表
	if len(blueprint.Scenes) > 0 {
		f.WriteString("## 场景列表\n\n")
		for _, scene := range blueprint.Scenes {
			f.WriteString(fmt.Sprintf("**第%d章-场景%d**: %s\n", scene.Chapter, scene.Scene, scene.Purpose))
			if scene.Location != "" {
				f.WriteString(fmt.Sprintf("- 地点: %s\n", scene.Location))
			}
			if scene.Mood != "" {
				f.WriteString(fmt.Sprintf("- 氛围: %s\n", scene.Mood))
			}
			f.WriteString("\n")
		}
	}
}

func parseNarrativeStructure(s string) narrative.NarrativeStructure {
	switch s {
	case "three_act":
		return narrative.StructureThreeAct
	case "heros_journey":
		return narrative.StructureHerosJourney
	case "save_the_cat":
		return narrative.StructureSaveTheCat
	case "kishotenketsu":
		return narrative.StructureKishotenketsu
	case "freytag_pyramid":
		return narrative.StructureFreytagPyramid
	default:
		return narrative.StructureThreeAct
	}
}
