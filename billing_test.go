package main

import (
	"testing"
)

// ==================== 时段判断测试 ====================

func TestGetTimePeriod(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		expected TimePeriod
	}{
		// 峰时测试：8:00-12:00
		{"峰时-8点", 8, PeakPeriod},
		{"峰时-9点", 9, PeakPeriod},
		{"峰时-11点", 11, PeakPeriod},
		
		// 峰时测试：18:00-22:00
		{"峰时-18点", 18, PeakPeriod},
		{"峰时-20点", 20, PeakPeriod},
		{"峰时-21点", 21, PeakPeriod},
		
		// 谷时测试：22:00-次日6:00
		{"谷时-22点", 22, ValleyPeriod},
		{"谷时-23点", 23, ValleyPeriod},
		{"谷时-0点", 0, ValleyPeriod},
		{"谷时-5点", 5, ValleyPeriod},
		
		// 平时测试
		{"平时-6点", 6, NormalPeriod},
		{"平时-7点", 7, NormalPeriod},
		{"平时-12点", 12, NormalPeriod},
		{"平时-15点", 15, NormalPeriod},
		{"平时-17点", 17, NormalPeriod},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimePeriod(tt.hour)
			if result != tt.expected {
				t.Errorf("GetTimePeriod(%d) = %v, want %v", tt.hour, result, tt.expected)
			}
		})
	}
}

func TestGetTimeMultiplier(t *testing.T) {
	tests := []struct {
		name     string
		period   TimePeriod
		expected float64
	}{
		{"峰时倍率", PeakPeriod, 1.5},
		{"平时倍率", NormalPeriod, 1.0},
		{"谷时倍率", ValleyPeriod, 0.5},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimeMultiplier(tt.period)
			if result != tt.expected {
				t.Errorf("GetTimeMultiplier(%v) = %v, want %v", tt.period, result, tt.expected)
			}
		})
	}
}

// ==================== 阶梯电费计算测试 ====================

func TestCalculateTieredCharge(t *testing.T) {
	tests := []struct {
		name     string
		usage    float64
		expected float64
	}{
		// 边界测试
		{"零用电", 0, 0},
		{"负数用电", -10, 0},
		
		// 第一阶梯（0-200度）
		{"第一阶梯-50度", 50, 50 * 0.50},
		{"第一阶梯-100度", 100, 100 * 0.50},
		{"第一阶梯-边界200度", 200, 200 * 0.50},
		
		// 第二阶梯（200-400度）
		{"第二阶梯-250度", 250, 200*0.50 + 50*0.60},
		{"第二阶梯-300度", 300, 200*0.50 + 100*0.60},
		{"第二阶梯-边界400度", 400, 200*0.50 + 200*0.60},
		
		// 第三阶梯（400度以上）
		{"第三阶梯-500度", 500, 200*0.50 + 200*0.60 + 100*0.80},
		{"第三阶梯-600度", 600, 200*0.50 + 200*0.60 + 200*0.80},
		{"第三阶梯-1000度", 1000, 200*0.50 + 200*0.60 + 600*0.80},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTieredCharge(tt.usage)
			// 使用容差比较浮点数
			if diff := result - tt.expected; diff > 0.01 || diff < -0.01 {
				t.Errorf("CalculateTieredCharge(%.2f) = %.2f, want %.2f", tt.usage, result, tt.expected)
			}
		})
	}
}

// ==================== 含时段电费计算测试 ====================

func TestCalculateBillWithTime(t *testing.T) {
	tests := []struct {
		name     string
		usage    float64
		hour     int
		expected float64
	}{
		// 第一阶梯 + 不同时段
		{"100度-峰时(10点)", 100, 10, 100 * 0.50 * 1.5},
		{"100度-平时(14点)", 100, 14, 100 * 0.50 * 1.0},
		{"100度-谷时(23点)", 100, 23, 100 * 0.50 * 0.5},
		
		// 跨阶梯 + 不同时段
		{"300度-峰时(20点)", 300, 20, (200*0.50 + 100*0.60) * 1.5},
		{"300度-平时(15点)", 300, 15, (200*0.50 + 100*0.60) * 1.0},
		{"300度-谷时(2点)", 300, 2, (200*0.50 + 100*0.60) * 0.5},
		
		// 第三阶梯 + 不同时段
		{"500度-峰时(9点)", 500, 9, (200*0.50 + 200*0.60 + 100*0.80) * 1.5},
		{"500度-平时(13点)", 500, 13, (200*0.50 + 200*0.60 + 100*0.80) * 1.0},
		{"500度-谷时(3点)", 500, 3, (200*0.50 + 200*0.60 + 100*0.80) * 0.5},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBillWithTime(tt.usage, tt.hour)
			if diff := result - tt.expected; diff > 0.01 || diff < -0.01 {
				t.Errorf("CalculateBillWithTime(%.2f, %d) = %.2f, want %.2f", tt.usage, tt.hour, result, tt.expected)
			}
		})
	}
}

// ==================== 违约金计算测试 ====================

