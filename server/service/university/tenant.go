package university

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/university"
	"gorm.io/gorm"
)

type TenantService struct{}

func (s *TenantService) Create(code, name, plan string) (*university.Tenant, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("数据库未初始化")
	}

	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	plan = strings.TrimSpace(plan)

	if code == "" || name == "" {
		return nil, errors.New("高校编码和名称不能为空")
	}
	if len(code) > 64 || len(name) > 200 || len(plan) > 32 {
		return nil, errors.New("字段长度超出限制")
	}
	if plan == "" {
		plan = "basic"
	}

	tenant := university.Tenant{
		Code:   code,
		Name:   name,
		Status: 1,
		Plan:   plan,
	}

	if err := global.GVA_DB.Create(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *TenantService) List(page, pageSize int, keyword string) ([]university.Tenant, int64, error) {
	var tenants []university.Tenant
	var total int64

	if global.GVA_DB == nil {
		return nil, 0, errors.New("数据库未初始化")
	}
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return nil, 0, errors.New("分页参数不合法")
	}

	db := global.GVA_DB.Model(&university.Tenant{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		db = db.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&tenants).Error

	return tenants, total, err
}

func (s *TenantService) Detail(id uint64) (*university.Tenant, error) {
	var tenant university.Tenant
	if global.GVA_DB == nil {
		return nil, errors.New("数据库未初始化")
	}

	err := global.GVA_DB.First(&tenant, id).Error
	return &tenant, err
}

func (s *TenantService) Update(id uint64, name, plan string) error {
	if global.GVA_DB == nil {
		return errors.New("数据库未初始化")
	}

	name = strings.TrimSpace(name)
	plan = strings.TrimSpace(plan)

	if id == 0 || name == "" || plan == "" {
		return errors.New("参数不完整")
	}
	if len(name) > 200 || len(plan) > 32 {
		return errors.New("字段长度超出限制")
	}

	result := global.GVA_DB.Model(&university.Tenant{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name": name,
			"plan": plan,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *TenantService) SetStatus(id uint64, status int8) error {
	if global.GVA_DB == nil {
		return errors.New("数据库未初始化")
	}
	if id == 0 || (status != 1 && status != 2) {
		return errors.New("高校ID或状态不合法")
	}

	result := global.GVA_DB.Model(&university.Tenant{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
