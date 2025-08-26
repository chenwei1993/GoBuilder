package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	defer func() {
		if runtime.GOOS == "windows" {
			fmt.Print("\n程序结束，按任意键退出...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
	}()

	fmt.Println("=======================================")
	fmt.Println("Go源文件编译成可运行程序（支持windows、linux、macos）")
	fmt.Println("=======================================\n")

	reader := bufio.NewReader(os.Stdin)

	// 自动查找 Go
	goExe := findGo()
	if goExe == "" {
		fmt.Println("未找到 Go 可执行文件，请先安装 Go 并确保 PATH 配置正确")
		return
	}
	fmt.Println("使用 Go 编译器路径:", goExe)

	// 输入项目路径
	fmt.Print("请输入Go项目路径（留空表示当前目录）: ")
	projectPath, _ := reader.ReadString('\n')
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		projectPath, _ = os.Getwd()
	} else {
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Println("路径无效:", err)
			return
		}
		projectPath = absPath
	}

	// 输入 Go 文件名
	fmt.Print("请输入要编译的Go文件名(默认 main.go): ")
	goFile, _ := reader.ReadString('\n')
	goFile = strings.TrimSpace(goFile)
	if goFile == "" {
		goFile = "main.go"
	}
	goFilePath := filepath.Join(projectPath, goFile)

	// 输入 GOOS
	fmt.Print("请输入目标GOOS(例如 windows（默认） / linux / darwin): ")
	goos, _ := reader.ReadString('\n')
	goos = strings.TrimSpace(goos)
	if goos == "" {
		goos = "windows"
	}

	// 输入 GOARCH
	fmt.Print("请输入目标GOARCH(例如 amd64（默认） / arm64): ")
	goarch, _ := reader.ReadString('\n')
	goarch = strings.TrimSpace(goarch)
	if goarch == "" {
		goarch = "amd64"
	}

	// 输入基础输出文件名
	fmt.Print("请输入编译后的基础文件名(默认 main): ")
	baseName, _ := reader.ReadString('\n')
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "main"
	}

	// 拼接输出文件名
	output := fmt.Sprintf("%s-%s-%s", baseName, goos, goarch)
	if goos == "windows" && !strings.HasSuffix(output, ".exe") {
		output += ".exe"
	}
	outputPath := filepath.Join(projectPath, output)

	fmt.Println("\n===== 开始编译 =====")
	fmt.Printf("项目路径: %s\n", projectPath)
	fmt.Printf("Go文件: %s\n", goFilePath)
	fmt.Printf("目标平台: %s/%s\n", goos, goarch)
	fmt.Printf("输出文件: %s\n", outputPath)
	fmt.Println("====================\n")

	// 设置 go build 命令
	cmd := exec.Command(goExe, "build", "-ldflags", "-s -w", "-o", outputPath, goFilePath)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	cmd.Dir = projectPath

	outputBytes, err := cmd.CombinedOutput()
	fmt.Println(string(outputBytes))
	if err != nil {
		fmt.Println("编译失败:", err)
		return
	}

	fmt.Printf("\n编译完成，输出文件：%s\n", outputPath)
}

// -------------------- 搜索 Go --------------------

func findGo() string {
	// 1. PATH
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}

	// 2. 默认安装路径
	if runtime.GOOS == "windows" {
		paths := []string{
			"C:\\Go\\bin\\go.exe",
			"C:\\Program Files\\Go\\bin\\go.exe",
			"C:\\Program Files (x86)\\Go\\bin\\go.exe",
			filepath.Join(os.Getenv("USERPROFILE"), "Go\\bin\\go.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}

		// 3. 盘符根目录下子目录扫描
		if goPath := findGoInRootDirs(); goPath != "" {
			return goPath
		}

		// 4. 高概率目录 + 限制深度遍历（可选）
		if goPath := findGoInDrives(); goPath != "" {
			return goPath
		}

	} else {
		// Linux/macOS 默认路径
		paths := []string{
			"/usr/local/go/bin/go",
			"/usr/bin/go",
			"/usr/local/bin/go",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return ""
}

// 获取所有盘符
func getDrives() []string {
	var drives []string
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drive := string(letter) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}

// 扫描盘符根目录下子目录
func findGoInRootDirs() string {
	drives := getDrives()
	for _, d := range drives {
		entries, _ := os.ReadDir(d)
		for _, e := range entries {
			if e.IsDir() {
				goPath := filepath.Join(d, e.Name(), "bin", "go.exe")
				if _, err := os.Stat(goPath); err == nil {
					return goPath
				}
			}
		}
	}
	return ""
}

// 在盘符下搜索高概率 Go 目录
func findGoInDrives() string {
	drives := getDrives()
	for _, d := range drives {
		paths := []string{
			filepath.Join(d, "Go"),
			filepath.Join(d, "Program Files", "Go"),
			filepath.Join(d, "Program Files (x86)", "Go"),
		}
		for _, p := range paths {
			if goPath := walkDirLimited(p, 4); goPath != "" {
				return goPath
			}
		}
	}
	return ""
}

// 限制深度遍历目录找 go.exe
func walkDirLimited(root string, maxDepth int) string {
	var goPath string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || goPath != "" {
			return nil
		}
		if info.IsDir() {
			depth := strings.Count(path, string(os.PathSeparator)) - strings.Count(root, string(os.PathSeparator))
			if depth > maxDepth {
				return filepath.SkipDir
			}
		} else if info.Name() == "go.exe" || info.Name() == "go" {
			goPath = path
			return fmt.Errorf("found")
		}
		return nil
	})
	return goPath
}
