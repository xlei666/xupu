// Package cli CLI命令实现 - 世界设定
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/worldbuilder"
)

// NewWorldCommand 创建世界命令组
func NewWorldCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "world",
		Short: "世界设定管理",
	}

	cmd.AddCommand(newWorldListCmd())
	cmd.AddCommand(newWorldCreateCmd())
	cmd.AddCommand(newWorldShowCmd())
	cmd.AddCommand(newWorldDeleteCmd())

	return cmd
}

// newWorldListCmd 列出所有世界
func newWorldListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有世界设定",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			worlds := database.ListWorlds()

			PrintHeader("世界设定列表")

			if len(worlds) == 0 {
				PrintInfo("暂无世界设定")
				fmt.Println()
				PrintInfo("使用 'xupu world create' 创建新世界")
				return
			}

			rows := make([][]string, 0, len(worlds))
			for _, w := range worlds {
				rows = append(rows, []string{
					w.ID[:12],
					w.Name,
					FormatWorldType(w.Type),
					FormatWorldScale(w.Scale),
					w.Philosophy.CoreQuestion,
				})
			}

			PrintTable([]string{"ID", "名称", "类型", "规模", "核心问题"}, rows)
			fmt.Println()
			PrintInfo("共 %d 个世界", len(worlds))
		},
	}
}

// newWorldCreateCmd 创建新世界
func newWorldCreateCmd() *cobra.Command {
	var (
		name     string
		worldType string
		scale    string
		theme    string
		style    string
		quick    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建新世界设定",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				PrintError("请指定世界名称 (--name)")
				return
			}

			PrintHeader("创建世界设定")
			PrintInfo("名称: %s", name)
			PrintInfo("类型: %s", worldType)
			PrintInfo("规模: %s", scale)

			// 创建世界构建器
			builder, err := worldbuilder.New()
			if err != nil {
				PrintError("初始化失败: %v", err)
				return
			}

			// 构建世界
			world, err := builder.Build(worldbuilder.BuildParams{
				Name:   name,
				Type:   parseWorldType(worldType),
				Scale:  parseWorldScale(scale),
				Theme:  theme,
				Style:  style,
			})

			if err != nil {
				PrintError("构建世界失败: %v", err)
				return
			}

			// 保存世界
			database := GetDBOrExit()
			if err := database.SaveWorld(world); err != nil {
				PrintError("保存失败: %v", err)
				return
			}

			PrintSuccess("世界创建成功!")
			fmt.Println()

			if !quick {
				printWorldDetail(world, database)
			} else {
				PrintInfo("世界ID: %s", world.ID)
				PrintInfo("使用 'xupu world show %s' 查看详情", world.ID)
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "世界名称")
	cmd.Flags().StringVarP(&worldType, "type", "t", "fantasy", "世界类型 (fantasy/scifi/historical/urban/wuxia/xianxia/mixed)")
	cmd.Flags().StringVarP(&scale, "scale", "s", "continent", "世界规模 (village/city/nation/continent/planet/universe)")
	cmd.Flags().StringVarP(&theme, "theme", "T", "", "世界主题")
	cmd.Flags().StringVar(&style, "style", "", "世界风格")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "快速模式（不显示详情）")

	return cmd
}

// newWorldShowCmd 查看世界详情
func newWorldShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "查看世界详情",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()
			world, err := database.GetWorld(args[0])
			if err != nil {
				PrintError("世界不存在: %s", args[0])
				return
			}

			PrintHeader("世界设定详情")
			printWorldDetail(world, database)
		},
	}
}

// newWorldDeleteCmd 删除世界
func newWorldDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "删除世界设定",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			database := GetDBOrExit()

			if !confirm {
				world, _ := database.GetWorld(args[0])
				if world != nil {
					PrintWarn("即将删除世界: %s", world.Name)
				}
				fmt.Print("确认删除? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					PrintInfo("已取消")
					return
				}
			}

			if err := database.DeleteWorld(args[0]); err != nil {
				PrintError("删除失败: %v", err)
				return
			}

			PrintSuccess("世界已删除")
		},
	}

	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "跳过确认")
	return cmd
}

// printWorldDetail 打印世界详情
func printWorldDetail(world *models.WorldSetting, database db.Database) {
	// 基本信息
	fmt.Println(cyan.Sprintf("名称: %s", world.Name))
	fmt.Printf("类型: %s\n", FormatWorldType(world.Type))
	fmt.Printf("规模: %s\n", FormatWorldScale(world.Scale))
	if world.Style != "" {
		fmt.Printf("风格: %s\n", world.Style)
	}
	fmt.Printf("ID: %s\n", world.ID)
	fmt.Println()

	// 哲学设定
	PrintSection("哲学思考")
	if world.Philosophy.CoreQuestion != "" {
		fmt.Printf("核心问题: %s\n", world.Philosophy.CoreQuestion)
	}
	if world.Philosophy.ValueSystem.HighestGood != "" {
		fmt.Printf("至善: %s\n", world.Philosophy.ValueSystem.HighestGood)
	}
	if world.Philosophy.ValueSystem.UltimateEvil != "" {
		fmt.Printf("至恶: %s\n", world.Philosophy.ValueSystem.UltimateEvil)
	}

	// 地理
	if len(world.Geography.Regions) > 0 {
		PrintSection("地理环境")
		for _, region := range world.Geography.Regions {
			fmt.Printf("• %s", region.Name)
			if region.Type != "" {
				fmt.Printf(" [%s]", region.Type)
			}
			fmt.Println()
			if region.Description != "" {
				fmt.Printf("  %s\n", region.Description)
			}
		}
	}

	// 文明
	if len(world.Civilization.Races) > 0 {
		PrintSection("文明种族")
		for _, race := range world.Civilization.Races {
			fmt.Printf("• %s\n", race.Name)
			if race.Description != "" {
				fmt.Printf("  %s\n", race.Description)
			}
		}
	}

	// 故事土壤
	if len(world.StorySoil.SocialConflicts) > 0 {
		PrintSection("故事土壤")
		fmt.Println("社会冲突:")
		for _, conflict := range world.StorySoil.SocialConflicts {
			fmt.Printf("• [%s] %s\n", conflict.Type, conflict.Description)
		}
	}

	// 关联角色
	characters := database.ListCharactersByWorld(world.ID)
	if len(characters) > 0 {
		PrintSection("关联角色")
		for _, char := range characters {
			fmt.Printf("• %s\n", char.Name)
		}
	}

	fmt.Println()
}

// 解析函数
func parseWorldType(t string) models.WorldType {
	switch t {
	case "fantasy", "奇幻":
		return models.WorldFantasy
	case "scifi", "科幻":
		return models.WorldScifi
	case "historical", "历史":
		return models.WorldHistorical
	case "urban", "都市":
		return models.WorldUrban
	case "wuxia", "武侠":
		return models.WorldWuxia
	case "xianxia", "仙侠":
		return models.WorldXianxia
	default:
		return models.WorldFantasy
	}
}

func parseWorldScale(s string) models.WorldScale {
	switch s {
	case "village", "村庄":
		return models.ScaleVillage
	case "city", "城市":
		return models.ScaleCity
	case "nation", "国家":
		return models.ScaleNation
	case "continent", "大陆":
		return models.ScaleContinent
	case "planet", "星球":
		return models.ScalePlanet
	case "universe", "宇宙":
		return models.ScaleUniverse
	default:
		return models.ScaleContinent
	}
}
