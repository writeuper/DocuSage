package models

import (
	"time"

	"gorm.io/gorm"
)

// Query 搜索查询记录
type Query struct {
	gorm.Model
	UserID      uint      `gorm:"not null" json:"user_id"`
	QueryText   string    `gorm:"type:text;not null" json:"query_text"`
	SearchType  string    `gorm:"type:varchar(50);not null;default:'keyword'" json:"search_type"` // keyword, semantic, hybrid
	Filters     string    `gorm:"type:text" json:"filters"`                                       // JSON格式的过滤条件
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	Error       string    `gorm:"type:text" json:"error"`
	ResultCount int       `gorm:"default:0" json:"result_count"`
}

// SearchResult 搜索结果
type SearchResult struct {
	gorm.Model
	QueryID    uint    `gorm:"not null;index" json:"query_id"`
	DocumentID string  `gorm:"type:varchar(255);not null;index" json:"document_id"`
	Score      float64 `gorm:"not null" json:"score"`
	Title      string  `gorm:"type:text" json:"title"`
	Content    string  `gorm:"type:text" json:"content"`
	PageNumber int     `json:"page_number"`
	Source     string  `gorm:"type:varchar(100)" json:"source"`

	// 关联
	Query Query `gorm:"foreignKey:QueryID" json:"query,omitempty"`
}

// SearchHistory 用户搜索历史
type SearchHistory struct {
	gorm.Model
	UserID    uint   `gorm:"not null;uniqueIndex:idx_user_query" json:"user_id"`
	QueryText string `gorm:"type:varchar(512);not null;uniqueIndex:idx_user_query" json:"query_text"`
}

// QueryLog 查询日志
type QueryLog struct {
	gorm.Model
	UserID       uint   `gorm:"index" json:"user_id"`
	QueryText    string `gorm:"type:text" json:"query_text"`
	IP           string `gorm:"type:varchar(50)" json:"ip"`
	UserAgent    string `gorm:"type:text" json:"user_agent"`
	ResponseTime int64  `json:"response_time"` // 毫秒
	Status       string `gorm:"type:varchar(20)" json:"status"`
}

// QueryStats 查询统计
type QueryStats struct {
	gorm.Model
	UserID         uint      `gorm:"index" json:"user_id"`
	DailyQueries   int       `gorm:"default:0" json:"daily_queries"`
	WeeklyQueries  int       `gorm:"default:0" json:"weekly_queries"`
	MonthlyQueries int       `gorm:"default:0" json:"monthly_queries"`
	TotalQueries   int64     `gorm:"default:0" json:"total_queries"`
	LastQueryAt    time.Time `json:"last_query_at"`
}

// SearchRequest 搜索请求模型
type SearchRequest struct {
	Query     string                 `json:"query" binding:"required"`
	Type      string                 `json:"type" binding:"omitempty,oneof=keyword semantic hybrid"`
	Filters   map[string]interface{} `json:"filters"`
	Limit     int                    `json:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int                    `json:"offset" binding:"omitempty,min=0"`
	SortBy    string                 `json:"sort_by" binding:"omitempty,oneof=relevance time"`
	SortOrder string                 `json:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// SearchResponse 搜索响应模型
type SearchResponse struct {
	QueryID        uint           `json:"query_id"`
	Total          int            `json:"total"`
	Limit          int            `json:"limit"`
	Offset         int            `json:"offset"`
	Results        []SearchResult `json:"results"`
	ProcessingTime int64          `json:"processing_time"` // 毫秒
}

// QueryAnalytics 搜索分析请求模型
type QueryAnalytics struct {
	UserID     *uint      `json:"user_id"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	GroupBy    string     `json:"group_by" binding:"omitempty,oneof=day week month user"`
	SearchType string     `json:"search_type"`
}

// AnalyticsResponse 分析响应模型
type AnalyticsResponse struct {
	TotalQueries        int               `json:"total_queries"`
	UniqueUsers         int               `json:"unique_users"`
	AverageResponseTime float64           `json:"average_response_time"`
	QueryBreakdown      map[string]int    `json:"query_breakdown"`
	TimeSeriesData      []TimeSeriesPoint `json:"time_series_data"`
	TopQueries          []TopQuery        `json:"top_queries"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Time  time.Time `json:"time"`
	Count int       `json:"count"`
}

// TopQuery 热门查询
type TopQuery struct {
	QueryText string `json:"query_text"`
	Count     int    `json:"count"`
}

// UserInfo 用户信息（从认证服务获取）
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID *uint  `json:"company_id"`
}
