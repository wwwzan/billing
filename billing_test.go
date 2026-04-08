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
		// 高峰时段测试：(8:00-22:00]
		{"高峰-8点", 8, PeakPeriod},
		{"高峰-9点", 9, PeakPeriod},
		{"高峰-11点", 11, PeakPeriod},
		{"高峰-12点", 12, PeakPeriod},
		{"高峰-14点", 14, PeakPeriod},
		{"高峰-18点", 18, PeakPeriod},
		{"高峰-20点", 20, PeakPeriod},
		{"高峰-21点", 21, PeakPeriod},

		// 低谷时段测试：(22:00-次日8:00]
		{"低谷-22点", 22, ValleyPeriod},
		{"低谷-23点", 23, ValleyPeriod},
		{"低谷-0点", 0, ValleyPeriod},
		{"低谷-5点", 5, ValleyPeriod},
		{"低谷-6点", 6, ValleyPeriod},
		{"低谷-7点", 7, ValleyPeriod},
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

func TestGetPeriodRateAdjustment(t *testing.T) {
	tests := []struct {
		name     string
		period   TimePeriod
		expected float64
	}{
		{"高峰调节", PeakPeriod, 1.10},   // +10%
		{"低谷调节", ValleyPeriod, 0.80}, // -20%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPeriodRateAdjustment(tt.period)
			if result != tt.expected {
				t.Errorf("GetPeriodRateAdjustment(%v) = %v, want %v", tt.period, result, tt.expected)
			}
		})
	}
}

// ==================== 时间解析测试 ====================

func TestParseTimeInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedHour int
		expectError  bool
	}{
		{"标准格式-14:00", "14:00", 14, false},
		{"标准格式-08:30", "08:30", 8, false},
		{"标准格式-22:00", "22:00", 22, false},
		{"仅小时-14", "14", 14, false},
		{"仅小时-8", "8", 8, false},
		{"带空格- 14:00 ", " 14:00 ", 14, false},
		{"带空格- 8", " 8", 8, false},
		{"凌晨-0:00", "0:00", 0, false},
		{"凌晨-0", "0", 0, false},
		{"错误格式-abc", "abc", 0, true},
		{"错误格式-25:00", "25:00", 0, true},
		{"错误格式--1", "-1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour, err := ParseTimeInput(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseTimeInput(%s) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseTimeInput(%s) unexpected error: %v", tt.input, err)
				}
				if hour != tt.expectedHour {
					t.Errorf("ParseTimeInput(%s) = %d, want %d", tt.input, hour, tt.expectedHour)
				}
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

		// 第一阶梯（0-200度）：0.5元/度
		{"第一阶梯-50度", 50, 50 * 0.50},
		{"第一阶梯-100度", 100, 100 * 0.50},
		{"第一阶梯-边界200度", 200, 200 * 0.50},

		// 第二阶梯（200-400度）：超过部分0.8元/度
		{"第二阶梯-250度", 250, 200*0.50 + 50*0.80},
		{"第二阶梯-300度", 300, 200*0.50 + 100*0.80},
		{"第二阶梯-边界400度", 400, 200*0.50 + 200*0.80},

		// 第三阶梯（400度以上）：超过部分1.2元/度
		{"第三阶梯-500度", 500, 200*0.50 + 200*0.80 + 100*1.20},
		{"第三阶梯-600度", 600, 200*0.50 + 200*0.80 + 200*1.20},
		{"第三阶梯-1000度", 1000, 200*0.50 + 200*0.80 + 600*1.20},
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
		{"100度-高峰(10点)", 100, 10, 100 * 0.50 * 1.10},  // +10%
		{"100度-低谷(23点)", 100, 23, 100 * 0.50 * 0.80}, // -20%
		{"100度-低谷(2点)", 100, 2, 100 * 0.50 * 0.80},   // -20%

		// 跨阶梯 + 不同时段
		{"300度-高峰(14点)", 300, 14, (200*0.50 + 100*0.80) * 1.10},
		{"300度-低谷(22点)", 300, 22, (200*0.50 + 100*0.80) * 0.80},
		{"300度-低谷(2点)", 300, 2, (200*0.50 + 100*0.80) * 0.80},

		// 第三阶梯 + 不同时段
		{"500度-高峰(9点)", 500, 9, (200*0.50 + 200*0.80 + 100*1.20) * 1.10},
		{"500度-低谷(3点)", 500, 3, (200*0.50 + 200*0.80 + 100*1.20) * 0.80},
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
			name:         "高峰计费-400度-14点",
			usage:        400,
			hour:         14,
			expectPeriod: PeakPeriod,
			expectCharge: (200*0.50 + 200*0.80) * 1.10, // 基础电费 * 1.10
		},
		{
			name:         "低谷计费-400度-23点",
			usage:        400,
			hour:         23,
			expectPeriod: ValleyPeriod,
			expectCharge: (200*0.50 + 200*0.80) * 0.80, // 基础电费 * 0.80
		},
		{
			name:         "低谷计费-400度-凌晨2点",
			usage:        400,
			hour:         2,
			expectPeriod: ValleyPeriod,
			expectCharge: (200*0.50 + 200*0.80) * 0.80,
		},
		{
			name:         "高峰边界-8点",
			usage:        200,
			hour:         8,
			expectPeriod: PeakPeriod,
			expectCharge: 200 * 0.50 * 1.10,
		},
		{
			name:         "低谷边界-22点",
			usage:        200,
			hour:         22,
			expectPeriod: ValleyPeriod,
			expectCharge: 200 * 0.50 * 0.80,
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

// ==================== 跨阶梯计费详细测试 ====================

func TestCrossTierBilling(t *testing.T) {
	// 测试跨阶梯计费的准确性
	t.Run("跨阶梯计费准确性", func(t *testing.T) {
		// 测试500度电费分解：
		// 第一阶梯：200度 × 0.50 = 100元
		// 第二阶梯：200度 × 0.80 = 160元
		// 第三阶梯：100度 × 1.20 = 120元
		// 总计：380元
		result := CalculateTieredCharge(500)
		expected := 200*0.50 + 200*0.80 + 100*1.20
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

// ==================== 需求示例验证测试 ====================

func TestRequirementExample(t *testing.T) {
	// 验证需求文档中的示例：400度，14:00
	t.Run("需求示例-400度-14点", func(t *testing.T) {
		usage := 400.0
		hour := 14

		result := CalculateBill(usage, hour)

		// 基础电费：200*0.5 + 200*0.8 = 100 + 160 = 260元
		expectedBase := 200*0.50 + 200*0.80
		if result.BaseCharge != expectedBase {
			t.Errorf("BaseCharge = %.2f, want %.2f", result.BaseCharge, expectedBase)
		}

		// 14点是高峰时段，费率+10%
		if result.Period != PeakPeriod {
			t.Errorf("Period should be PeakPeriod, got %v", result.Period)
		}

		// 最终电费：260 * 1.10 = 286元
		expectedFinal := expectedBase * 1.10
		if diff := result.FinalCharge - expectedFinal; diff > 0.01 || diff < -0.01 {
			t.Errorf("FinalCharge = %.2f, want %.2f", result.FinalCharge, expectedFinal)
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

func BenchmarkParseTimeInput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseTimeInput("14:00")
	}
}
