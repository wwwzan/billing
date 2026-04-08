package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ==================== 常量定义 ====================

// 计费规则版本号
const BillingVersion = "v1.0.0"

// 阶梯阈值（度）
const (
	Threshold1 = 200.0 // 第一阶梯上限
	Threshold2 = 400.0 // 第二阶梯上限
	// 400度以上为第三阶梯
)

// 基础电价（元/度）
const (
	Price1 = 0.50 // 第一阶梯电价：0-200度
	Price2 = 0.80 // 第二阶梯电价：200-400度（超过部分）
	Price3 = 1.20 // 第三阶梯电价：400度以上（超过部分）
)

// 峰谷时段调节因子
const (
	PeakRateIncrease  = 0.10  // 高峰时段总费率增加10%
	ValleyRateDecrease = 0.20 // 低谷时段总费率减少20%
)

// 时段定义（小时，24小时制）
const (
	PeakStartHour   = 8  // 高峰开始时间（8:00，包含）
	PeakEndHour     = 22 // 高峰结束时间（22:00，不包含）
	ValleyStartHour = 22 // 低谷开始时间（22:00，包含）
	ValleyEndHour   = 8  // 低谷结束时间（次日8:00，不包含）
)

// 系统初始化时间
var initTime string

// ==================== 初始化函数 ====================

// init 初始化系统
func init() {
	initTime = time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("=================================")
	fmt.Printf("计费规则版本号: %s\n", BillingVersion)
	fmt.Printf("系统初始化时间: %s\n", initTime)
	fmt.Println("=================================")
	fmt.Println()
}

// ==================== 时段判断函数 ====================

// TimePeriod 表示时段类型
type TimePeriod int

const (
	PeakPeriod   TimePeriod = iota // 高峰时段
	ValleyPeriod                    // 低谷时段
)

// GetTimePeriod 根据时间判断当前属于哪个时段
// 参数: hour - 小时数（0-23）
// 返回: TimePeriod - 时段类型
func GetTimePeriod(hour int) TimePeriod {
	// 高峰时段: (8:00-22:00]
	if hour >= PeakStartHour && hour < PeakEndHour {
		return PeakPeriod
	}

	// 低谷时段: (22:00-次日8:00]
	return ValleyPeriod
}

// GetPeriodName 获取时段名称
func GetPeriodName(period TimePeriod) string {
	switch period {
	case PeakPeriod:
		return "高峰时段"
	default:
		return "低谷时段"
	}
}

// GetPeriodRateAdjustment 获取时段费率调节因子
// 返回: 调节后的倍率（1+增加比例 或 1-减少比例）
func GetPeriodRateAdjustment(period TimePeriod) float64 {
	switch period {
	case PeakPeriod:
		return 1 + PeakRateIncrease // 1.10，增加10%
	default:
		return 1 - ValleyRateDecrease // 0.80，减少20%
	}
}

// ==================== 阶梯电费计算函数 ====================

// CalculateTieredCharge 计算阶梯电费（不含时段调节）
// 参数: usage - 用电量（度）
// 返回: 基础电费（元）
func CalculateTieredCharge(usage float64) float64 {
	var charge float64 = 0.0

	// 用电量 <= 0，返回0
	if usage <= 0 {
		return 0.0
	}

	// 第一阶梯：0-200度，单价0.5元/度
	if usage <= Threshold1 {
		charge = usage * Price1
		return charge
	}

	// 第一阶梯电费（满额）
	charge = Threshold1 * Price1

	// 第二阶梯：200-400度，超过部分单价0.8元/度
	if usage <= Threshold2 {
		charge += (usage - Threshold1) * Price2
		return charge
	}

	// 第二阶梯电费（满额）
	charge += (Threshold2 - Threshold1) * Price2

	// 第三阶梯：400度以上，超过部分单价1.2元/度
	charge += (usage - Threshold2) * Price3

	return charge
}

