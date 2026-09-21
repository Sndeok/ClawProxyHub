// credit.go — 账号积分快照记录。
//
// 上游只给「剩余积分」快照，没有「本次请求花了多少积分」。所以这里按账号×天记录
// 当天首次与最近一次看到的剩余积分，差值即当天净消耗（充值会把基线一起抬高，
// 不会算成负消耗）。单次请求级口径（插件上报的 credit_used）在网关侧单独统计。
package account

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ShadowSmallBaby/ClawProxyHub/internal/model"
)

// CreditRemaining 从 credits_json 解析剩余积分，兼容字符串与数字两种写法。
func CreditRemaining(creditsJSON string) (float64, bool) {
	if strings.TrimSpace(creditsJSON) == "" {
		return 0, false
	}
	var raw struct {
		Remaining json.RawMessage `json:"remaining"`
	}
	if json.Unmarshal([]byte(creditsJSON), &raw) != nil || len(raw.Remaining) == 0 {
		return 0, false
	}
	s := strings.Trim(string(raw.Remaining), `"`)
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// RecordCreditSnapshot 记录一次积分观测（账号刷新时调用）。
// 当天首采写入基线；剩余高于基线（充值/补发）时把基线上移。
func RecordCreditSnapshot(db *gorm.DB, accountID int64, creditsJSON string) {
	remaining, ok := CreditRemaining(creditsJSON)
	if !ok || accountID <= 0 {
		return
	}
	day := time.Now().Format("2006-01-02")
	now := time.Now()

	var cur model.AccountCreditDaily
	if err := db.Where("account_id = ? AND day = ?", accountID, day).First(&cur).Error; err != nil {
		row := model.AccountCreditDaily{
			AccountID: accountID, Day: day, Remaining: remaining,
			Baseline: remaining, Used: 0, Samples: 1, UpdatedAt: now,
		}
		// 并发刷新同一账号时忽略冲突：赢家已写入当天基线
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
		return
	}
	baseline := cur.Baseline
	if remaining > baseline {
		baseline = remaining
	}
	used := baseline - remaining
	if used < 0 {
		used = 0
	}
	db.Model(&model.AccountCreditDaily{}).
		Where("account_id = ? AND day = ?", accountID, day).
		Updates(map[string]interface{}{
			"remaining": remaining, "baseline": baseline, "used": used,
			"samples": cur.Samples + 1, "updated_at": now,
		})
}
