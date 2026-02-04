package crawler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/image/font/sfnt"
)

// FontDecoder 字体解码器，纯Go实现
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

// BuildMapping 从字体数据构建解密映射表
// 原理：番茄字体中同一个字形(glyph)同时被PUA字符和CJK字符引用
// 通过解析cmap表找到这种共享关系即可建立解密映射
func (d *FontDecoder) BuildMapping(fontData []byte) error {
	// 使用sfnt库解析字体获取字形到字符的映射
	f, err := sfnt.Parse(fontData)
	if err != nil {
		return fmt.Errorf("解析字体失败: %w", err)
	}

	// 构建字形ID到字符的映射
	glyphToChars := make(map[sfnt.GlyphIndex][]rune)
	buf := &sfnt.Buffer{}

	// 只扫描PUA和CJK范围
	ranges := []struct{ start, end rune }{
		{0xE000, 0xF8FF}, // PUA
		{0x4E00, 0x9FFF}, // CJK
	}

	for _, r := range ranges {
		for code := r.start; code <= r.end; code++ {
			idx, err := f.GlyphIndex(buf, code)
			if err == nil && idx != 0 {
				glyphToChars[idx] = append(glyphToChars[idx], code)
			}
		}
	}

	// 构建PUA到CJK的映射
	mapping := make(map[rune]rune)
	for _, chars := range glyphToChars {
		var puaChar, cjkChar rune
		for _, c := range chars {
			if c >= 0xE000 && c <= 0xF8FF {
				puaChar = c
			} else if c >= 0x4E00 && c <= 0x9FFF {
				cjkChar = c
			}
		}
		if puaChar != 0 && cjkChar != 0 {
			mapping[puaChar] = cjkChar
		}
	}

	d.mu.Lock()
	d.mapping = mapping
	d.mu.Unlock()

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

// LoadFromPythonOutput 兼容Python脚本输出格式
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
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			continue
		}
		var code int
		if _, err := fmt.Sscanf(string(parts[0]), "0x%x", &code); err != nil {
			continue
		}
		chars := []rune(string(bytes.TrimSpace(parts[1])))
		if len(chars) > 0 {
			d.mapping[rune(code)] = chars[0]
		}
	}
	return nil
}

// 保留手动cmap解析作为备用（当sfnt解析失败时）
func (d *FontDecoder) parseCmapManually(data []byte) (map[uint16][]rune, error) {
	r := bytes.NewReader(data)

	var offsetTable struct {
		SfntVersion   uint32
		NumTables     uint16
		SearchRange   uint16
		EntrySelector uint16
		RangeShift    uint16
	}
	if err := binary.Read(r, binary.BigEndian, &offsetTable); err != nil {
		return nil, err
	}

	var cmapOffset, cmapLength uint32
	for i := 0; i < int(offsetTable.NumTables); i++ {
		var entry struct {
			Tag      [4]byte
			Checksum uint32
			Offset   uint32
			Length   uint32
		}
		if err := binary.Read(r, binary.BigEndian, &entry); err != nil {
			return nil, err
		}
		if string(entry.Tag[:]) == "cmap" {
			cmapOffset = entry.Offset
			cmapLength = entry.Length
			break
		}
	}

	if cmapOffset == 0 {
		return nil, fmt.Errorf("未找到cmap表")
	}

	return d.parseCmapTable(data[cmapOffset : cmapOffset+cmapLength])
}

func (d *FontDecoder) parseCmapTable(data []byte) (map[uint16][]rune, error) {
	r := bytes.NewReader(data)

	var header struct {
		Version   uint16
		NumTables uint16
	}
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return nil, err
	}

	glyphToChars := make(map[uint16][]rune)

	for i := 0; i < int(header.NumTables); i++ {
		var record struct {
			PlatformID uint16
			EncodingID uint16
			Offset     uint32
		}
		if err := binary.Read(r, binary.BigEndian, &record); err != nil {
			return nil, err
		}

		subtableData := data[record.Offset:]
		if len(subtableData) < 2 {
			continue
		}
		format := binary.BigEndian.Uint16(subtableData[:2])

		switch format {
		case 4:
			d.parseFormat4(subtableData, glyphToChars)
		case 12:
			d.parseFormat12(subtableData, glyphToChars)
		}
	}

	return glyphToChars, nil
}

func (d *FontDecoder) parseFormat4(data []byte, result map[uint16][]rune) {
	if len(data) < 14 {
		return
	}
	segCount := binary.BigEndian.Uint16(data[6:8]) / 2
	endOffset := 14
	startOffset := endOffset + int(segCount)*2 + 2
	deltaOffset := startOffset + int(segCount)*2
	rangeOffset := deltaOffset + int(segCount)*2

	for i := 0; i < int(segCount); i++ {
		if endOffset+i*2+2 > len(data) || startOffset+i*2+2 > len(data) {
			break
		}
		endCode := binary.BigEndian.Uint16(data[endOffset+i*2:])
		startCode := binary.BigEndian.Uint16(data[startOffset+i*2:])
		idDelta := int16(binary.BigEndian.Uint16(data[deltaOffset+i*2:]))
		idRangeOffset := binary.BigEndian.Uint16(data[rangeOffset+i*2:])

		if startCode == 0xFFFF {
			break
		}

		for c := startCode; c <= endCode; c++ {
			var glyphID uint16
			if idRangeOffset == 0 {
				glyphID = uint16(int16(c) + idDelta)
			} else {
				idx := rangeOffset + i*2 + int(idRangeOffset) + int(c-startCode)*2
				if idx+2 <= len(data) {
					glyphID = binary.BigEndian.Uint16(data[idx:])
					if glyphID != 0 {
						glyphID = uint16(int16(glyphID) + idDelta)
					}
				}
			}
			if glyphID != 0 {
				result[glyphID] = append(result[glyphID], rune(c))
			}
		}
	}
}

func (d *FontDecoder) parseFormat12(data []byte, result map[uint16][]rune) {
	if len(data) < 16 {
		return
	}
	numGroups := binary.BigEndian.Uint32(data[12:16])
	for i := uint32(0); i < numGroups && int(16+i*12+12) <= len(data); i++ {
		offset := 16 + i*12
		startCode := binary.BigEndian.Uint32(data[offset:])
		endCode := binary.BigEndian.Uint32(data[offset+4:])
		startGlyph := binary.BigEndian.Uint32(data[offset+8:])
		for c := startCode; c <= endCode; c++ {
			glyphID := uint16(startGlyph + (c - startCode))
			result[glyphID] = append(result[glyphID], rune(c))
		}
	}
}
