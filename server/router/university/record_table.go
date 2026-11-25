package university

import (
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func deletionIDs(c *gin.Context, allowBatch bool) ([]uint64, bool) {
	raw := c.Query("id")
	batch, present := c.GetQuery("ids")
	if present {
		if !allowBatch || raw != "" {
			return nil, false
		}
		raw = batch
	}
	parts := strings.Split(raw, ",")
	if len(parts) > 100 || (len(parts) > 1 && !present) {
		return nil, false
	}
	ids := make([]uint64, 0, len(parts))
	seen := make(map[uint64]bool, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseUint(part, 10, 64)
		if err != nil || id == 0 || seen[id] {
			return nil, false
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, true
}

func deleteRecords(remove func(*gorm.DB) *gorm.DB, count int) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		result := remove(tx)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(count) {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func recordOrder(c *gin.Context, query *gorm.DB, idColumn string, columns map[string]string) *gorm.DB {
	column := columns[c.Query("sortBy")]
	direction := c.Query("sortOrder")
	if column != "" && (direction == "asc" || direction == "desc") {
		query = query.Order(column + " " + direction)
	}
	return query.Order(idColumn + " DESC")
}
