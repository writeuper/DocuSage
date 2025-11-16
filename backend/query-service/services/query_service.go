package services

import (
	"context"
	"docusage/query-service/config"
	"docusage/query-service/models"
	"docusage/query-service/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// QueryService 查询服务
type QueryService struct {
	db     *gorm.DB
	config *config.Config
	log    *zap.Logger
}

// NewQueryService 创建查询服务实例
func NewQueryService(db *gorm.DB, cfg *config.Config, log *zap.Logger) *QueryService {
	return &QueryService{
		db:     db,
		config: cfg,
		log:    log,
	}
}

// Search 执行搜索
func (s *QueryService) Search(ctx context.Context, req models.SearchRequest, userID uint) (*models.SearchResponse, error) {
	// 记录开始时间
	startTime := time.Now()

	// 验证并设置默认值
	if req.Type == "" {
		req.Type = "keyword"
	}
	if req.Limit == 0 {
		req.Limit = s.config.DefaultSearchLimit
	}
	if req.Limit > s.config.MaxSearchResults {
		req.Limit = s.config.MaxSearchResults
	}

	// 生成缓存键
	cacheKey := s.generateCacheKey(userID, req)

	// 尝试从缓存获取
	var cachedResponse models.SearchResponse
	err := utils.GetCachedSearchResult(ctx, cacheKey, &cachedResponse)
	if err == nil {
		s.log.Debug("Cache hit for search query", zap.String("query", req.Query))
		return &cachedResponse, nil
	}

	// 创建查询记录
	query := models.Query{
		UserID:     userID,
		QueryText:  req.Query,
		SearchType: req.Type,
		StartTime:  time.Now(),
		Status:     "processing",
	}

	// 将过滤器转换为JSON字符串
	if req.Filters != nil {
		filtersJSON, _ := json.Marshal(req.Filters)
		query.Filters = string(filtersJSON)
	}

	if err := s.db.Create(&query).Error; err != nil {
		s.log.Error("Failed to create query record", zap.Error(err))
		return nil, fmt.Errorf("failed to create query record: %w", err)
	}

	// 执行搜索逻辑（模拟）
	results, total, err := s.executeSearch(ctx, req, userID)
	if err != nil {
		// 更新查询状态为失败
		s.db.Model(&query).Updates(map[string]interface{}{
			"status":   "failed",
			"error":    err.Error(),
			"end_time": time.Now(),
		})
		return nil, err
	}

	// 更新查询状态为成功
	s.db.Model(&query).Updates(map[string]interface{}{
		"status":       "completed",
		"end_time":     time.Now(),
		"result_count": len(results),
	})

	// 保存搜索结果到数据库
	if err := s.saveSearchResults(ctx, query.ID, results); err != nil {
		s.log.Warn("Failed to save search results", zap.Error(err))
		// 不影响主流程
	}

	// 记录搜索历史
	if err := utils.RecordSearchHistory(ctx, userID, req.Query, s.config.SearchHistoryMaxItems); err != nil {
		s.log.Warn("Failed to record search history", zap.Error(err))
	}

	// 构建响应
	response := &models.SearchResponse{
		QueryID:        query.ID,
		Total:          total,
		Limit:          req.Limit,
		Offset:         req.Offset,
		Results:        results,
		ProcessingTime: int64(time.Since(startTime).Milliseconds()),
	}

	// 缓存结果（10分钟）
	if err := utils.CacheSearchResult(ctx, cacheKey, response, 10*time.Minute); err != nil {
		s.log.Warn("Failed to cache search result", zap.Error(err))
	}

	return response, nil
}

// executeSearch 执行实际的搜索逻辑
func (s *QueryService) executeSearch(ctx context.Context, req models.SearchRequest, userID uint) ([]models.SearchResult, int, error) {
	// 根据搜索类型选择不同的搜索逻辑
	switch req.Type {
	case "semantic":
		return s.executeSemanticSearch(ctx, req, userID)
	case "hybrid":
		// 混合搜索：先执行语义搜索，再执行关键词搜索，最后合并结果
		semanticResults, semanticTotal, err := s.executeSemanticSearch(ctx, req, userID)
		if err != nil {
			s.log.Warn("Semantic search failed, falling back to keyword search", zap.Error(err))
			return s.executeKeywordSearch(ctx, req, userID)
		}

		// 这里简化处理，仅返回语义搜索结果
		// 实际项目中应实现结果合并和排序逻辑
		return semanticResults, semanticTotal, nil
	case "keyword":
		return s.executeKeywordSearch(ctx, req, userID)
	default:
		// 默认使用关键词搜索
		return s.executeKeywordSearch(ctx, req, userID)
	}
}

// executeSemanticSearch 执行语义搜索
func (s *QueryService) executeSemanticSearch(ctx context.Context, req models.SearchRequest, userID uint) ([]models.SearchResult, int, error) {
	// 准备请求体
	searchBody := map[string]interface{}{
		"query":   req.Query,
		"top_k":   req.Limit,
		"filters": req.Filters,
		"user_id": userID,
	}

	bodyJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal search request: %w", err)
	}

	// 构建请求URL
	requestURL := fmt.Sprintf("%s/api/v1/search/semantic", s.config.VectorSearchServiceAddr)
	s.log.Debug("Executing semantic search", zap.String("url", requestURL), zap.String("query", req.Query))

	// 创建HTTP请求
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create search request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(s.config.VectorSearchTimeout) * time.Second,
	}

	// 发送请求
	response, err := client.Do(request)
	if err != nil {
		s.log.Error("Failed to call vector search service", zap.Error(err))
		// 如果向量搜索服务不可用，返回降级结果
		return s.getFallbackResults(ctx, req), 0, nil
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		s.log.Error("Vector search service returned non-OK status", zap.Int("status_code", response.StatusCode))
		return s.getFallbackResults(ctx, req), 0, nil
	}

	// 解析响应
	var searchResponse struct {
		Results []models.SearchResult `json:"results"`
		Total   int                   `json:"total"`
	}

	if err := json.NewDecoder(response.Body).Decode(&searchResponse); err != nil {
		s.log.Error("Failed to decode vector search response", zap.Error(err))
		return s.getFallbackResults(ctx, req), 0, nil
	}

	return searchResponse.Results, searchResponse.Total, nil
}

