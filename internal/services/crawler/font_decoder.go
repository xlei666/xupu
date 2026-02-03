package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// FontDecoder 字体解码器，使用Python脚本进行字体解密
type FontDecoder struct {
	mapping map[rune]rune
	mu      sync.RWMutex
}

// NewFontDecoder 创建字体解码器
func NewFontDecoder() *FontDecoder {
	return &FontDecoder{
		mapping: make(map[rune]rune),
	}
}

// LoadFromURL 从URL下载字体并构建映射
func (d *FontDecoder) LoadFromURL(fontURL string) error {
	resp, err := http.Get(fontURL)
	if err != nil {
		return fmt.Errorf("下载字体失败: %w", err)
	}
	defer resp.Body.Close()

	fontData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取字体数据失败: %w", err)
	}

	return d.BuildMapping(fontData)
}

// BuildMapping 使用Python脚本构建解密映射表
func (d *FontDecoder) BuildMapping(fontData []byte) error {
	// 保存字体到临时文件
	tmpDir := os.TempDir()
	fontPath := filepath.Join(tmpDir, "fanqie_font.woff")
	if err := os.WriteFile(fontPath, fontData, 0644); err != nil {
		return fmt.Errorf("保存临时字体失败: %w", err)
	}
	defer os.Remove(fontPath)

	// 查找Python脚本
	scriptPath := d.findSolveScript()
	if scriptPath == "" {
		// 如果找不到脚本，使用内置简化逻辑
		return d.buildSimpleMapping(fontData)
	}

	// 调用Python脚本
	cmd := exec.Command("python3", scriptPath, fontPath)
	output, err := cmd.Output()
	if err != nil {
		// Python脚本失败，使用简化逻辑
		return d.buildSimpleMapping(fontData)
	}

	// 解析输出
	return d.LoadFromPythonOutput(output)
}

// findSolveScript 查找solve_font.py脚本
func (d *FontDecoder) findSolveScript() string {
	// 检查多个可能的位置
	paths := []string{
		"solve_font.py",
		"./solve_font.py",
		"../solve_font.py",
		"/home/xlei/project/xupu/solve_font.py",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// buildSimpleMapping 简化的映射构建（当Python不可用时）
func (d *FontDecoder) buildSimpleMapping(fontData []byte) error {
	// 这里可以添加简化的Go原生逻辑
	// 目前返回空映射，让原文直接显示
	d.mu.Lock()
	d.mapping = make(map[rune]rune)
	d.mu.Unlock()
	return nil
}

// LoadFromPythonOutput 从Python脚本输出加载映射
func (d *FontDecoder) LoadFromPythonOutput(data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mapping = make(map[rune]rune)

	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// 解析格式: 0xXXXX:字
		lineStr := string(line)
		parts := strings.SplitN(lineStr, ":", 2)
		if len(parts) != 2 {
			continue
		}

		codeStr := strings.TrimSpace(parts[0])
		charStr := strings.TrimSpace(parts[1])

		if len(codeStr) < 3 || len(charStr) == 0 {
			continue
		}

		// 解析十六进制编码
		var code int
		if _, err := fmt.Sscanf(codeStr, "0x%x", &code); err != nil {
			continue
		}

		// 获取第一个rune
		chars := []rune(charStr)
		if len(chars) > 0 {
			d.mapping[rune(code)] = chars[0]
		}
	}

	return nil
}

// Decrypt 解密文本
func (d *FontDecoder) Decrypt(encryptedText string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []rune
	for _, r := range encryptedText {
		if mapped, ok := d.mapping[r]; ok {
			result = append(result, mapped)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SetMapping 直接设置映射表
func (d *FontDecoder) SetMapping(mapping map[rune]rune) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.mapping = mapping
}

// GetMapping 获取当前映射表
func (d *FontDecoder) GetMapping() map[rune]rune {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[rune]rune)
	for k, v := range d.mapping {
		result[k] = v
	}
	return result
}

// MappingSize 返回映射表大小
func (d *FontDecoder) MappingSize() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.mapping)
}
