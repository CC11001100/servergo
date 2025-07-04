package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"errors"

	"github.com/CC11001100/servergo/pkg/config"
	"github.com/CC11001100/servergo/pkg/dirlist"
	"github.com/CC11001100/servergo/pkg/i18n"
	"github.com/CC11001100/servergo/pkg/logger"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 支持的配置键列表
var validConfigKeys = []string{
	"auto-open",              // 是否自动打开浏览器
	"enable-dir-listing",     // 是否启用目录列表功能
	"theme",                  // 目录列表主题
	"language",               // 界面语言
	"enable-log-persistence", // 是否启用日志持久化
	"start-port",             // 从哪个端口开始递增寻找空闲端口
	// 在这里添加其他支持的配置键
}

// 使用dirlist包中定义的主题列表
// 支持的主题列表在 pkg/dirlist/themes.go 中定义

// configCmd 表示配置相关的命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("cmd.config.short"),
	Long:  i18n.T("cmd.config.long"),
}

// configListCmd 列出所有配置
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("cmd.config.list.short"),
	Long:  i18n.T("cmd.config.list.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		if err := config.InitConfig(); err != nil {
			return err
		}

		// 获取当前配置
		cfg := config.GetConfig()

		// 确保使用当前语言
		if err := i18n.Init(cfg.Language); err != nil {
			logger.Warning(fmt.Sprintf("Failed to initialize i18n: %v", err))
		}

		// 获取配置文件路径
		cfgPath, err := config.GetConfigFilePath()
		if err != nil {
			return err
		}

		// 检查配置文件是否实际存在
		fileExists := false
		if _, err := os.Stat(cfgPath); err == nil {
			fileExists = true
		}

		// 先单独显示配置文件路径信息
		logger.Info(i18n.T("config.current_config"))
		if fileExists {
			logger.Info(i18n.Tf("config.file_path", cfgPath))
		} else {
			logger.Info(i18n.T("config.file_not_created"))
			logger.Info(i18n.Tf("config.default_path", cfgPath))
		}
		logger.Info("")

		// 创建表格
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		// 设置表格样式
		t.SetStyle(table.StyleColoredBright)

		// 设置表头 - 确保使用当前语言的翻译
		UpdateConfigTableHeaders(t)

		// 添加配置信息行 - 确保使用当前语言的翻译
		t.AppendRows([]table.Row{
			{"auto-open", formatBoolValue(cfg.AutoOpen), i18n.T("config.auto_open_desc")},
			{"enable-dir-listing", formatBoolValue(cfg.EnableDirListing), i18n.T("config.enable_dir_listing_desc")},
			{"theme", cfg.Theme, i18n.T("config.theme_desc")},
			{"language", formatLanguageValue(cfg.Language), i18n.T("config.language_desc")},
			{"enable-log-persistence", formatBoolValue(cfg.EnableLogPersistence), i18n.T("config.enable_log_persistence_desc")},
			{"start-port", fmt.Sprintf("%d", cfg.StartPort), i18n.T("config.start_port_desc")},
		})

		// 设置列对齐方式
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignLeft, AlignHeader: text.AlignCenter, WidthMax: 30},
			{Number: 2, Align: text.AlignCenter, AlignHeader: text.AlignCenter, WidthMax: 20},
			{Number: 3, Align: text.AlignLeft, AlignHeader: text.AlignCenter},
		})

		// 输出表格
		t.Render()

		return nil
	},
}

