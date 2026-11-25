package university

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	universityService "github.com/flipped-aurora/gin-vue-admin/server/service/university"
	"github.com/gin-gonic/gin"
)

type AdminApi struct{}

var adminService = new(universityService.AdminService)

type createAdminRequest struct {
	TenantID    uint64 `json:"tenantId" binding:"required"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	NickName    string `json:"nickName"`
	AuthorityID uint   `json:"authorityId"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type listAdminRequest struct {
	TenantID uint64 `json:"tenantId"`
	Keyword  string `json:"keyword"`
	Page     int    `json:"page" binding:"required"`
	PageSize int    `json:"pageSize" binding:"required"`
}

type statusAdminRequest struct {
	TenantID     uint64 `json:"tenantId" binding:"required"`
	MembershipID uint64 `json:"membershipId" binding:"required"`
	Status       int8   `json:"status" binding:"required"`
}

func (AdminApi) Create(c *gin.Context) {
	var r createAdminRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}
	if strings.TrimSpace(r.Username) == "" {
		response.FailWithMessage("用户名不能为空", c)
		return
	}
	id, err := adminService.CreateAccount(r.TenantID, r.Username, r.Password, r.NickName, r.AuthorityID, r.Phone, r.Email)
	if err != nil {
		if errors.Is(err, universityService.ErrUsernameTaken) || errors.Is(err, universityService.ErrTenantNotActive) {
			response.FailWithMessage(err.Error(), c)
		} else {
			response.FailWithMessage("创建管理员失败，请检查参数或服务日志", c)
		}
		return
	}
	response.OkWithData(gin.H{"userId": id}, c)
}

func (AdminApi) List(c *gin.Context) {
	var r listAdminRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}
	rows, total, err := adminService.ListAdmins(r.TenantID, r.Page, r.PageSize, r.Keyword)
	if err != nil {
		response.FailWithMessage("查询管理员失败", c)
		return
	}
	response.OkWithData(gin.H{"list": rows, "total": total, "page": r.Page, "pageSize": r.PageSize}, c)
}

func (AdminApi) SetStatus(c *gin.Context) {
	var r statusAdminRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage("请求参数不合法", c)
		return
	}
	if err := adminService.SetAdminStatus(r.TenantID, r.MembershipID, r.Status); err != nil {
		if errors.Is(err, universityService.ErrMemberNotFound) {
			response.FailWithMessage(err.Error(), c)
		} else {
			response.FailWithMessage("更新管理员状态失败", c)
		}
		return
	}
	response.OkWithMessage("状态更新成功", c)
}

type adminEditRequest struct {
	TenantID         uint64  `json:"tenantId" binding:"required"`
	MembershipID     uint64  `json:"membershipId" binding:"required"`
	NickName         string  `json:"nickName"`
	Phone            string  `json:"phone"`
	Email            string  `json:"email"`
	Password         string  `json:"password"`
	AdvisorTeacherNo *string `json:"advisorTeacherNo"`
}

func (AdminApi) Edit(c *gin.Context) {
	var r adminEditRequest
	if c.ShouldBindJSON(&r) != nil {
		response.FailWithMessage("参数不合法", c)
		return
	}
	if err := adminService.Edit(r.TenantID, r.MembershipID, r.NickName, r.Phone, r.Email, r.AdvisorTeacherNo); err != nil {
		if errors.Is(err, universityService.ErrAdvisorLink) {
			response.FailWithMessage(err.Error(), c)
		} else {
			response.FailWithMessage("修改失败", c)
		}
		return
	}
	response.OkWithMessage("修改成功", c)
}
func (AdminApi) Reset(c *gin.Context) {
	var r adminEditRequest
	if c.ShouldBindJSON(&r) != nil {
		response.FailWithMessage("参数不合法", c)
		return
	}
	if adminService.Reset(r.TenantID, r.MembershipID, r.Password) != nil {
		response.FailWithMessage("重置失败，密码须为12至72字节", c)
		return
	}
	response.OkWithMessage("重置成功", c)
}
