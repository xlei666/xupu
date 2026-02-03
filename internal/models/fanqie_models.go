package models

// FanqieBookDetail 书籍详情
type FanqieBookDetail struct {
	BookID         string  `json:"bookId"`
	BookName       string  `json:"bookName"`
	Author         string  `json:"author"`
	AuthorID       string  `json:"authorId"`
	ThumbURI       string  `json:"thumbUri"`
	Category       string  `json:"category"`
	Score          float64 `json:"score"`
	Description    string  `json:"description"`
	WordCount      int     `json:"wordCount"`
	ChapterCount   int     `json:"chapterCount"`
	CreationStatus int     `json:"creationStatus"` // 1=连载中, 2=已完结
	LastChapterID  string  `json:"lastChapterId"`
	LastChapter    string  `json:"lastChapter"`
	UpdateTime     string  `json:"updateTime"`
}

// FanqieChapter 章节信息
type FanqieChapter struct {
	ChapterID  string `json:"chapterId"`
	Title      string `json:"title"`
	Order      int    `json:"order"`
	WordCount  int    `json:"wordCount"`
	CreateTime string `json:"createTime"`
	IsVIP      bool   `json:"isVip"`
	VolumeName string `json:"volumeName,omitempty"`
}

// FanqieChapterContent 章节内容
type FanqieChapterContent struct {
	ChapterID   string `json:"chapterId"`
	BookID      string `json:"bookId"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	WordCount   int    `json:"wordCount"`
	PrevChapter string `json:"prevChapter,omitempty"`
	NextChapter string `json:"nextChapter,omitempty"`
}
