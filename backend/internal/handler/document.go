// Package handler 提供 HTTP 请求处理
package handler

import (
	"net/http"
	"strconv"

	"claw/internal/middleware"
	"claw/internal/service"
	"claw/pkg/utils"
	"claw/pkg/validator"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档 Handler
type DocumentHandler struct {
	docService *service.DocumentService
}

// NewDocumentHandler 创建文档 Handler
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{
		docService: service.NewDocumentService(),
	}
}

// Create 创建文档
// POST /api/v1/documents?channel_id=xxx
func (h *DocumentHandler) Create(c *gin.Context) {
	channelID := c.Query("channel_id")
	if channelID == "" {
		utils.ValidationError(c, "频道ID不能为空")
		return
	}

	var req service.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	// 获取当前用户ID
	employeeID, _ := c.Get(string(middleware.ContextKeyEmployeeID))
	req.AuthorID = employeeID.(string)

	// 执行创建
	doc, err := h.docService.Create(c.Request.Context(), channelID, &req)
	if err != nil {
		switch err {
		case service.ErrChannelNotFound:
			utils.Error(c, http.StatusNotFound, "频道不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "创建文档失败")
		}
		return
	}

	utils.SuccessWithData(c, doc)
}

// Get 获取文档详情
// GET /api/v1/documents/:id
func (h *DocumentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "文档ID不能为空")
		return
	}

	doc, err := h.docService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取文档失败")
		}
		return
	}

	utils.SuccessWithData(c, doc)
}

// Update 更新文档
// PUT /api/v1/documents/:id
func (h *DocumentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "文档ID不能为空")
		return
	}

	var req service.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	doc, err := h.docService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "更新文档失败")
		}
		return
	}

	utils.SuccessWithData(c, doc)
}

// SaveContent 保存文档内容
// PUT /api/v1/documents/:id/content
func (h *DocumentHandler) SaveContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "文档ID不能为空")
		return
	}

	var req service.SaveContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "请求参数错误: "+err.Error())
		return
	}

	// 参数校验
	if err := validator.ValidateStruct(req); err != nil {
		errors := validator.FormatValidationError(err)
		if len(errors) > 0 {
			utils.ValidationError(c, errors[0].Message)
			return
		}
	}

	// 获取当前用户ID
	employeeID, _ := c.Get(string(middleware.ContextKeyEmployeeID))
	req.EditorID = employeeID.(string)

	doc, err := h.docService.SaveContent(c.Request.Context(), id, &req)
	if err != nil {
		switch {
		case err == service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		case service.IsVersionConflict(err):
			conflict := err.(*service.VersionConflictError)
			utils.Error(c, http.StatusConflict, "文档已被他人修改，请刷新后重新编辑")
			c.JSON(http.StatusConflict, gin.H{
				"code": 409,
				"message": "文档已被他人修改，请刷新后重新编辑",
				"data": gin.H{
					"current_version":  conflict.CurrentVersion,
					"expected_version": conflict.ExpectedVersion,
				},
			})
			return
		default:
			utils.Error(c, http.StatusInternalServerError, "保存文档失败")
		}
		return
	}

	utils.SuccessWithData(c, doc)
}

// Delete 删除文档
// DELETE /api/v1/documents/:id
func (h *DocumentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "文档ID不能为空")
		return
	}

	if err := h.docService.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "删除文档失败")
		}
		return
	}

	utils.Success(c)
}

// ListByChannel 获取频道文档列表
// GET /api/v1/documents?channel_id=xxx
func (h *DocumentHandler) ListByChannel(c *gin.Context) {
	channelID := c.Query("channel_id")
	if channelID == "" {
		utils.ValidationError(c, "频道ID不能为空")
		return
	}

	// 获取查询参数
	keyword := c.Query("keyword")
	sort := c.DefaultQuery("sort", "updated_at")
	order := c.DefaultQuery("order", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.docService.ListByChannel(c.Request.Context(), channelID, keyword, sort, order, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取文档列表失败")
		return
	}

	utils.SuccessWithData(c, result)
}

// GetVersions 获取文档版本列表
// GET /api/v1/documents/:id/versions
func (h *DocumentHandler) GetVersions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ValidationError(c, "文档ID不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.docService.GetVersionList(c.Request.Context(), id, page, pageSize)
	if err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取版本列表失败")
		}
		return
	}

	utils.SuccessWithData(c, result)
}

// GetVersion 获取指定版本
// GET /api/v1/documents/:id/versions/:version
func (h *DocumentHandler) GetVersion(c *gin.Context) {
	docID := c.Param("id")
	versionStr := c.Param("version")

	if docID == "" || versionStr == "" {
		utils.ValidationError(c, "文档ID和版本号不能为空")
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		utils.ValidationError(c, "版本号格式不正确")
		return
	}

	ver, err := h.docService.GetVersion(c.Request.Context(), docID, version)
	if err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		case service.ErrVersionNotFound:
			utils.Error(c, http.StatusNotFound, "版本不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "获取版本失败")
		}
		return
	}

	utils.SuccessWithData(c, ver)
}

// RestoreVersion 恢复到指定版本
// POST /api/v1/documents/:id/versions/:version/restore
func (h *DocumentHandler) RestoreVersion(c *gin.Context) {
	docID := c.Param("id")
	versionStr := c.Param("version")

	if docID == "" || versionStr == "" {
		utils.ValidationError(c, "文档ID和版本号不能为空")
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		utils.ValidationError(c, "版本号格式不正确")
		return
	}

	// 获取当前用户ID
	employeeID, _ := c.Get(string(middleware.ContextKeyEmployeeID))

	doc, err := h.docService.RestoreVersion(c.Request.Context(), docID, version, employeeID.(string))
	if err != nil {
		switch err {
		case service.ErrDocumentNotFound:
			utils.Error(c, http.StatusNotFound, "文档不存在")
		case service.ErrVersionNotFound:
			utils.Error(c, http.StatusNotFound, "版本不存在")
		default:
			utils.Error(c, http.StatusInternalServerError, "恢复版本失败")
		}
		return
	}

	utils.SuccessWithData(c, doc)
}

// RegisterRoutes 注册路由
// 注意：路由顺序很重要，具体路径必须在通配符路径之前注册
func (h *DocumentHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 频道文档列表和创建（使用查询参数，必须在 /:id 之前）
	r.GET("/documents", h.ListByChannel)
	r.POST("/documents", h.Create)

	// 版本相关路由（必须在 /:id 之前）
	r.GET("/documents/:id/versions", h.GetVersions)
	r.GET("/documents/:id/versions/:version", h.GetVersion)
	r.POST("/documents/:id/versions/:version/restore", h.RestoreVersion)
	r.PUT("/documents/:id/content", h.SaveContent)

	// 基础 CRUD（最后注册）
	r.GET("/documents/:id", h.Get)
	r.PUT("/documents/:id", h.Update)
	r.DELETE("/documents/:id", h.Delete)
}