func TestCalculatePenalty(t *testing.T) {
	tests := []struct {
		name        string
		billAmount  float64
		overdueDays int
		expected    float64
	}{
		// 宽限期内
		{"宽限期内-5天", 100, 5, 0},
		{"宽限期内-10天", 100, 10, 0},
		
		// 超过宽限期
		{"超宽限期-11天", 100, 11, 100 * 0.001 * 1},  // 1天违约金
		{"超宽限期-20天", 100, 20, 100 * 0.001 * 10}, // 10天违约金
		{"超宽限期-30天", 100, 30, 100 * 0.001 * 20}, // 20天违约金
		
		// 违约金封顶测试（最大10%）
		{"封顶测试-200天", 100, 200, 10},  // 应该封顶在10元
		{"封顶测试-365天", 100, 365, 10}, // 应该封顶在10元
		
		// 大额账单测试
		{"大额账单-30天", 1000, 30, 1000 * 0.001 * 20},
		{"大额账单封顶", 1000, 500, 100}, // 封顶在100元
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePenalty(tt.billAmount, tt.overdueDays)
			if diff := result - tt.expected; diff > 0.01 || diff < -0.01 {
				t.Errorf("CalculatePenalty(%.2f, %d) = %.2f, want %.2f", 
					tt.billAmount, tt.overdueDays, result, tt.expected)
			}
		})
	}
}

// ==================== 主计费函数测试 ====================

func TestCalculateBill(t *testing.T) {
	tests := []struct {
		name           string
		usage          float64
		hour           int
		expectPeriod   TimePeriod
		expectCharge   float64
	}{
		{
			name:         "正常计费-400度-14点",
			usage:        400,
			hour:         14,
			expectPeriod: NormalPeriod,
			expectCharge: (200*0.50 + 200*0.60) * 1.0,
		},
		{
			name:         "峰时计费-400度-10点",
			usage:        400,
			hour:         10,
			expectPeriod: PeakPeriod,
			expectCharge: (200*0.50 + 200*0.60) * 1.5,
		},
		{
			name:         "谷时计费-400度-23点",
			usage:        400,
			hour:         23,
			expectPeriod: ValleyPeriod,
			expectCharge: (200*0.50 + 200*0.60) * 0.5,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBill(tt.usage, tt.hour)
			
			if result.Period != tt.expectPeriod {
				t.Errorf("Period = %v, want %v", result.Period, tt.expectPeriod)
			}
			
			if diff := result.FinalCharge - tt.expectCharge; diff > 0.01 || diff < -0.01 {
				t.Errorf("FinalCharge = %.2f, want %.2f", result.FinalCharge, tt.expectCharge)
			}
		})
	}
}

// ==================== 边界条件测试 ====================

func TestBoundaryConditions(t *testing.T) {
	// 测试无效小时数
	t.Run("无效小时数处理", func(t *testing.T) {
		// 小时数小于0，应被修正为0
		result1 := CalculateBill(100, -1)
		if result1.Hour != 0 {
			t.Errorf("Hour should be 0, got %d", result1.Hour)
		}
		
		// 小时数大于23，应被修正为23
		result2 := CalculateBill(100, 25)
		if result2.Hour != 23 {
			t.Errorf("Hour should be 23, got %d", result2.Hour)
		}
	})
	
	// 测试负数用电量
	t.Run("负数用电量处理", func(t *testing.T) {
		result := CalculateBill(-100, 10)
		if result.Usage != 0 {
			t.Errorf("Usage should be 0, got %.2f", result.Usage)
		}
	})
}

// ==================== 完整账单测试 ====================

func TestCalculateFullBill(t *testing.T) {
	tests := []struct {
		name        string
		usage       float64
		hour        int
		overdueDays int
	}{
		{"完整账单-400度-14点-无欠费", 400, 14, 5},
		{"完整账单-400度-14点-欠费20天", 400, 14, 20},
		{"完整账单-400度-峰时-欠费30天", 400, 10, 30},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charge, penalty, total := CalculateFullBill(tt.usage, tt.hour, tt.overdueDays)
			
			// 验证总金额 = 电费 + 违约金
			if total != charge+penalty {
				t.Errorf("Total %.2f != Charge %.2f + Penalty %.2f", total, charge, penalty)
			}
			
			// 验证金额非负
			if charge < 0 || penalty < 0 || total < 0 {
				t.Errorf("Negative values detected: charge=%.2f, penalty=%.2f, total=%.2f", 
					charge, penalty, total)
			}
		})
	}
}

// ==================== 跨阶梯计费详细测试 ====================

func TestCrossTierBilling(t *testing.T) {
	// 测试跨阶梯计费的准确性
	t.Run("跨阶梯计费准确性", func(t *testing.T) {
		// 测试500度电费分解：
		// 第一阶梯：200度 × 0.50 = 100元
		// 第二阶梯：200度 × 0.60 = 120元
		// 第三阶梯：100度 × 0.80 = 80元
		// 总计：300元
		result := CalculateTieredCharge(500)
		expected := 200*0.50 + 200*0.60 + 100*0.80
		if result != expected {
			t.Errorf("Cross-tier calculation: got %.2f, want %.2f", result, expected)
		}
	})
	
	// 测试阶梯边界
	t.Run("阶梯边界测试", func(t *testing.T) {
		// 恰好200度（第一阶梯边界）
		r1 := CalculateTieredCharge(200)
		// 201度（刚进入第二阶梯）
		r2 := CalculateTieredCharge(201)
		// 第二阶梯的电费应该更高
		if r2 <= r1 {
			t.Errorf("201度电费(%.2f)应该大于200度电费(%.2f)", r2, r1)
		}
		
		// 恰好400度（第二阶梯边界）
		r3 := CalculateTieredCharge(400)
		// 401度（刚进入第三阶梯）
		r4 := CalculateTieredCharge(401)
		// 第三阶梯的电费应该更高
		if r4 <= r3 {
			t.Errorf("401度电费(%.2f)应该大于400度电费(%.2f)", r4, r3)
		}
	})
}

// ==================== 基准测试 ====================

func BenchmarkCalculateBill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateBill(400, 14)
	}
}

func BenchmarkCalculateTieredCharge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateTieredCharge(500)
	}
}

func BenchmarkCalculatePenalty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculatePenalty(100, 30)
	}
}
