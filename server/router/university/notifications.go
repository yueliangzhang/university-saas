package university

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	notificationStudent = "student"
	notificationAdvisor = "advisor"
)

type uniLeaveNotification struct {
	ID            uint64     `json:"id" gorm:"primaryKey"`
	TenantID      uint64     `json:"-" gorm:"uniqueIndex:uk_uni_notification_event"`
	LeaveID       uint64     `json:"leaveId" gorm:"uniqueIndex:uk_uni_notification_event"`
	Event         string     `json:"event" gorm:"uniqueIndex:uk_uni_notification_event"`
	RecipientKind string     `json:"-"`
	RecipientID   uint64     `json:"-"`
	ReadAt        *time.Time `json:"readAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (uniLeaveNotification) TableName() string { return "uni_leave_notifications" }

// recordLeaveNotifications runs in the same transaction as the leave change.
func recordLeaveNotifications(tx *gorm.DB, tenantID uint64, leaveIDs []uint64, event string) error {
	kind := notificationStudent
	q := tx.Table("uni_leave_requests l").Where("l.tenant_id = ? AND l.id IN ?", tenantID, leaveIDs)
	recipient := "l.student_id"
	if event == "submitted" || event == "cancelled" {
		kind = notificationAdvisor
		q = q.Joins("JOIN uni_students s ON s.tenant_id=l.tenant_id AND s.id=l.student_id").
			Where("s.advisor_id IS NOT NULL AND s.advisor_id <> 0")
		recipient = "s.advisor_id"
	}
	var notices []uniLeaveNotification
	if err := q.Select("l.tenant_id, l.id AS leave_id, " + recipient + " AS recipient_id").Scan(&notices).Error; err != nil {
		return err
	}
	if len(notices) == 0 {
		return nil
	}
	for i := range notices {
		notices[i].Event = event
		notices[i].RecipientKind = kind
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&notices).Error
}

func scopedNotifications(c *gin.Context, kind string) *gorm.DB {
	tenantID := c.GetUint64(studentPortalTenantID)
	recipientID := c.GetUint64(studentPortalStudentID)
	if kind == notificationAdvisor {
		tenantID = c.GetUint64(teacherTenantKey)
		recipientID = c.GetUint64(teacherIDKey)
	}
	db := global.GVA_DB.WithContext(c.Request.Context())
	leave := db.Table("uni_leave_requests l").Select("1").
		Where("l.tenant_id=uni_leave_notifications.tenant_id AND l.id=uni_leave_notifications.leave_id")
	if kind == notificationAdvisor {
		leave = leave.Joins("JOIN uni_students s ON s.tenant_id=l.tenant_id AND s.id=l.student_id").
			Where("s.advisor_id=?", recipientID)
	} else {
		leave = leave.Where("l.student_id=?", recipientID)
	}
	return db.Model(&uniLeaveNotification{}).
		Where("tenant_id=? AND recipient_kind=? AND recipient_id=?", tenantID, kind, recipientID).
		Where("EXISTS (?)", leave)
}

func registerLeaveNotifications(group gin.IRoutes, kind string) {
	group.GET("/notifications", func(c *gin.Context) {
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 || page > 1000000 {
			leaveFail(c, http.StatusBadRequest, "页码不合法")
			return
		}
		const pageSize = 15
		var total, unread int64
		q := scopedNotifications(c, kind)
		if err := q.Session(&gorm.Session{}).Where("read_at IS NULL").Count(&unread).Error; err != nil {
			leaveFail(c, http.StatusInternalServerError, "查询未读通知失败")
			return
		}
		if c.Query("unreadOnly") == "true" {
			leaveOK(c, gin.H{"unread": unread})
			return
		}
		if err := q.Count(&total).Error; err != nil {
			leaveFail(c, http.StatusInternalServerError, "查询通知失败")
			return
		}
		notices := make([]uniLeaveNotification, 0)
		if err := q.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&notices).Error; err != nil {
			leaveFail(c, http.StatusInternalServerError, "查询通知失败")
			return
		}
		leaveOK(c, gin.H{"list": notices, "total": total, "unread": unread, "page": page, "pageSize": pageSize})
	})
	group.POST("/notifications/read", func(c *gin.Context) {
		var input struct {
			ID  uint64 `json:"id"`
			All bool   `json:"all"`
		}
		if c.ShouldBindJSON(&input) != nil || input.All == (input.ID != 0) {
			leaveFail(c, http.StatusBadRequest, "请选择单条通知或全部已读")
			return
		}
		q := scopedNotifications(c, kind).Where("read_at IS NULL")
		if !input.All {
			q = q.Where("id=?", input.ID)
		}
		if err := q.Update("read_at", time.Now()).Error; err != nil {
			leaveFail(c, http.StatusInternalServerError, "标记已读失败")
			return
		}
		leaveOK(c, gin.H{})
	})
}

func filterLeaveID(c *gin.Context, q *gorm.DB) (*gorm.DB, bool) {
	raw := c.Query("id")
	if raw == "" {
		return q, true
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		leaveFail(c, http.StatusBadRequest, "请假单ID不合法")
		return nil, false
	}
	return q.Where("l.id=?", id), true
}
