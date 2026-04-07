package main

import (
	"fmt"
	"math"
)

// ==================== 常量定义 ====================

// 阶梯阈值（度）
const (
	Threshold1 = 200.0  // 第一阶梯上限
	Threshold2 = 400.0  // 第二阶梯上限
	// 400度以上为第三阶梯
)

// 基础电价（元/度）
const (
	Price1 = 0.50 // 第一阶梯电价
	Price2 = 0.60 // 第二阶梯电价
	Price3 = 0.80 // 第三阶梯电价
)

// 峰谷时段电价倍率
const (
	PeakMultiplier   = 1.5  // 峰时倍率
	NormalMultiplier = 1.0  // 平时倍率
	ValleyMultiplier = 0.5  // 谷时倍率
)

// 峰谷时段定义（小时，24小时制）
const (
	PeakStartHour   = 8
	PeakEndHour     = 12
	PeakStartHour2  = 18
	PeakEndHour2    = 22
	ValleyStartHour = 22
	ValleyEndHour   = 6
)

// 违约金参数
const (
	GracePeriod       = 10  // 宽限期（天）
	PenaltyRate       = 0.001 // 违约金日费率（千分之一）
	MaxPenaltyRate    = 0.10  // 最大违约金比例（10%封顶）
)

// ==================== 时段判断函数 ====================

// TimePeriod 表示时段类型
type TimePeriod int

const (
	PeakPeriod   TimePeriod = iota // 峰时
	NormalPeriod                    // 平时
	ValleyPeriod                    // 谷时
)

// GetTimePeriod 根据时间判断当前属于哪个时段
// 参数: hour - 小时数（0-23）
// 返回: TimePeriod - 时段类型
func GetTimePeriod(hour int) TimePeriod {
	// 峰时: 8:00-12:00, 18:00-22:00
	if (hour >= PeakStartHour && hour < PeakEndHour) || 
	   (hour >= PeakStartHour2 && hour < PeakEndHour2) {
		return PeakPeriod
	}
	
	// 谷时: 22:00-次日6:00
	if hour >= ValleyStartHour || hour < ValleyEndHour {
		return ValleyPeriod
	}
	
	// 平时: 其他时段
	return NormalPeriod
}

// GetTimeMultiplier 获取时段电价倍率
func GetTimeMultiplier(period TimePeriod) float64 {
	switch period {
	case PeakPeriod:
		return PeakMultiplier
	case ValleyPeriod:
		return ValleyMultiplier
	default:
		return NormalMultiplier
	}
}

// GetPeriodName 获取时段名称
func GetPeriodName(period TimePeriod) string {
	switch period {
	case PeakPeriod:
		return "峰时"
	case ValleyPeriod:
		return "谷时"
	default:
		return "平时"
	}
}

// ==================== 阶梯电费计算函数 ====================

// CalculateTieredCharge 计算阶梯电费（不含时段倍率）
// 参数: usage - 用电量（度）
// 返回: 基础电费（元）
func CalculateTieredCharge(usage float64) float64 {
	var charge float64 = 0.0
	
	// 用电量 <= 0，返回0
	if usage <= 0 {
		return 0.0
	}
	
	// 第一阶梯：0-200度
	if usage <= Threshold1 {
		charge = usage * Price1
		return charge
	}
	
	// 第一阶梯电费（满额）
	charge = Threshold1 * Price1
	
	// 第二阶梯：200-400度
	if usage <= Threshold2 {
		charge += (usage - Threshold1) * Price2
		return charge
	}
	
	// 第二阶梯电费（满额）
	charge += (Threshold2 - Threshold1) * Price2
	
	// 第三阶梯：400度以上
	charge += (usage - Threshold2) * Price3
	
	return charge
}

// CalculateBillWithTime 计算含时段倍率的电费
// 参数: usage - 用电量（度）, hour - 小时数（0-23）
// 返回: 电费（元）
func CalculateBillWithTime(usage float64, hour int) float64 {
	// 先计算基础阶梯电费
	baseCharge := CalculateTieredCharge(usage)
	
	// 获取时段倍率
	period := GetTimePeriod(hour)
	multiplier := GetTimeMultiplier(period)
	
	// 应用时段倍率
	return baseCharge * multiplier
}

