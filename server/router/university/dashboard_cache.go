package university

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const dashboardCacheTTL = 30 * time.Second

func dashboardCacheKey(tenant, role uint64, days int, filters map[string]uint64) string {
	return fmt.Sprintf("gva:university:dashboard:v1:%d:%d:%d:%d:%d:%d", tenant, role, days,
		filters["collegeId"], filters["majorId"], filters["classId"])
}

func readDashboardCache(ctx context.Context, key string) json.RawMessage {
	if global.GVA_REDIS == nil {
		return nil
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
	defer cancel()
	data, err := global.GVA_REDIS.Get(cacheCtx, key).Bytes()
	if err != nil || len(data) == 0 || data[0] != '{' || !json.Valid(data) {
		return nil
	}
	return json.RawMessage(data)
}

func writeDashboardCache(ctx context.Context, key string, dashboard any) {
	if global.GVA_REDIS == nil {
		return
	}
	data, err := json.Marshal(dashboard)
	if err != nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
	defer cancel()
	_ = global.GVA_REDIS.Set(cacheCtx, key, data, dashboardCacheTTL).Err()
}