// 生成配置命令缺少参数的友好错误信息
func generateConfigCommandHelp(cmdName string, args []string) string {
	var msg strings.Builder

	if cmdName == "get" {
		msg.WriteString(i18n.T("cmd.get.missing_arg") + "\n\n")
	} else if cmdName == "set" {
		if len(args) == 0 {
			msg.WriteString(i18n.T("cmd.set.missing_key_value") + "\n\n")
		} else {
			msg.WriteString(i18n.T("cmd.set.missing_value") + "\n\n")
		}
	}

	msg.WriteString(i18n.T("cmd.available_items") + "\n")
	for _, key := range validConfigKeys {
		fmt.Fprintf(&msg, "  - %s\n", key)
	}

	msg.WriteString("\n" + i18n.T("cmd.usage") + "\n")
	if cmdName == "get" {
		msg.WriteString("  " + i18n.T("cmd.get.usage") + "\n\n")
		msg.WriteString(i18n.T("cmd.examples") + "\n")
		msg.WriteString("  " + i18n.T("cmd.get.example1") + "\n")
		msg.WriteString("  " + i18n.T("cmd.get.example2") + "\n")
		msg.WriteString("  " + i18n.T("cmd.get.example3") + "\n")
		msg.WriteString("  " + i18n.T("cmd.get.example4") + "\n")
	} else if cmdName == "set" {
		msg.WriteString("  " + i18n.T("cmd.set.usage") + "\n\n")
		msg.WriteString(i18n.T("cmd.examples") + "\n")
		msg.WriteString("  " + i18n.T("cmd.set.example1") + "\n")
		msg.WriteString("  " + i18n.T("cmd.set.example2") + "\n")
		msg.WriteString("  " + i18n.T("cmd.set.example3") + "\n")
		msg.WriteString("  " + i18n.T("cmd.set.example4") + "\n")

		if len(args) == 1 {
			msg.WriteString("\n" + i18n.T("cmd.provided_item") + args[0] + "\n")
			if args[0] == "theme" {
				// 使用全局定义的有效主题列表
				themesStr := strings.Join(dirlist.GetSupportedThemes(), ", ")
				msg.WriteString(i18n.T("cmd.theme.options") + themesStr + "\n")
			} else if args[0] == "auto-open" || args[0] == "enable-dir-listing" || args[0] == "enable-log-persistence" {
				msg.WriteString(i18n.T("cmd.bool.options") + "\n")
			} else if args[0] == "language" {
				// 使用语言模块提供的支持语言列表
				supportedLangs := strings.Join(i18n.GetSupportedLanguages(), ", ")
				msg.WriteString(i18n.T("cmd.language.options") + supportedLangs + "\n")
			} else if args[0] == "start-port" {
				msg.WriteString(i18n.T("cmd.start_port.options") + "\n")
			}
		}
	}

	return msg.String()
}

// configGetCmd 获取指定配置
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: i18n.T("cmd.config.get.short"),
	Long:  i18n.T("cmd.config.get.long"),
	// 不使用标准参数验证，改用自定义验证
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// 没有提供配置项，列出所有可用的配置项
			fmt.Println(i18n.T("cmd.available_items"))
			for _, key := range validConfigKeys {
				fmt.Println("  -", key)
			}
			os.Exit(0)
		} else if len(args) > 1 {
			// 提供了过多的参数
			return errors.New(generateConfigCommandHelp("get", args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		if err := config.InitConfig(); err != nil {
			return err
		}

		key := args[0]

		// 验证key是否合法
		if !isValidConfigKey(key) {
			return generateInvalidKeyError(key)
		}

		// 获取配置值
		value := viper.Get(key)
		if value == nil {
			return fmt.Errorf(i18n.Tf("error.config_item_not_exist", key))
		}

		// 直接输出原始值，不做格式转换
		fmt.Println(value)
		return nil
	},
}

// configSetCmd 设置指定配置
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: i18n.T("cmd.config.set.short"),
	Long:  i18n.T("cmd.config.set.long"),
	// 不使用标准参数验证，改用自定义验证
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// 没有提供任何参数，显示常规帮助信息
			return errors.New(generateConfigCommandHelp("set", args))
		} else if len(args) == 1 {
			// 只提供了配置项名称，但没有提供值
			key := args[0]
			if !isValidConfigKey(key) {
				return generateInvalidKeyError(key)
			}

			// 根据配置项类型，直接显示可选的值并退出程序
			switch key {
			case "theme":
				// 直接显示所有可用主题
				fmt.Println(i18n.T("cmd.theme.options"))
				for _, theme := range dirlist.GetSupportedThemes() {
					fmt.Println("  -", theme)
				}
				os.Exit(0)
			case "auto-open", "enable-dir-listing", "enable-log-persistence":
				fmt.Println(i18n.T("cmd.bool.options"))
				fmt.Println("  - true, yes, y, 1, on")
				fmt.Println("  - false, no, n, 0, off")
				os.Exit(0)
			case "language":
				fmt.Println(i18n.T("cmd.language.options"))
				langs := i18n.GetSupportedLanguages()
				for _, lang := range langs {
					// 尝试获取该语言的显示名称
					displayName := i18n.GetLanguageDisplayName(lang)
					fmt.Printf("  - %s (%s)\n", lang, displayName)
				}
				os.Exit(0)
			}
		} else if len(args) > 2 {
			// 参数太多
			return errors.New(generateConfigCommandHelp("set", []string{args[0]}))
		}

		// 正确提供了两个参数
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		if err := config.InitConfig(); err != nil {
			return err
		}

		key := args[0]
		value := args[1]

		// 验证key是否合法
		if !isValidConfigKey(key) {
			return generateInvalidKeyError(key)
		}

		// 设置配置值
		if err := setConfigValue(key, value); err != nil {
			return err
		}

		// 获取完整的配置对象
		cfg := config.GetConfig()

		// 保存配置
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf(i18n.Tf("error.cannot_save_config", err))
		}

		// 特殊处理语言变化的消息
		if key == "language" {
			if err := i18n.Init(value); err != nil {
				logger.Warning(i18n.Tf("config.language_init_error", err))
			}

			// 语言改变后，更新所有命令描述
			UpdateCommandDescriptions()

			// 语言改变后，也需要重新加载表格表头的翻译
			// 这里不需要显式操作，因为表头是在每次configListCmd运行时创建的

			languageDisplayName := i18n.GetLanguageDisplayName(value)
			logger.Info(i18n.Tf("config.language_changed", languageDisplayName))
		} else {
			logger.Info(i18n.Tf("config.item_set", key, value))
		}

		return nil
	},
}