// executeKeywordSearch 执行关键词搜索
func (s *QueryService) executeKeywordSearch(ctx context.Context, req models.SearchRequest, userID uint) ([]models.SearchResult, int, error) {
	// 关键词搜索实现
	// 这里简化处理，实际项目中应调用全文搜索引擎
	// 现在调用向量搜索服务的关键词搜索接口

	// 准备请求体
	searchBody := map[string]interface{}{
		"query":   req.Query,
		"limit":   req.Limit,
		"offset":  req.Offset,
		"filters": req.Filters,
		"user_id": userID,
	}

	bodyJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal keyword search request: %w", err)
	}

	// 构建请求URL
	requestURL := fmt.Sprintf("%s/api/v1/search/keyword", s.config.VectorSearchServiceAddr)
	s.log.Debug("Executing keyword search", zap.String("url", requestURL), zap.String("query", req.Query))

	// 创建HTTP请求
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create keyword search request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(s.config.VectorSearchTimeout) * time.Second,
	}

	// 发送请求
	response, err := client.Do(request)
	if err != nil {
		s.log.Error("Failed to call keyword search service", zap.Error(err))
		return s.getFallbackResults(ctx, req), 0, nil
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		s.log.Error("Keyword search service returned non-OK status", zap.Int("status_code", response.StatusCode))
		return s.getFallbackResults(ctx, req), 0, nil
	}

	// 解析响应
	var searchResponse struct {
		Results []models.SearchResult `json:"results"`
		Total   int                   `json:"total"`
	}

	if err := json.NewDecoder(response.Body).Decode(&searchResponse); err != nil {
		s.log.Error("Failed to decode keyword search response", zap.Error(err))
		return s.getFallbackResults(ctx, req), 0, nil
	}

	return searchResponse.Results, searchResponse.Total, nil
}

