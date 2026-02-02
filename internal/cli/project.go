// Package cli CLI命令实现 - 项目管理
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/orchestrator"
	"github.com/xlei/xupu/pkg/scheduler"
)

// NewProjectCommand 创建项目命令组
func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "项目管理",
	}

	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectCreateCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectDeleteCmd())
	cmd.AddCommand(newProjectGenerateCmd())

	return cmd
}

// newProjectListCmd 列出所有项目
func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有项目",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			projects := database.ListProjects()

			PrintHeader("项目列表")

			if len(projects) == 0 {
				PrintInfo("暂无项目")
				fmt.Println()
				PrintInfo("使用 'xupu project create' 创建新项目")
				return
			}

			rows := make([][]string, 0, len(projects))
			for _, p := range projects {
				statusColor := FormatProjectStatus(p.Status)
				if p.Status == models.StatusCompleted {
					statusColor = green.Sprint(statusColor)
				} else if p.Status == models.StatusFailed {
					statusColor = red.Sprint(statusColor)
				}

				rows = append(rows, []string{
					p.ID[:12],
					p.Name,
					FormatProjectMode(p.Mode),
					statusColor,
					fmt.Sprintf("%.0f%%", p.Progress),
				})
			}

			PrintTable([]string{"ID", "名称", "模式", "状态", "进度"}, rows)
			fmt.Println()
			PrintInfo("共 %d 个项目", len(projects))
		},
	}
}

// newProjectCreateCmd 创建新项目
func newProjectCreateCmd() *cobra.Command {
	var (
		name        string
		description string
		mode        string
		// 世界参数
		worldType   string
		worldScale  string
		worldTheme  string
		worldStyle  string
		// 故事参数
		storyType   string
		theme       string
		protagonist string
		length      string
		chapters    int
		structure   string
		// 异步选项
		async       bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建新项目",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				PrintError("请指定项目名称 (--name)")
				return
			}

			PrintHeader("创建新项目")

			params := orchestrator.CreationParams{
				ProjectName: name,
				Description: description,
				WorldType:   worldType,
				WorldScale:  worldScale,
				WorldTheme:  worldTheme,
				WorldStyle:  worldStyle,
				StoryType:   storyType,
				StoryTheme:  theme,
				Protagonist: protagonist,
				StoryLength: length,
				ChapterCount: chapters,
				Structure:   structure,
				Options: orchestrator.GenerationOptions{
					GenerateContent: false, // 默认不生成内容
				},
			}

			if async {
				// 异步创建
				orc, err := orchestrator.New()
				if err != nil {
					PrintError("初始化编排器失败: %v", err)
					return
				}

				task, err := orchestrator.CreateProjectAsync(params, orc)
				if err != nil {
					PrintError("创建任务失败: %v", err)
					return
				}

				PrintSuccess("项目已提交到任务队列")
				PrintInfo("任务ID: %s", task.ID)
				fmt.Println()
				PrintInfo("使用以下命令查看进度:")
				yellow.Printf("  xupu task wait %s\n", task.ID)

			} else {
				// 同步创建
				orc, err := orchestrator.New()
				if err != nil {
					PrintError("初始化编排器失败: %v", err)
					return
				}

				PrintInfo("正在创建项目...")
				project, err := orc.CreateProject(params)
				if err != nil {
					PrintError("创建项目失败: %v", err)
					return
				}

				PrintSuccess("项目创建成功!")
				fmt.Println()
				printProjectDetail(project, GetDBOrExit())
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "项目名称")
	cmd.Flags().StringVarP(&description, "description", "d", "", "项目描述")
	cmd.Flags().StringVarP(&mode, "mode", "m", "planning", "创作模式 (planning/intervention/random)")
	// 世界参数
	cmd.Flags().StringVar(&worldType, "world-type", "fantasy", "世界类型 (fantasy/scifi/historical/urban/wuxia/xianxia/mixed)")
	cmd.Flags().StringVar(&worldScale, "world-scale", "continent", "世界规模 (village/city/nation/continent/planet/universe)")
	cmd.Flags().StringVar(&worldTheme, "world-theme", "", "世界主题")
	cmd.Flags().StringVar(&worldStyle, "world-style", "", "世界风格")
	// 故事参数
	cmd.Flags().StringVar(&storyType, "story-type", "adventure", "故事类型")
	cmd.Flags().StringVar(&theme, "theme", "", "故事主题")
	cmd.Flags().StringVar(&protagonist, "protagonist", "", "主角设定")
	cmd.Flags().StringVar(&length, "length", "medium", "故事长度 (short/medium/long)")
	cmd.Flags().IntVar(&chapters, "chapters", 12, "章节数量")
	cmd.Flags().StringVar(&structure, "structure", "three_act", "叙事结构 (three_act/heros_journey/save_the_cat)")
	// 选项
	cmd.Flags().BoolVar(&async, "async", false, "异步创建")

	return cmd
}

// newProjectShowCmd 查看项目详情
func newProjectShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "查看项目详情",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			project, err := database.GetProject(args[0])
			if err != nil {
				PrintError("项目不存在: %s", args[0])
				return
			}

			PrintHeader("项目详情")
			printProjectDetail(project, database)
		},
	}
}

// newProjectDeleteCmd 删除项目
func newProjectDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "删除项目",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()

			// 确认
			if !confirm {
				PrintWarn("此操作将删除项目及其所有数据")
				fmt.Print("确认删除? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					PrintInfo("已取消")
					return
				}
			}

			if err := database.DeleteProject(args[0]); err != nil {
				PrintError("删除失败: %v", err)
				return
			}

			PrintSuccess("项目已删除")
		},
	}

	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "跳过确认")
	return cmd
}