// 检查配置键是否有效
func isValidConfigKey(key string) bool {
	for _, validKey := range validConfigKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

// 生成无效key的友好错误信息
func generateInvalidKeyError(key string) error {
	var msg strings.Builder

	// 不支持的配置项
	fmt.Fprintf(&msg, "%s\n\n", i18n.Tf("error.invalid_config_key", key))

	// 支持的配置项列表
	msg.WriteString(i18n.T("error.available_keys") + "\n")
	for _, validKey := range validConfigKeys {
		fmt.Fprintf(&msg, "  - %s\n", validKey)
	}

	// 添加配置项说明
	msg.WriteString("\n" + i18n.T("error.key_descriptions") + "\n")
	msg.WriteString("  - " + i18n.T("error.auto_open_desc") + "\n")
	msg.WriteString("  - " + i18n.T("error.enable_dir_listing_desc") + "\n")
	msg.WriteString("  - " + i18n.T("error.theme_desc") + "\n")
	msg.WriteString("  - " + i18n.T("error.language_desc") + "\n")
	msg.WriteString("  - " + i18n.T("error.enable_log_persistence_desc") + "\n")
	msg.WriteString("  - " + i18n.T("error.start_port_desc") + "\n")

	return fmt.Errorf(msg.String())
}

// 将字符串转换为布尔值，支持多种表示形式
func parseBoolValue(value string) (bool, error) {
	// 转为小写便于比较
	val := strings.ToLower(value)

	switch val {
	case "true", "yes", "y", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf(i18n.T("error.invalid_bool"))
	}
}

// 设置配置值（根据类型转换）
func setConfigValue(key, value string) error {
	switch key {
	case "auto-open", "enable-dir-listing", "enable-log-persistence":
		// 将输入转换为布尔值
		boolValue, err := parseBoolValue(value)
		if err != nil {
			return err
		}
		viper.Set(key, boolValue)

	case "start-port":
		// 将输入转换为整数
		portValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf(i18n.Tf("error.invalid_port_number", value))
		}
		// 验证端口范围
		if portValue < 0 || portValue > 65535 {
			return fmt.Errorf(i18n.Tf("error.port_out_of_range", portValue))
		}
		viper.Set(key, portValue)

	case "theme":
		// 验证主题名称是否有效
		isValid := false
		for _, theme := range dirlist.GetSupportedThemes() {
			if value == theme {
				isValid = true
				break
			}
		}
		if !isValid {
			// 传入所有支持的主题列表
			themesStr := strings.Join(dirlist.GetSupportedThemes(), ", ")
			return fmt.Errorf(i18n.Tf("error.invalid_theme", value, themesStr))
		}
		viper.Set(key, value)

	case "language":
		// 验证语言是否被支持
		if !i18n.IsSupportedLanguage(value) {
			supportedLangs := strings.Join(i18n.GetSupportedLanguages(), ", ")
			return fmt.Errorf(i18n.Tf("error.invalid_language", value, supportedLangs))
		}
		viper.Set(key, value)

		// 语言设置特殊处理：同时更新i18n包的语言设置
		if err := config.SetLanguage(value); err != nil {
			return fmt.Errorf(i18n.Tf("error.cannot_set_language", err))
		}

	default:
		// 这里不应该到达，因为已经在前面验证了key的有效性
		return fmt.Errorf(i18n.Tf("error.unknown_config_item", key))
	}

	return nil
}

// formatBoolValue 格式化布尔值为更友好的显示文本
func formatBoolValue(value bool) string {
	if value {
		return i18n.T("config.enabled")
	}
	return i18n.T("config.disabled")
}

// 格式化语言值，显示友好的语言名称
func formatLanguageValue(lang string) string {
	return i18n.GetLanguageDisplayName(lang)
}

// UpdateConfigTableHeaders 更新配置表格的表头为当前语言
func UpdateConfigTableHeaders(t table.Writer) {
	t.ResetHeaders()

	// 强制重新获取当前语言的翻译
	itemHeader := i18n.T("config.item")
	valueHeader := i18n.T("config.current_value")
	descHeader := i18n.T("config.description")

	t.AppendHeader(table.Row{itemHeader, valueHeader, descHeader})
}

func init() {
	RootCmd.AddCommand(configCmd)

	// 添加子命令
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}
