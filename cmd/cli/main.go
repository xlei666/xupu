// Package main Xupu AI小说创作系统 - CLI工具
package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xlei/xupu/internal/cli"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/orchestrator"
)

var (
	cfgFile string
	verbose bool
)

func main() {
	// 初始化数据库
	_ = db.Get()

	// 初始化全局调度器
	if err := orchestrator.InitScheduler(); err != nil {
		color.Red("初始化调度器失败: %v", err)
		os.Exit(1)
	}
	defer orchestrator.StopScheduler()

	var rootCmd = &cobra.Command{
		Use:   "xupu",
		Short: "Xupu AI小说创作系统",
		Long: `Xupu - AI驱动的小说创作系统
支持世界设定构建、叙事规划、场景生成等完整创作流程。`,
		SilenceUsage: true,
	}

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	// 添加子命令
	rootCmd.AddCommand(cli.NewProjectCommand())
	rootCmd.AddCommand(cli.NewWorldCommand())
	rootCmd.AddCommand(cli.NewBlueprintCommand())
	rootCmd.AddCommand(cli.NewGenerateCommand())
	rootCmd.AddCommand(cli.NewExportCommand())
	rootCmd.AddCommand(cli.NewConfigCommand())
	rootCmd.AddCommand(cli.NewVersionCommand())

	// 执行
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		color.Red("错误: %v", err)
		os.Exit(1)
	}
}
