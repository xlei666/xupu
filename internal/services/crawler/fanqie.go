package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/xlei/xupu/internal/models"
)

const (
	// Mobile API Endpoint does not require complex a_bogus signature for this data
	FanqieMobileBaseURL = "https://api-lf.fanqiesdk.com/api/novel/channel/homepage/rank/rank_list/v2/"
	// 书籍详情API
	FanqieBookDetailURL = "https://api-lf.fanqiesdk.com/api/novel/book/detail/v1/"
	// 章节列表API
	FanqieChapterListURL = "https://api-lf.fanqiesdk.com/api/novel/book/directory/list/v1/"
	// 章节内容API (Web)
	FanqieReaderURL = "https://fanqienovel.com/reader/"
	// 移动端用户代理
	AndroidUserAgent = "Dalvik/2.1.0 (Linux; U; Android 10; Pixel 4 Build/QD1A.190821.011)"
	// Web用户代理
	WebUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type FanqieService struct {
	client      *http.Client
	fontDecoder *FontDecoder
	fontCache   map[string]*FontDecoder // 缓存不同字体的解码器
	fontCacheMu sync.RWMutex
}

func NewFanqieService() *FanqieService {
	return &FanqieService{
		client:      &http.Client{},
		fontDecoder: NewFontDecoder(),
		fontCache:   make(map[string]*FontDecoder),
	}
}

// GetRankList fetches the leaderboard using the Mobile API endpoint to bypass web-only anti-scraping
func (s *FanqieService) GetRankList(categoryID string) ([]models.FanqieBook, error) {
	params := url.Values{}
	params.Add("aid", "1967") // Common App ID for Fanqie Android
	params.Add("channel", "0")
	params.Add("device_platform", "android")
	params.Add("device_type", "0")
	params.Add("limit", "20")
	params.Add("offset", "0")

	// Map category_id to side_type or use as is
	// Default generic rank seems to be side_type=15
	if categoryID == "" || categoryID == "1" {
		params.Add("side_type", "15")
	} else {
		params.Add("side_type", categoryID)
	}
	params.Add("type", "1")

	reqURL := fmt.Sprintf("%s?%s", FanqieMobileBaseURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mimic Android Client
	req.Header.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 10; Pixel 4 Build/QD1A.190821.011)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Internal structs for parsing Mobile API response
	type MobileBookItem struct {
		BookID   string `json:"book_id"`
		BookName string `json:"book_name"`
		Author   string `json:"author"`
		Score    string `json:"score"`
		Abstract string `json:"abstract"`
		ThumbURL string `json:"thumb_url"`
		Category string `json:"category"`
	}

	type MobileData struct {
		Result []MobileBookItem `json:"result"`
	}

	type MobileResponse struct {
		Code int        `json:"code"`
		Data MobileData `json:"data"`
		Msg  string     `json:"message"`
	}

	var result MobileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("api error code: %d, message: %s", result.Code, result.Msg)
	}

	var books []models.FanqieBook
	for _, item := range result.Data.Result {
		var score float64
		fmt.Sscanf(item.Score, "%f", &score)

		books = append(books, models.FanqieBook{
			BookID:      item.BookID,
			BookName:    item.BookName,
			Author:      item.Author,
			Category:    item.Category,
			Score:       score,
			ReadCount:   0, // Not available in this list view
			Description: item.Abstract,
			ThumbURI:    item.ThumbURL,
		})
	}

	return books, nil
}

// GetBookDetail 获取书籍详情
func (s *FanqieService) GetBookDetail(bookID string) (*models.FanqieBookDetail, error) {
	params := url.Values{}
	params.Add("aid", "1967")
	params.Add("book_id", bookID)
	params.Add("device_platform", "android")

	reqURL := fmt.Sprintf("%s?%s", FanqieBookDetailURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", AndroidUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	type BookDetailData struct {
		BookID         string `json:"book_id"`
		BookName       string `json:"book_name"`
		Author         string `json:"author"`
		AuthorID       string `json:"author_id"`
		ThumbURL       string `json:"thumb_url"`
		Category       string `json:"category"`
		Abstract       string `json:"abstract"`
		WordCount      string `json:"word_count"`
		SerialCount    string `json:"serial_count"`
		CreationStatus string `json:"creation_status"`
		LastChapterID  string `json:"last_item_id"`
		LastChapter    string `json:"last_chapter_title"`
		UpdateTime     string `json:"last_chapter_update_time"`
	}

	type BookDetailResponse struct {
		Code int            `json:"code"`
		Data BookDetailData `json:"data"`
		Msg  string         `json:"message"`
	}

	var result BookDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API错误: %d, %s", result.Code, result.Msg)
	}

	wordCount, _ := strconv.Atoi(result.Data.WordCount)
	chapterCount, _ := strconv.Atoi(result.Data.SerialCount)
	creationStatus, _ := strconv.Atoi(result.Data.CreationStatus)

	return &models.FanqieBookDetail{
		BookID:         result.Data.BookID,
		BookName:       result.Data.BookName,
		Author:         result.Data.Author,
		AuthorID:       result.Data.AuthorID,
		ThumbURI:       result.Data.ThumbURL,
		Category:       result.Data.Category,
		Description:    result.Data.Abstract,
		WordCount:      wordCount,
		ChapterCount:   chapterCount,
		CreationStatus: creationStatus,
		LastChapterID:  result.Data.LastChapterID,
		LastChapter:    result.Data.LastChapter,
		UpdateTime:     result.Data.UpdateTime,
	}, nil
}

