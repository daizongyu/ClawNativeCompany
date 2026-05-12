// Package service 提供业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"

	"claw/internal/logger"
	"claw/internal/model"
	"claw/internal/repository"
)

// 文档相关错误
var (
	ErrDocumentNotFound    = errors.New("文档不存在")
	ErrVersionNotFound     = errors.New("版本不存在")
	ErrVersionConflict     = errors.New("版本冲突，文档已被他人修改")
	ErrInvalidDocumentData = errors.New("无效的文档数据")
)

// DocumentService 文档服务
type DocumentService struct {
	docRepo     repository.DocumentRepository
	versionRepo repository.DocumentVersionRepository
	channelRepo repository.ChannelRepository
	empRepo     repository.EmployeeRepository
	maxVersions int // 最大保留版本数
}

// NewDocumentService 创建文档服务
func NewDocumentService() *DocumentService {
	return &DocumentService{
		docRepo:     repository.NewDocumentRepository(),
		versionRepo: repository.NewDocumentVersionRepository(),
		channelRepo: repository.NewChannelRepository(),
		empRepo:     repository.NewEmployeeRepository(),
		maxVersions: 20, // 默认保留20个版本
	}
}

// CreateDocumentRequest 创建文档请求
type CreateDocumentRequest struct {
	Title    string `json:"title" validate:"required,min=1,max=200"`
	Content  string `json:"content"`  // 创建时可为空，编辑时填写
	AuthorID string `json:"author_id"`
}

// UpdateDocumentRequest 更新文档请求
type UpdateDocumentRequest struct {
	Title   string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Content string `json:"content,omitempty"`
}

// SaveContentRequest 保存内容请求
type SaveContentRequest struct {
	Content          string `json:"content" validate:"required"`
	ExpectedVersion  int    `json:"expected_version" validate:"required,min=1"`
	EditorID         string `json:"editor_id"`
}

// DocumentDTO 文档响应 DTO
type DocumentDTO struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	Title        string `json:"title"`
	Content      string `json:"content,omitempty"`
	Summary      string `json:"summary"`
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	EditorID     string `json:"editor_id,omitempty"`
	EditorName   string `json:"editor_name,omitempty"`
	Version      int    `json:"version"`
	FileSize     int64  `json:"file_size"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// DocumentListDTO 文档列表响应
type DocumentListDTO struct {
	List     []*DocumentListItem `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// DocumentListItem 文档列表项
type DocumentListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	AuthorID    string `json:"author_id"`
	AuthorName  string `json:"author_name"`
	EditorID    string `json:"editor_id,omitempty"`
	EditorName  string `json:"editor_name,omitempty"`
	Version     int    `json:"version"`
	FileSize    int64  `json:"file_size"`
	UpdatedAt   string `json:"updated_at"`
}

// VersionDTO 版本响应 DTO
type VersionDTO struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	Summary    string `json:"summary"`
	EditorID   string `json:"editor_id"`
	EditorName string `json:"editor_name"`
	CreatedAt  string `json:"created_at"`
}

