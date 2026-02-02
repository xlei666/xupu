// Package cli CLI命令实现
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
)

var (
	cyan    = color.New(color.FgCyan)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	red     = color.New(color.FgRed)
	white   = color.New(color.FgWhite)
	gray    = color.New(color.FgHiBlack)
)

// PrintHeader 打印标题
func PrintHeader(title string) {
	fmt.Println()
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	white.Printf("  %s\n", title)
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// PrintSuccess 打印成功消息
func PrintSuccess(format string, args ...interface{}) {
	green.Printf("✓ %s\n", fmt.Sprintf(format, args...))
}

// PrintInfo 打印信息
func PrintInfo(format string, args ...interface{}) {
	cyan.Printf("  %s\n", fmt.Sprintf(format, args...))
}

// PrintWarn 打印警告
func PrintWarn(format string, args ...interface{}) {
	yellow.Printf("⚠ %s\n", fmt.Sprintf(format, args...))
}

// PrintError 打印错误
func PrintError(format string, args ...interface{}) {
	red.Printf("✗ %s\n", fmt.Sprintf(format, args...))
}

// PrintSection 打印分节
func PrintSection(title string) {
	fmt.Println()
	yellow.Printf("▶ %s", title)
	fmt.Println()
	cyan.Println("────────────────────────────────────────────────────")
}

// FormatWorldType 格式化世界类型
func FormatWorldType(t models.WorldType) string {
	switch t {
	case models.WorldFantasy:
		return "奇幻"
	case models.WorldScifi:
		return "科幻"
	case models.WorldHistorical:
		return "历史"
	case models.WorldUrban:
		return "都市"
	case models.WorldWuxia:
		return "武侠"
	case models.WorldXianxia:
		return "仙侠"
	case models.WorldMixed:
		return "混合"
	default:
		return string(t)
	}
}

// FormatWorldScale 格式化世界规模
func FormatWorldScale(s models.WorldScale) string {
	switch s {
	case models.ScaleVillage:
		return "村庄"
	case models.ScaleCity:
		return "城市"
	case models.ScaleNation:
		return "国家"
	case models.ScaleContinent:
		return "大陆"
	case models.ScalePlanet:
		return "星球"
	case models.ScaleUniverse:
		return "宇宙"
	default:
		return string(s)
	}
}

// FormatProjectStatus 格式化项目状态
func FormatProjectStatus(s models.ProjectStatus) string {
	switch s {
	case models.StatusDraft:
		return "草稿"
	case models.StatusBuilding:
		return "构建中"
	case models.StatusGenerating:
		return "生成中"
	case models.StatusCompleted:
		return "已完成"
	case models.StatusPaused:
		return "已暂停"
	case models.StatusFailed:
		return "失败"
	default:
		return string(s)
	}
}

// FormatProjectMode 格式化项目模式
func FormatProjectMode(m models.OrchestrationMode) string {
	switch m {
	case models.ModePlanning:
		return "规划生成"
	case models.ModeIntervention:
		return "干预生成"
	case models.ModeRandom:
		return "随机生成"
	case models.ModeStoryCore:
		return "故事核"
	case models.ModeShort:
		return "短篇模式"
	case models.ModeScript:
		return "剧本模式"
	default:
		return string(m)
	}
}

// PrintTable 打印表格
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(color.Output, 0, 0, 2, ' ', 0)

	// 打印表头
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		white.Fprint(w, h)
	}
	fmt.Fprintln(w)

	// 打印分隔线
	for i := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		gray.Fprint(w, "──────")
	}
	fmt.Fprintln(w)

	// 打印数据行
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, cell)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// GetDBOrExit 获取数据库或退出
func GetDBOrExit() db.Database {
	database := db.Get()
	return database
}

// ============================================
// 版本命令
// ============================================

// NewVersionCommand 创建版本命令
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			cyan.Println("   ╔═══════════════════════════════╗")
			cyan.Println("   ║                               ║")
			cyan.Println("   ║    NovelFlow 叙谱 AI 创作系统   ║")
			cyan.Println("   ║                               ║")
			cyan.Println("   ║      版本: 0.1.0               ║")
			cyan.Println("   ║                               ║")
			cyan.Println("   ╚═══════════════════════════════╝")
			fmt.Println()
			white.Println("  智能AI驱动的小说创作平台")
			fmt.Println()
			gray.Println("  功能模块:")
			fmt.Println("    • 世界设定构建器")
			fmt.Println("    • 叙事蓝图引擎")
			fmt.Println("    • 场景内容生成器")
			fmt.Println("    • 异步任务调度")
			fmt.Println()
		},
	}
}
