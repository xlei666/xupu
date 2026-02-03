package handlers

import (
	"net/http"

	"github.com/xlei/xupu/internal/services/crawler"

	"github.com/gin-gonic/gin"
)

type ExternalRankHandler struct {
	fanqieService *crawler.FanqieService
}

func NewExternalRankHandler() *ExternalRankHandler {
	return &ExternalRankHandler{
		fanqieService: crawler.NewFanqieService(),
	}
}

// GetFanqieRank handles GET /api/v1/external/ranks/fanqie
func (h *ExternalRankHandler) GetFanqieRank(c *gin.Context) {
	categoryID := c.Query("category_id")
	if categoryID == "" {
		categoryID = "15"
	}

	books, err := h.fanqieService.GetRankList(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "获取番茄小说排行榜失败",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"source": "fanqienovel.com",
			"books":  books,
		},
	})
}

// GetFanqieBookDetail handles GET /api/v1/external/fanqie/books/:bookId
func (h *ExternalRankHandler) GetFanqieBookDetail(c *gin.Context) {
	bookID := c.Param("bookId")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PARAM",
				"message": "缺少书籍ID参数",
			},
		})
		return
	}

	book, err := h.fanqieService.GetBookDetail(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "获取书籍详情失败",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    book,
	})
}

// GetFanqieChapterList handles GET /api/v1/external/fanqie/books/:bookId/chapters
func (h *ExternalRankHandler) GetFanqieChapterList(c *gin.Context) {
	bookID := c.Param("bookId")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PARAM",
				"message": "缺少书籍ID参数",
			},
		})
		return
	}

	chapters, err := h.fanqieService.GetChapterList(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "获取章节列表失败",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bookId":   bookID,
			"chapters": chapters,
			"total":    len(chapters),
		},
	})
}

// GetFanqieChapterContent handles GET /api/v1/external/fanqie/chapters/:chapterId
func (h *ExternalRankHandler) GetFanqieChapterContent(c *gin.Context) {
	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PARAM",
				"message": "缺少章节ID参数",
			},
		})
		return
	}

	content, err := h.fanqieService.GetChapterContent(chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "获取章节内容失败",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    content,
	})
}