// getFallbackResults 获取降级结果
func (s *QueryService) getFallbackResults(ctx context.Context, req models.SearchRequest) []models.SearchResult {
	// 生成降级结果
	results := make([]models.SearchResult, 0, req.Limit)

	for i := 0; i < req.Limit; i++ {
		result := models.SearchResult{
			DocumentID: fmt.Sprintf("fallback_doc_%d", i),
			Score:      1.0 - float64(i)*0.05,
			Title:      fmt.Sprintf("Fallback Document Result %d", i+1),
			Content:    fmt.Sprintf("This is a fallback result for query: '%s'. The search service may be temporarily unavailable.", req.Query),
			PageNumber: 1,
			Source:     "fallback",
		}
		results = append(results, result)
	}

	return results
}

// saveSearchResults 保存搜索结果到数据库
func (s *QueryService) saveSearchResults(ctx context.Context, queryID uint, results []models.SearchResult) error {
	if len(results) == 0 {
		return nil
	}

	// 批量插入结果
	for _, result := range results {
		result.QueryID = queryID
		if err := s.db.Create(&result).Error; err != nil {
			return err
		}
	}

	return nil
}

// generateCacheKey 生成缓存键
func (s *QueryService) generateCacheKey(userID uint, req models.SearchRequest) string {
	parts := []string{
		"search",
		fmt.Sprintf("user:%d", userID),
		req.Type,
		req.Query,
		fmt.Sprintf("limit:%d", req.Limit),
		fmt.Sprintf("offset:%d", req.Offset),
	}

	if req.Filters != nil {
		filtersJSON, _ := json.Marshal(req.Filters)
		parts = append(parts, fmt.Sprintf("filters:%s", string(filtersJSON)))
	}

	if req.SortBy != "" {
		parts = append(parts, fmt.Sprintf("sort:%s:%s", req.SortBy, req.SortOrder))
	}

	return strings.Join(parts, ":")
}

// GetSearchHistory 获取用户搜索历史
func (s *QueryService) GetSearchHistory(ctx context.Context, userID uint, limit int) ([]string, error) {
	if limit <= 0 || limit > s.config.SearchHistoryMaxItems {
		limit = s.config.SearchHistoryMaxItems
	}

	return utils.GetSearchHistory(ctx, userID, limit)
}

// GetQueryStats 获取查询统计信息
func (s *QueryService) GetQueryStats(ctx context.Context, userID uint) (*models.QueryStats, error) {
	var stats models.QueryStats
	result := s.db.Where("user_id = ?", userID).First(&stats)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新的统计记录
			stats = models.QueryStats{
				UserID: userID,
			}
			if err := s.db.Create(&stats).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	}

	return &stats, nil
}

// GetAnalytics 获取搜索分析数据
func (s *QueryService) GetAnalytics(ctx context.Context, req models.QueryAnalytics) (*models.AnalyticsResponse, error) {
	// 模拟分析数据
	return &models.AnalyticsResponse{
		TotalQueries:        1234,
		UniqueUsers:         42,
		AverageResponseTime: 156.7,
		QueryBreakdown: map[string]int{
			"keyword":  800,
			"semantic": 300,
			"hybrid":   134,
		},
		TimeSeriesData: []models.TimeSeriesPoint{
			{Time: time.Now().AddDate(0, 0, -7), Count: 120},
			{Time: time.Now().AddDate(0, 0, -6), Count: 150},
			{Time: time.Now().AddDate(0, 0, -5), Count: 130},
			{Time: time.Now().AddDate(0, 0, -4), Count: 180},
			{Time: time.Now().AddDate(0, 0, -3), Count: 210},
			{Time: time.Now().AddDate(0, 0, -2), Count: 190},
			{Time: time.Now().AddDate(0, 0, -1), Count: 254},
		},
		TopQueries: []models.TopQuery{
			{QueryText: "project documentation", Count: 45},
			{QueryText: "API reference", Count: 38},
			{QueryText: "setup guide", Count: 32},
			{QueryText: "troubleshooting", Count: 27},
			{QueryText: "best practices", Count: 22},
		},
	}, nil
}

// LogQuery 记录查询日志
func (s *QueryService) LogQuery(ctx context.Context, userID uint, queryText, ip, userAgent string, responseTime int64, status string) error {
	log := models.QueryLog{
		UserID:       userID,
		QueryText:    queryText,
		IP:           ip,
		UserAgent:    userAgent,
		ResponseTime: responseTime,
		Status:       status,
	}

	return s.db.Create(&log).Error
}
