package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// selectServices 使用上下键交互式选择要生成的服务
func selectServices(services []ServiceInfo) []ServiceInfo {
	if len(services) == 0 {
		return []ServiceInfo{}
	}

	// 检查是否是终端
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return selectServicesSimple(services)
	}

	// 保存终端状态
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return selectServicesSimple(services)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	current := 0

	// 传入 false，表示第一次渲染
	renderMenu(services, current, false)

	b := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			break
		}

		if b[0] == 27 && b[1] == '[' {
			switch b[2] {
			case 'A': // 上箭头
				if current > 0 {
					current--
					// 传入 true，表示是更新，需要光标上移
					renderMenu(services, current, true)
				}
			case 'B': // 下箭头
				if current < len(services)-1 {
					current++
					// 传入 true，表示是更新
					renderMenu(services, current, true)
				}
			}
		} else if b[0] == 13 || b[0] == 10 { // 回车
			break
		} else if b[0] == 3 || b[0] == 27 { // Ctrl+C 或 ESC
			return []ServiceInfo{}
		}
	}

	return []ServiceInfo{services[current]}
}

// selectServicesSimple 简单输入方式（用于非终端环境）
func selectServicesSimple(services []ServiceInfo) []ServiceInfo {
	fmt.Println("\n请选择要生成的文件（输入编号，多个用逗号分隔，如: 1,3,5 或输入 'all' 生成全部）：")
	fmt.Print("> ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		return []ServiceInfo{}
	}

	if strings.ToLower(input) == "all" {
		return services
	}

	// 解析输入
	selectedMap := make(map[int]bool)
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if num := parseInt(part); num > 0 && num <= len(services) {
			selectedMap[num-1] = true
		}
	}

	var result []ServiceInfo
	for i := range selectedMap {
		result = append(result, services[i])
	}

	return result
}

// renderMenu 渲染菜单
// isUpdate: 如果是 true，说明之前画过，需要把光标移上去覆盖
func renderMenu(services []ServiceInfo, current int, isUpdate bool) {
	// 1. 如果是更新模式，先让光标往上跑，回到菜单开头
	if isUpdate {
		// 上移行数 = 服务列表数量 + 2行标题
		fmt.Printf("\033[%dA", len(services)+2)
	}

	// 2. 输出标题 (注意这里用了 \r\n 解决阶梯换行)
	os.Stdout.WriteString("请使用上下键选择，回车确认，ESC 退出：\r\n\r\n")

	// 3. 循环输出每一行
	for i, service := range services {
		// \033[K 表示清除光标这一行后面的内容（防止残留）
		line := fmt.Sprintf("%s (Handler: %s, %d 个方法)\033[K",
			service.FileName, service.HandlerName, len(service.Methods))

		if i == current {
			// 选中行：高亮 + \r\n
			os.Stdout.WriteString("  \033[7m" + line + "\033[0m\r\n")
		} else {
			// 普通行 + \r\n
			os.Stdout.WriteString("  " + line + "\r\n")
		}
	}

	// 立即刷新显示
	os.Stdout.Sync()
}
