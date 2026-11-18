package university

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/university"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantApi struct{}

var tenantService = new(university.TenantService)

type createTenantRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
	Plan string `json:"plan"`
}

type listTenantRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
}

type updateTenantRequest struct {
	ID   uint64 `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Plan string `json:"plan" binding:"required"`
}

type statusTenantRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Status int8   `json:"status" binding:"required"`
}

func (a *TenantApi) Create(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}

	tenant, err := tenantService.Create(req.Code, req.Name, req.Plan)
	if err != nil {
		response.FailWithMessage("创建高校失败，请检查编码是否重复或参数是否正确", c)
		return
	}
	response.OkWithDetailed(gin.H{"tenant": tenant}, "创建成功", c)
}

func (a *TenantApi) List(c *gin.Context) {
	var req listTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	tenants, total, err := tenantService.List(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}

	response.OkWithData(gin.H{
		"list":     tenants,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}, c)
}

func (a *TenantApi) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("高校ID不合法", c)
		return
	}

	tenant, err := tenantService.Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.FailWithMessage("高校不存在", c)
		return
	}
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}

	response.OkWithData(gin.H{"tenant": tenant}, c)
}

func (a *TenantApi) Update(c *gin.Context) {
	var req updateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}

	if err := tenantService.Update(req.ID, req.Name, req.Plan); err != nil {
		response.FailWithMessage("更新失败或高校不存在", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *TenantApi) SetStatus(c *gin.Context) {
	var req statusTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}

	if err := tenantService.SetStatus(req.ID, req.Status); err != nil {
		response.FailWithMessage("状态更新失败", c)
		return
	}
	response.OkWithMessage("状态更新成功", c)
}

// 防止意外将高校管理接口注册为公开接口。
var _ = http.MethodPost
