package crawler

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"io"
	"net/http"
	"sort"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontDecoder 字体解码器，用于解密番茄小说的自定义字体
type FontDecoder struct {
	mapping map[rune]rune // PUA编码 -> 标准汉字
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

// BuildMapping 从字体数据构建解密映射表
func (d *FontDecoder) BuildMapping(fontData []byte) error {
	// 解析字体
	f, err := opentype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("解析字体失败: %w", err)
	}

	// 创建字体face用于渲染
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("创建字体face失败: %w", err)
	}
	defer face.Close()

	// 收集字体中的PUA编码（私有使用区域 E000-F8FF）
	puaCodes := d.collectPUACodes(f)
	if len(puaCodes) == 0 {
		return fmt.Errorf("字体中未找到PUA编码")
	}

	// 预渲染PUA字形
	puaGlyphs := make(map[rune]*image.Gray)
	for _, code := range puaCodes {
		glyph := d.renderGlyph(face, code)
		if glyph != nil {
			puaGlyphs[code] = glyph
		}
	}

	// 加载参考字体并匹配
	refMapping, err := d.matchWithReferenceChars(puaGlyphs)
	if err != nil {
		return fmt.Errorf("匹配参考字符失败: %w", err)
	}

	d.mu.Lock()
	d.mapping = refMapping
	d.mu.Unlock()

	return nil
}

// collectPUACodes 收集字体中的PUA编码
func (d *FontDecoder) collectPUACodes(f *opentype.Font) []rune {
	var codes []rune

	// 扫描PUA区域 (E000-F8FF)
	for code := rune(0xE000); code <= rune(0xF8FF); code++ {
		idx, err := f.GlyphIndex(nil, code)
		if err == nil && idx != 0 {
			codes = append(codes, code)
		}
	}

	return codes
}

// renderGlyph 渲染单个字形到灰度图像
func (d *FontDecoder) renderGlyph(face font.Face, r rune) *image.Gray {
	// 获取字形边界
	bounds, advance, ok := face.GlyphBounds(r)
	if !ok || advance == 0 {
		return nil
	}

	// 创建图像
	width := (bounds.Max.X - bounds.Min.X).Ceil() + 4
	height := (bounds.Max.Y - bounds.Min.Y).Ceil() + 4
	if width <= 0 || height <= 0 {
		width, height = 40, 40
	}

	img := image.NewGray(image.Rect(0, 0, width, height))

	// 绘制背景
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// 绘制字形
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(2) - bounds.Min.X, Y: fixed.I(height-2) - bounds.Max.Y},
	}
	drawer.DrawString(string(r))

	return img
}

// matchWithReferenceChars 使用常用汉字进行匹配
func (d *FontDecoder) matchWithReferenceChars(puaGlyphs map[rune]*image.Gray) (map[rune]rune, error) {
	mapping := make(map[rune]rune)

	// 常用汉字列表（约3500个常用字）
	commonChars := getCommonChineseChars()

	// 由于Go标准库没有内置中文字体，我们使用简化的特征匹配
	// 实际项目中可能需要加载系统字体或嵌入字体文件

	// 简化方案：基于像素统计和形状特征匹配
	for puaCode, puaGlyph := range puaGlyphs {
		bestMatch := d.findBestMatch(puaGlyph, commonChars)
		if bestMatch != 0 {
			mapping[puaCode] = bestMatch
		}
	}

	return mapping, nil
}

// findBestMatch 基于图像特征找到最佳匹配
func (d *FontDecoder) findBestMatch(targetGlyph *image.Gray, candidates []rune) rune {
	// 计算目标字形的特征
	targetFeatures := extractFeatures(targetGlyph)

	// 这里使用简化的匹配逻辑
	// 实际实现需要加载参考字体并比较

	// 暂时返回占位符 - 实际实现需要完整的字体比较
	_ = targetFeatures
	_ = candidates

	return 0
}

// extractFeatures 提取图像特征用于匹配
func extractFeatures(img *image.Gray) []float64 {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算基本特征：密度、重心、象限分布
	var totalPixels, blackPixels float64
	var sumX, sumY float64

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			totalPixels++
			if img.GrayAt(x, y).Y < 128 {
				blackPixels++
				sumX += float64(x)
				sumY += float64(y)
			}
		}
	}

	if blackPixels == 0 {
		return []float64{0, 0, 0}
	}

	density := blackPixels / totalPixels
	centerX := sumX / blackPixels / float64(width)
	centerY := sumY / blackPixels / float64(height)

	return []float64{density, centerX, centerY}
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

// SetMapping 直接设置映射表（用于加载预计算的映射）
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

// LoadFromPythonOutput 从Python脚本输出加载映射
// 格式: "0xE3E8:字\n0xE3E9:符\n..."
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
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			continue
		}

		codeStr := string(bytes.TrimSpace(parts[0]))
		charStr := string(bytes.TrimSpace(parts[1]))

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

// getCommonChineseChars 返回常用汉字列表
func getCommonChineseChars() []rune {
	// 常用3500字（按使用频率排序的前500个作为示例）
	common := "的一是不了在人有我他这个们中来上大为和国地到以说" +
		"时要就出会可也你对生能而子那得于着下自之年过发后作里" +
		"用道行所然家种事成方多经么去法学如都同现当没动面起看" +
		"定天分还进好小部其些主样理心她本前开但因只从想实日军" +
		"者意无力它与长把机十民第公此已工使情明性知全三又关点" +
		"正业外将两高间由问很最也重新国电回神给等被走北水几月" +
		"身孩做界门利头己女西身斯德克阿那各于如战儿你死位西世" +
		"山白口感放热爱路母色手完气边表解间许话张接什少内真已" +
		"教别果特平通声万代车太认让信报吓风先像打原听步老处" +
		"候考刻任周边元四相领息区叫活死量名次条系虽单直眼轻应" +
		"根济切往术统住每林东黑半武空运算级站结金怕选派却食至" +
		"红收服试型况件落容飞百拿设求响决料持续反城布深局且联" +
		"备包离字紧计极究京命哪满基共规离展议参格江资市般调习" +
		"农村强何党快更思拉科县城呢题难提质易象医石官视跟总示" +
		"流责形史除保华语增首护建压影希望员组章送复米故支卫该" +
		"集言须古南效假记断七志商节转团史顺病达程低富备细丽" +
		"写球某权求识技引状验环突夜跑速随功差土治青确团青讲段" +
		"供局双座客南险约育阳产演维育准存纪态余停府标终额临越" +
		"观列值终整树楼杀推获究温毛止永适争严省初推亲印改击究" +
		"承识药诉维持围传批破推众演举亿察投领脸岁副编观居既待"

	chars := []rune(common)

	// 去重并排序
	seen := make(map[rune]bool)
	var unique []rune
	for _, c := range chars {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}

	sort.Slice(unique, func(i, j int) bool {
		return unique[i] < unique[j]
	})

	return unique
}