// GetChapterList 获取章节列表
func (s *FanqieService) GetChapterList(bookID string) ([]models.FanqieChapter, error) {
	params := url.Values{}
	params.Add("aid", "1967")
	params.Add("book_id", bookID)
	params.Add("device_platform", "android")

	reqURL := fmt.Sprintf("%s?%s", FanqieChapterListURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", AndroidUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	type ChapterItem struct {
		ItemID     string `json:"item_id"`
		Title      string `json:"title"`
		Order      int    `json:"order"`
		WordCount  string `json:"word_count"`
		IsVIP      int    `json:"is_paid"`
		VolumeName string `json:"volume_name"`
	}

	type ChapterListData struct {
		ItemIdsList []ChapterItem `json:"item_ids_list"`
	}

	type ChapterListResponse struct {
		Code int             `json:"code"`
		Data ChapterListData `json:"data"`
		Msg  string          `json:"message"`
	}

	var result ChapterListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API错误: %d, %s", result.Code, result.Msg)
	}

	var chapters []models.FanqieChapter
	for i, item := range result.Data.ItemIdsList {
		wordCount, _ := strconv.Atoi(item.WordCount)
		chapters = append(chapters, models.FanqieChapter{
			ChapterID:  item.ItemID,
			Title:      item.Title,
			Order:      i + 1,
			WordCount:  wordCount,
			IsVIP:      item.IsVIP == 1,
			VolumeName: item.VolumeName,
		})
	}

	return chapters, nil
}

// GetChapterContent 获取章节内容（从Web页面解析）
func (s *FanqieService) GetChapterContent(chapterID string) (*models.FanqieChapterContent, error) {
	reqURL := FanqieReaderURL + chapterID

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	// 增加Web端伪装
	req.Header.Set("User-Agent", WebUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", reqURL)
	req.Header.Set("Cache-Control", "no-cache")

	// 注入Cookie (如果有)
	if cookie := os.Getenv("FANQIE_COOKIE"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	htmlContent := string(body)

	// 提取 window.__INITIAL_STATE__ JSON
	stateRe := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*(\{.*?\});`)
	matches := stateRe.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return nil, fmt.Errorf("无法找到页面状态数据")
	}

	// 清理JSON字符串：将undefined替换为null，否则Go无法解析
	jsonStr := matches[1]
	jsonStr = strings.ReplaceAll(jsonStr, "undefined", "null")

	// 解析JSON
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &state); err != nil {
		return nil, fmt.Errorf("解析状态JSON失败: %w", err)
	}

	// 提取章节数据
	reader, ok := state["reader"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无法找到reader数据")
	}

	chapterData, ok := reader["chapterData"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无法找到chapterData数据")
	}

	content, ok := chapterData["content"].(string)
	if !ok || content == "" {
		// 检查是否有错误信息或其他提示
		return nil, fmt.Errorf("章节内容为空 (可能是反爬限制)")
	}

	title, _ := chapterData["title"].(string)
	bookID, _ := chapterData["bookId"].(string)
	prevChapter, _ := chapterData["preItemId"].(string)
	nextChapter, _ := chapterData["nextItemId"].(string)
	wordCountStr, _ := chapterData["chapterWordNumber"].(string)
	wordCount, _ := strconv.Atoi(wordCountStr)

	// 提取字体CSS URL并解密内容
	fontURL := s.extractFontURL(htmlContent)
	decryptedContent := content

	if fontURL != "" {
		decoder, err := s.getOrCreateDecoder(fontURL)
		if err == nil && decoder != nil {
			decryptedContent = decoder.Decrypt(content)
		}
	}

	// 清理HTML标签
	decryptedContent = s.cleanHTMLContent(decryptedContent)

	return &models.FanqieChapterContent{
		ChapterID:   chapterID,
		BookID:      bookID,
		Title:       title,
		Content:     decryptedContent,
		WordCount:   wordCount,
		PrevChapter: prevChapter,
		NextChapter: nextChapter,
	}, nil
}

// extractFontURL 从HTML中提取字体URL
func (s *FanqieService) extractFontURL(html string) string {
	// 查找字体CSS中的字体URL
	fontRe := regexp.MustCompile(`url\((https://[^)]+\.woff2?)\)`)
	matches := fontRe.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// getOrCreateDecoder 获取或创建字体解码器
func (s *FanqieService) getOrCreateDecoder(fontURL string) (*FontDecoder, error) {
	s.fontCacheMu.RLock()
	if decoder, ok := s.fontCache[fontURL]; ok {
		s.fontCacheMu.RUnlock()
		return decoder, nil
	}
	s.fontCacheMu.RUnlock()

	// 创建新的解码器
	decoder := NewFontDecoder()
	if err := decoder.LoadFromURL(fontURL); err != nil {
		return nil, err
	}

	s.fontCacheMu.Lock()
	s.fontCache[fontURL] = decoder
	s.fontCacheMu.Unlock()

	return decoder, nil
}

// cleanHTMLContent 清理HTML内容
func (s *FanqieService) cleanHTMLContent(content string) string {
	// 移除HTML标签
	content = strings.ReplaceAll(content, "<p>", "\n")
	content = strings.ReplaceAll(content, "</p>", "")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "&nbsp;", " ")

	// 移除其他HTML标签
	tagRe := regexp.MustCompile(`<[^>]+>`)
	content = tagRe.ReplaceAllString(content, "")

	// 清理多余空行
	content = strings.TrimSpace(content)

	return content
}