// CalculateBillWithTime 计算含时段调节的电费
// 参数: usage - 用电量（度）, hour - 小时数（0-23）
// 返回: 电费（元）
func CalculateBillWithTime(usage float64, hour int) float64 {
	// 先计算基础阶梯电费
	baseCharge := CalculateTieredCharge(usage)

	// 获取时段调节因子
	period := GetTimePeriod(hour)
	adjustment := GetPeriodRateAdjustment(period)

	// 应用时段调节（总费率增加或减少）
	return baseCharge * adjustment
}

// ==================== 主计费函数 ====================

// BillResult 账单结果结构体
type BillResult struct {
	Usage           float64    // 用电量（度）
	Hour            int        // 用电时段（小时）
	Period          TimePeriod // 时段类型
	BaseCharge      float64    // 基础阶梯电费
	AdjustmentRate  float64    // 时段调节比例
	FinalCharge     float64    // 最终电费（含时段调节）
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

	// 计算最终电费（含时段调节）
	adjustment := GetPeriodRateAdjustment(period)
	finalCharge := baseCharge * adjustment

	// 保留两位小数
	finalCharge = math.Round(finalCharge*100) / 100

	var adjustmentRate float64
	if period == PeakPeriod {
		adjustmentRate = PeakRateIncrease
	} else {
		adjustmentRate = -ValleyRateDecrease
	}

	return BillResult{
		Usage:          usage,
		Hour:           hour,
		Period:         period,
		BaseCharge:     baseCharge,
		AdjustmentRate: adjustmentRate,
		FinalCharge:    finalCharge,
	}
}

// ==================== 用户交互函数 ====================

// ParseTimeInput 解析用户输入的时间字符串
// 支持格式: "14:00", "14", "8:30" 等
// 返回: 小时数（0-23），错误信息
func ParseTimeInput(timeStr string) (int, error) {
	timeStr = strings.TrimSpace(timeStr)

	// 尝试直接解析为整数（小时）
	if hour, err := strconv.Atoi(timeStr); err == nil {
		if hour >= 0 && hour <= 23 {
			return hour, nil
		}
		return 0, fmt.Errorf("小时数必须在0-23之间")
	}

	// 尝试解析 "HH:MM" 格式
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, fmt.Errorf("时间格式错误")
		}
		if hour >= 0 && hour <= 23 {
			return hour, nil
		}
		return 0, fmt.Errorf("小时数必须在0-23之间")
	}

	return 0, fmt.Errorf("时间格式错误，请使用 HH:MM 或 HH 格式")
}

// PrintBill 打印账单明细
func PrintBill(result BillResult) {
	fmt.Println()
	fmt.Println("========== 账单明细 ==========")
	fmt.Printf("用电量：%.2f 度\n", result.Usage)
	fmt.Printf("用电时段：%02d:00（%s）\n", result.Hour, GetPeriodName(result.Period))
	fmt.Println("------------------------------")
	fmt.Printf("基础阶梯电费：%.2f 元\n", result.BaseCharge)

	if result.Period == PeakPeriod {
		fmt.Printf("时段调节：+%.0f%%（高峰时段）\n", result.AdjustmentRate*100)
	} else {
		fmt.Printf("时段调节：%.0f%%（低谷时段）\n", result.AdjustmentRate*100)
	}

	fmt.Println("------------------------------")
	fmt.Printf("最终电费：%.2f 元\n", result.FinalCharge)
	fmt.Println("==============================")
}

// RunBilling 运行计费流程（供main调用）
func RunBilling() {
	var usageInput string
	var timeInput string

	// 引导用户输入用电量
	fmt.Print("请输入用电量（比如 400.00）：")
	fmt.Scanln(&usageInput)

	// 解析用电量
	usage, err := strconv.ParseFloat(strings.TrimSpace(usageInput), 64)
	if err != nil {
		fmt.Printf("用电量格式错误：%v\n", err)
		return
	}

	// 引导用户输入用电时段
	fmt.Print("请输入用电时段（比如 14:00）：")
	fmt.Scanln(&timeInput)

	// 解析时间
	hour, err := ParseTimeInput(timeInput)
	if err != nil {
		fmt.Printf("时间格式错误：%v\n", err)
		return
	}

	// 计算电费
	result := CalculateBill(usage, hour)

	// 打印账单
	PrintBill(result)
}

func main() {
	RunBilling()
}
