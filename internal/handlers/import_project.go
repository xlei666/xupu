package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
)

// ImportProject 导入项目
// @Summary 导入本地小说
// @Description 上传 TXT 文件并导入为项目
// @Tags projects
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "TXT文件"
// @Param author formData string false "作者"
// @Param description formData string false "简介"
// @Param cover formData file false "封面图片"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/import [post]
func (h *ProjectHandler) ImportProject(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_FILE", "未找到上传文件", err.Error()))
		return
	}

	// 检查文件类型
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".txt") {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_FILE_TYPE", "仅支持txt文件", ""))
		return
	}

	// 打开文件读取内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("READ_FAILED", "读取文件失败", err.Error()))
		return
	}
	defer src.Close()

	// 读取文件内容
	contentBytes := make([]byte, file.Size)
	if _, err := src.Read(contentBytes); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("READ_FAILED", "读取文件内容失败", err.Error()))
		return
	}
	content := string(contentBytes)

	// 获取用户ID
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "未授权", ""))
		return
	}

	// 解析项目名称（文件名去掉后缀）
	projectName := strings.TrimSuffix(file.Filename, ".txt")

	// 获取其他元数据
	author := c.PostForm("author")
	description := c.PostForm("description")

	// 处理封面上传
	var coverURL string
	coverFile, err := c.FormFile("cover")
	if err == nil {
		// 确保目录存在
		uploadDir := "static/uploads/covers"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			// 忽略错误，继续
		}

		// 生成唯一文件名
		ext := filepath.Ext(coverFile.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		coverFilename := fmt.Sprintf("cover_%s%s", db.GenerateID("img"), ext)
		coverPath := filepath.Join(uploadDir, coverFilename)

		if err := c.SaveUploadedFile(coverFile, coverPath); err == nil {
			coverURL = "/static/uploads/covers/" + coverFilename
		}
	}

	// 创建项目
	project := &models.Project{
		ID:          db.GenerateID("project"),
		UserID:      userID,
		Name:        projectName,
		Author:      author,
		Description: description,
		CoverURL:    coverURL,
		Mode:        models.OrchestrationMode("local_import"),
		Status:      models.StatusCompleted, // 导入的项目直接标记为完成
		Progress:    100,
	}

	// 解析章节
	// 简单的正则匹配：第X章 或 第X节
	chapterRegex := regexp.MustCompile(`\n\s*(第[0-9一二三四五六七八九十百千]+[章节][^\n]*)`)
	indexes := chapterRegex.FindAllStringIndex(content, -1)

	var chapters []*models.Chapter

	// 处理引子/序章（如果第一章之前有内容）
	if len(indexes) > 0 && indexes[0][0] > 0 {
		prefaceContent := strings.TrimSpace(content[:indexes[0][0]])
		if len(prefaceContent) > 0 {
			chapters = append(chapters, &models.Chapter{
				ID:         db.GenerateID("chapter"),
				ProjectID:  project.ID,
				Title:      "序章/引子",
				Content:    prefaceContent,
				ChapterNum: 0,
				Status:     models.ChapterStatusCompleted,
			})
		}
	}

	for i, idx := range indexes {
		// start := idx[0]
		end := len(content)
		if i < len(indexes)-1 {
			end = indexes[i+1][0]
		}

		// 提取标题和内容
		// idx[0] 是匹配到的开始，idx[1] 是匹配到的结束（即标题行的结束）
		// 但这里的正则包含了 \n，所以我们需要更精确地提取标题

		fullMatch := content[idx[0]:idx[1]]
		title := strings.TrimSpace(fullMatch)
		chapterContent := strings.TrimSpace(content[idx[1]:end])

		chapters = append(chapters, &models.Chapter{
			ID:         db.GenerateID("chapter"),
			ProjectID:  project.ID,
			Title:      title,
			Content:    chapterContent,
			ChapterNum: len(chapters) + 1,
			Status:     models.ChapterStatusCompleted,
		})
	}

	// 如果正则没匹配到任何章节，将整个文件作为一个章节
	if len(chapters) == 0 {
		chapters = append(chapters, &models.Chapter{
			ID:         db.GenerateID("chapter"),
			ProjectID:  project.ID,
			Title:      projectName,
			Content:    content,
			ChapterNum: 1,
			Status:     models.ChapterStatusCompleted,
		})
	}

	// 保存项目和章节
	if err := db.Get().SaveProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存项目失败", err.Error()))
		return
	}

	for _, chap := range chapters {
		if err := db.Get().SaveChapter(chap); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存章节失败", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"project_id":    project.ID,
		"chapter_count": len(chapters),
	}))
}
