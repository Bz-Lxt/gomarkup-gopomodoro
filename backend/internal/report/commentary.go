package report

import (
	"fmt"
	"strings"
)

func Commentary(f Forecast, abortRate float64, remaining int) string {
	var parts []string
	switch {
	case remaining <= 0:
		parts = append(parts, "该里程碑剩余点数已归零，燃尽完成。")
	case f.OnTrack && f.SlackDays >= 3:
		parts = append(parts, fmt.Sprintf("按当前日均 %.1f 点，进度宽裕，缓冲 %d 天。", f.Velocity, f.SlackDays))
	case f.OnTrack:
		parts = append(parts, fmt.Sprintf("按当前日均 %.1f 点卡点完成，缓冲仅 %d 天。", f.Velocity, f.SlackDays))
	case f.Velocity <= 0:
		parts = append(parts, "近一周没有成功结算的番茄钟，无法预测完成日。")
	default:
		parts = append(parts, fmt.Sprintf("按当前产能将延期，悲观完成日 %s。", f.PessimisticDate))
	}
	if abortRate >= 0.3 {
		parts = append(parts, "专注废弃率偏高，优先减少中途放弃。")
	} else if abortRate > 0 {
		parts = append(parts, fmt.Sprintf("专注废弃率 %.0f%%，Aborted 未计入燃尽。", abortRate*100))
	}
	if f.LikelyDate != "" && remaining > 0 {
		parts = append(parts, "中位预测完成日 "+f.LikelyDate+"。")
	}
	return strings.Join(parts, " ")
}