// VersionListDTO 版本列表响应
type VersionListDTO struct {
	Versions []*VersionDTO `json:"versions"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// Create 创建文档
func (s *DocumentService) Create(ctx context.Context, channelID string, req *CreateDocumentRequest) (*DocumentDTO, error) {
	// 1. 检查频道是否存在
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	// 2. 创建文档
	doc := &model.Document{
		ChannelID: channelID,
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  req.AuthorID,
		Version:   1,
		Status:    model.DocumentStatusActive,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		logger.Error("创建文档失败", "error", err, "channel_id", channelID)
		return nil, fmt.Errorf("创建文档失败: %w", err)
	}

	// 3. 创建第一个版本记录
	version := &model.DocumentVersion{
		DocumentID: doc.ID,
		Version:    1,
		Content:    req.Content,
		EditorID:   req.AuthorID,
		EditReason: "初始创建",
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		logger.Error("创建初始版本失败", "error", err, "doc_id", doc.ID)
		// 继续执行，不阻塞文档创建
	}

	// 4. 更新频道文档数量
	if err := s.channelRepo.UpdateDocCount(ctx, channelID, 1); err != nil {
		logger.Warn("更新频道文档数量失败", "error", err, "channel_id", channelID)
	}

	return s.toDTO(doc, channel), nil
}

// GetByID 根据ID获取文档
func (s *DocumentService) GetByID(ctx context.Context, id string) (*DocumentDTO, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	channel, _ := s.channelRepo.GetByID(ctx, doc.ChannelID)
	return s.toDTO(doc, channel), nil
}

// Update 更新文档（标题等元数据）
func (s *DocumentService) Update(ctx context.Context, id string, req *UpdateDocumentRequest) (*DocumentDTO, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	// 更新字段
	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Content != "" {
		doc.Content = req.Content
	}

	if err := s.docRepo.Update(ctx, doc); err != nil {
		logger.Error("更新文档失败", "error", err, "doc_id", id)
		return nil, fmt.Errorf("更新文档失败: %w", err)
	}

	channel, _ := s.channelRepo.GetByID(ctx, doc.ChannelID)
	return s.toDTO(doc, channel), nil
}

// SaveContent 保存文档内容（带版本历史）
func (s *DocumentService) SaveContent(ctx context.Context, docID string, req *SaveContentRequest) (*DocumentDTO, error) {
	// 1. 获取文档
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	// 2. 乐观锁检查
	if doc.Version != req.ExpectedVersion {
		return nil, &VersionConflictError{
			CurrentVersion:  doc.Version,
			ExpectedVersion: req.ExpectedVersion,
		}
	}

	// 3. 创建历史版本
	newVersionNum := doc.Version + 1
	version := &model.DocumentVersion{
		DocumentID: docID,
		Version:    newVersionNum,
		Content:    req.Content,
		EditorID:   req.EditorID,
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		logger.Error("创建版本记录失败", "error", err, "doc_id", docID)
		return nil, fmt.Errorf("创建版本记录失败: %w", err)
	}

	// 4. 更新文档当前版本
	doc.Content = req.Content
	doc.Version = newVersionNum
	doc.EditorID = &req.EditorID

	if err := s.docRepo.Update(ctx, doc); err != nil {
		logger.Error("更新文档失败", "error", err, "doc_id", docID)
		return nil, fmt.Errorf("更新文档失败: %w", err)
	}

	// 5. 清理过期版本
	if err := s.cleanOldVersions(ctx, docID); err != nil {
		logger.Warn("清理旧版本失败", "error", err, "doc_id", docID)
	}

	channel, _ := s.channelRepo.GetByID(ctx, doc.ChannelID)
	return s.toDTO(doc, channel), nil
}

// cleanOldVersions 清理过期版本
func (s *DocumentService) cleanOldVersions(ctx context.Context, docID string) error {
	count, err := s.versionRepo.CountByDocument(ctx, docID)
	if err != nil {
		return err
	}

	// 保留最新的 maxVersions 个版本
	if count > int64(s.maxVersions) {
		oldVersions, err := s.versionRepo.GetOldestVersions(ctx, docID, int(count)-s.maxVersions)
		if err != nil {
			return err
		}
		for _, v := range oldVersions {
			if err := s.versionRepo.Delete(ctx, v.ID); err != nil {
				logger.Warn("删除旧版本失败", "error", err, "version_id", v.ID)
			}
		}
	}

	return nil
}

// Delete 删除文档
func (s *DocumentService) Delete(ctx context.Context, id string) error {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}

	// 删除文档（级联删除版本记录）
	if err := s.docRepo.Delete(ctx, id); err != nil {
		logger.Error("删除文档失败", "error", err, "doc_id", id)
		return fmt.Errorf("删除文档失败: %w", err)
	}

	// 删除所有版本记录
	if err := s.versionRepo.DeleteByDocument(ctx, id); err != nil {
		logger.Warn("删除文档版本记录失败", "error", err, "doc_id", id)
	}

	// 更新频道文档数量
	if err := s.channelRepo.UpdateDocCount(ctx, doc.ChannelID, -1); err != nil {
		logger.Warn("更新频道文档数量失败", "error", err, "channel_id", doc.ChannelID)
	}

	return nil
}

// ListByChannel 获取频道文档列表
func (s *DocumentService) ListByChannel(ctx context.Context, channelID string, keyword string, sort string, order string, page, pageSize int) (*DocumentListDTO, error) {
	filter := &repository.DocumentFilter{
		Keyword: keyword,
		Sort:    sort,
		Order:   order,
	}

	docs, total, err := s.docRepo.ListByChannel(ctx, channelID, filter, page, pageSize)
	if err != nil {
		logger.Error("获取文档列表失败", "error", err, "channel_id", channelID)
		return nil, fmt.Errorf("获取文档列表失败: %w", err)
	}

	items := make([]*DocumentListItem, 0, len(docs))
	for _, doc := range docs {
		item := &DocumentListItem{
			ID:        doc.ID,
			Title:     doc.Title,
			Summary:   doc.Summary,
			AuthorID:  doc.AuthorID,
			Version:   doc.Version,
			FileSize:  doc.FileSize,
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 获取作者名称
		if author, err := s.empRepo.GetByID(ctx, doc.AuthorID); err == nil {
			item.AuthorName = author.DisplayName
		}

		// 获取编辑者名称
		if doc.EditorID != nil {
			item.EditorID = *doc.EditorID
			if editor, err := s.empRepo.GetByID(ctx, *doc.EditorID); err == nil {
				item.EditorName = editor.DisplayName
			}
		}

		items = append(items, item)
	}

	return &DocumentListDTO{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetVersionList 获取版本列表
func (s *DocumentService) GetVersionList(ctx context.Context, docID string, page, pageSize int) (*VersionListDTO, error) {
	// 检查文档是否存在
	if _, err := s.docRepo.GetByID(ctx, docID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	versions, total, err := s.versionRepo.ListByDocument(ctx, docID, page, pageSize)
	if err != nil {
		logger.Error("获取版本列表失败", "error", err, "doc_id", docID)
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}

	items := make([]*VersionDTO, 0, len(versions))
	for _, v := range versions {
		item := &VersionDTO{
			ID:        v.ID,
			Version:   v.Version,
			Summary:   v.Summary,
			EditorID:  v.EditorID,
			CreatedAt: v.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 获取编辑者名称
		if editor, err := s.empRepo.GetByID(ctx, v.EditorID); err == nil {
			item.EditorName = editor.DisplayName
		}

		items = append(items, item)
	}

	return &VersionListDTO{
		Versions: items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetVersion 获取指定版本
func (s *DocumentService) GetVersion(ctx context.Context, docID string, version int) (*VersionDTO, error) {
	ver, err := s.versionRepo.GetByVersion(ctx, docID, version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	dto := &VersionDTO{
		ID:        ver.ID,
		Version:   ver.Version,
		Summary:   ver.Summary,
		EditorID:  ver.EditorID,
		CreatedAt: ver.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if editor, err := s.empRepo.GetByID(ctx, ver.EditorID); err == nil {
		dto.EditorName = editor.DisplayName
	}

	return dto, nil
}

// RestoreVersion 恢复到指定版本
func (s *DocumentService) RestoreVersion(ctx context.Context, docID string, version int, editorID string) (*DocumentDTO, error) {
	// 1. 获取指定版本内容
	oldVer, err := s.versionRepo.GetByVersion(ctx, docID, version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	// 2. 获取当前文档
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	// 3. 创建新版本（恢复的内容）
	newVersionNum := doc.Version + 1
	newVersion := &model.DocumentVersion{
		DocumentID: docID,
		Version:    newVersionNum,
		Content:    oldVer.Content,
		EditorID:   editorID,
		EditReason: fmt.Sprintf("从版本 v%d 恢复", version),
	}
	if err := s.versionRepo.Create(ctx, newVersion); err != nil {
		logger.Error("创建恢复版本失败", "error", err, "doc_id", docID)
		return nil, fmt.Errorf("创建恢复版本失败: %w", err)
	}

	// 4. 更新文档
	doc.Content = oldVer.Content
	doc.Version = newVersionNum
	doc.EditorID = &editorID

	if err := s.docRepo.Update(ctx, doc); err != nil {
		logger.Error("更新文档失败", "error", err, "doc_id", docID)
		return nil, fmt.Errorf("更新文档失败: %w", err)
	}

	channel, _ := s.channelRepo.GetByID(ctx, doc.ChannelID)
	return s.toDTO(doc, channel), nil
}

// toDTO 转换为 DTO
func (s *DocumentService) toDTO(doc *model.Document, channel *model.Channel) *DocumentDTO {
	dto := &DocumentDTO{
		ID:        doc.ID,
		ChannelID: doc.ChannelID,
		Title:     doc.Title,
		Content:   doc.Content,
		Summary:   doc.Summary,
		AuthorID:  doc.AuthorID,
		Version:   doc.Version,
		FileSize:  doc.FileSize,
		Status:    string(doc.Status),
		CreatedAt: doc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 获取作者名称
	if doc.Author != nil {
		dto.AuthorName = doc.Author.DisplayName
	} else if author, err := s.empRepo.GetByID(context.Background(), doc.AuthorID); err == nil {
		dto.AuthorName = author.DisplayName
	}

	// 获取编辑者名称
	if doc.EditorID != nil {
		dto.EditorID = *doc.EditorID
		if doc.Editor != nil {
			dto.EditorName = doc.Editor.DisplayName
		} else if editor, err := s.empRepo.GetByID(context.Background(), *doc.EditorID); err == nil {
			dto.EditorName = editor.DisplayName
		}
	}

	return dto
}

// VersionConflictError 版本冲突错误
type VersionConflictError struct {
	CurrentVersion  int
	ExpectedVersion int
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("版本冲突：当前版本 %d，期望版本 %d", e.CurrentVersion, e.ExpectedVersion)
}

// IsVersionConflict 检查是否为版本冲突错误
func IsVersionConflict(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*VersionConflictError)
	return ok
}