// ==================== 违约金计算函数 ====================

// CalculatePenalty 计算违约金
// 参数: billAmount - 账单金额, overdueDays - 欠费天数
// 返回: 违约金金额
func CalculatePenalty(billAmount float64, overdueDays int) float64 {
	// 宽限期内不收违约金
	if overdueDays <= GracePeriod {
		return 0.0
	}
	
	// 计算违约天数
	penaltyDays := overdueDays - GracePeriod
	
	// 按日费率计算违约金
	penalty := billAmount * PenaltyRate * float64(penaltyDays)
	
	// 违约金上限为账单金额的10%
	maxPenalty := billAmount * MaxPenaltyRate
	
	if penalty > maxPenalty {
		penalty = maxPenalty
	}
	
	// 保留两位小数
	return math.Round(penalty*100) / 100
}

// ==================== 主计费函数 ====================

// BillResult 账单结果结构体
type BillResult struct {
	Usage        float64    // 用电量（度）
	Hour         int        // 用电时段（小时）
	Period       TimePeriod // 时段类型
	BaseCharge   float64    // 基础阶梯电费
	FinalCharge  float64    // 最终电费（含时段倍率）
}

// CalculateBill 主计费函数
// 参数: usage - 用电量（度）, hour - 小时数（0-23）
// 返回: BillResult - 账单结果
func CalculateBill(usage float64, hour int) BillResult {
	// 确保用电量非负
	if usage < 0 {
		usage = 0
	}
	
	// 确保小时在有效范围内
	if hour < 0 {
		hour = 0
	} else if hour > 23 {
		hour = 23
	}
	
	// 获取时段信息
	period := GetTimePeriod(hour)
	
	// 计算基础阶梯电费
	baseCharge := CalculateTieredCharge(usage)
	
	// 计算最终电费（含时段倍率）
	multiplier := GetTimeMultiplier(period)
	finalCharge := baseCharge * multiplier
	
	// 保留两位小数
	finalCharge = math.Round(finalCharge*100) / 100
	
	return BillResult{
		Usage:       usage,
		Hour:        hour,
		Period:      period,
		BaseCharge:  baseCharge,
		FinalCharge: finalCharge,
	}
}

// ==================== 用户交互函数 ====================

// PrintBill 打印账单明细
func PrintBill(result BillResult) {
	fmt.Println("-------账单明细-------")
	fmt.Printf("当前用电：%.2f度\n", result.Usage)
	fmt.Printf("当前时段：%d:00点（%s）\n", result.Hour, GetPeriodName(result.Period))
	fmt.Printf("基础电费：%.2f元\n", result.BaseCharge)
	fmt.Printf("时段倍率：%.1f倍\n", GetTimeMultiplier(result.Period))
	fmt.Printf("最终电费：%.2f元\n", result.FinalCharge)
	fmt.Println("---------------------")
}

// RunBilling 运行计费流程（供main调用）
func RunBilling() {
	var usage float64
	var hour int
	
	// 引导用户输入用电量
	fmt.Print("请输入用电量（度）：")
	fmt.Scanln(&usage)
	
	// 引导用户输入用电时段
	fmt.Print("请输入用电时段（小时，0-23）：")
	fmt.Scanln(&hour)
	
	// 计算电费
	result := CalculateBill(usage, hour)
	
	// 打印账单
	PrintBill(result)
}

// ==================== 辅助函数 ====================

// CalculateFullBill 计算包含违约金的完整账单
// 参数: usage - 用电量, hour - 小时, overdueDays - 欠费天数
// 返回: 电费, 违约金, 总计
func CalculateFullBill(usage float64, hour int, overdueDays int) (float64, float64, float64) {
	result := CalculateBill(usage, hour)
	penalty := CalculatePenalty(result.FinalCharge, overdueDays)
	total := result.FinalCharge + penalty
	return result.FinalCharge, penalty, total
}

func main() {
	RunBilling()
}