// newProjectGenerateCmd 生成项目内容
func newProjectGenerateCmd() *cobra.Command {
	var startChapter, endChapter int

	cmd := &cobra.Command{
		Use:   "generate <id>",
		Short: "生成项目内容",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			project, err := database.GetProject(args[0])
			if err != nil {
				PrintError("项目不存在: %s", args[0])
				return
			}

			if project.Status != models.StatusCompleted {
				PrintError("项目状态不正确，需要先完成规划")
				return
			}

			PrintHeader("开始生成内容")
			PrintInfo("项目: %s", project.Name)

			if endChapter == 0 {
				endChapter = 999 // 生成所有章节
			}

			// TODO: 实现内容生成逻辑
			PrintWarn("内容生成功能开发中")
		},
	}

	cmd.Flags().IntVar(&startChapter, "start", 1, "起始章节")
	cmd.Flags().IntVar(&endChapter, "end", 0, "结束章节")

	return cmd
}

// printProjectDetail 打印项目详情
func printProjectDetail(project *models.Project, database db.Database) {
	fmt.Printf("ID:         %s\n", project.ID)
	fmt.Printf("名称:       %s\n", project.Name)
	if project.Description != "" {
		fmt.Printf("描述:       %s\n", project.Description)
	}
	fmt.Printf("模式:       %s\n", FormatProjectMode(project.Mode))
	fmt.Printf("状态:       %s\n", FormatProjectStatus(project.Status))
	fmt.Printf("进度:       %.1f%%\n", project.Progress)

	if project.WorldID != "" {
		fmt.Printf("世界ID:     %s\n", project.WorldID)
		if world, err := database.GetWorld(project.WorldID); err == nil {
			fmt.Printf("世界名称:   %s\n", world.Name)
		}
	}

	if project.NarrativeID != "" {
		fmt.Printf("叙事ID:     %s\n", project.NarrativeID)
	}

	fmt.Printf("创建时间:   %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 获取进度信息
	if project.NarrativeID != "" {
		if blueprint, err := database.GetNarrativeBlueprint(project.NarrativeID); err == nil {
			fmt.Println("叙事蓝图:")
			fmt.Printf("  结构类型: %s\n", blueprint.StoryOutline.StructureType)
			fmt.Printf("  章节数量: %d\n", len(blueprint.ChapterPlans))
			fmt.Printf("  场景数量: %d\n", len(blueprint.Scenes))

			scenes := database.ListScenesByBlueprint(project.NarrativeID)
			fmt.Printf("  已生成场景: %d/%d\n", len(scenes), len(blueprint.Scenes))
		}
	}
}

// ============================================
// 任务命令
// ============================================

// NewTaskCommand 创建任务命令
func NewTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "任务管理",
	}

	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskShowCmd())
	cmd.AddCommand(newTaskCancelCmd())
	cmd.AddCommand(newTaskWaitCmd())

	return cmd
}

func newTaskListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有任务",
		Run: func(cmd *cobra.Command, args []string) {
			PrintHeader("任务列表")
			PrintInfo("任务功能开发中")
		},
	}
}

func newTaskShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "查看任务状态",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			task, err := orchestrator.GetTask(args[0])
			if err != nil {
				PrintError("任务不存在: %s", args[0])
				return
			}

			PrintHeader("任务状态")
			fmt.Printf("任务ID:    %s\n", task.ID)
			fmt.Printf("类型:      %s\n", task.Type)
			fmt.Printf("状态:      %s\n", task.GetStatus())
			fmt.Printf("进度:      %.1f%%\n", task.GetProgress())
			if task.Error != "" {
				fmt.Printf("错误:      %s\n", task.Error)
			}
		},
	}
}

func newTaskCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "取消任务",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := orchestrator.CancelTask(args[0]); err != nil {
				PrintError("取消任务失败: %v", err)
				return
			}
			PrintSuccess("任务已取消")
		},
	}
}

func newTaskWaitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wait <id>",
		Short: "等待任务完成",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := args[0]

			PrintInfo("等待任务完成...")
			fmt.Println()

			ticker := 0
			for {
				task, err := orchestrator.GetTask(taskID)
				if err != nil {
					PrintError("任务不存在: %s", taskID)
					return
				}

				status := task.GetStatus()
				progress := task.GetProgress()

				// 打印进度
				fmt.Printf("\r[%-50s] %.1f%%", getProgressBar(progress), progress)

				if status == scheduler.StatusCompleted {
					fmt.Println()
					PrintSuccess("任务完成!")
					if task.GetResult() != nil {
						fmt.Printf("结果: %+v\n", task.GetResult())
					}
					return
				}

				if status == scheduler.StatusFailed {
					fmt.Println()
					PrintError("任务失败: %s", task.Error)
					return
				}

				if status == scheduler.StatusCancelled {
					fmt.Println()
					PrintInfo("任务已取消")
					return
				}

				ticker++
				if ticker > 300 { // 5分钟超时
					fmt.Println()
					PrintWarn("等待超时")
					return
				}
			}
		},
	}
}

func getProgressBar(progress float64) string {
	bars := int(progress / 2)
	result := ""
	for i := 0; i < 50; i++ {
		if i < bars {
			result += "█"
		} else {
			result += "░"
		}
	}
	return result
}
