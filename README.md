# 智能阶梯计费系统

## 一、基本信息

- **学校**：[请填写你的学校]
- **姓名**：[请填写你的姓名]
- **学号**：[请填写你的学号]

## 二、完成功能

- [x] 阶梯电价计算（三阶梯：200度/400度分界）
- [x] 峰谷时段判断（峰时、平时、谷时）
- [x] 时段电价倍率应用（峰时1.5倍、平时1.0倍、谷时0.5倍）
- [x] 违约金计算（宽限期、日费率、封顶限制）
- [x] 跨阶梯精确计费
- [x] 用户交互输入
- [x] 账单明细输出
- [x] 完整单元测试覆盖

## 三、关键逻辑思路

### 3.1 阶梯电价计算

采用分段累加的方式计算跨阶梯电费：

```
用电量 ≤ 200度：   电费 = 用电量 × 0.50元/度
200度 < 用电量 ≤ 400度： 电费 = 200×0.50 + (用电量-200)×0.60
用电量 > 400度：   电费 = 200×0.50 + 200×0.60 + (用电量-400)×0.80
```

### 3.2 峰谷时段判断

```
峰时：8:00-12:00，18:00-22:00（电价×1.5）
谷时：22:00-次日6:00（电价×0.5）
平时：其余时段（电价×1.0）
```

### 3.3 违约金计算

```
宽限期：10天（期内无违约金）
日费率：0.1%（欠费金额的千分之一/天）
封顶：账单金额的10%
```

## 四、函数列表

| 函数名 | 功能说明 |
|--------|---------|
| `GetTimePeriod(hour int) TimePeriod` | 根据小时判断当前属于峰时、平时还是谷时 |
| `GetTimeMultiplier(period TimePeriod) float64` | 获取指定时段的电价倍率 |
| `GetPeriodName(period TimePeriod) string` | 获取时段的中文名称 |
| `CalculateTieredCharge(usage float64) float64` | 计算基础阶梯电费（不含时段倍率） |
| `CalculateBillWithTime(usage float64, hour int) float64` | 计算含时段倍率的电费 |
| `CalculatePenalty(billAmount float64, overdueDays int) float64` | 计算违约金（含宽限期和封顶处理） |
| `CalculateBill(usage float64, hour int) BillResult` | 主计费函数，返回完整账单结果 |
| `CalculateFullBill(usage float64, hour int, overdueDays int) (float64, float64, float64)` | 计算包含违约金的完整账单 |
| `PrintBill(result BillResult)` | 格式化打印账单明细 |
| `RunBilling()` | 用户交互入口函数 |

## 五、单元测试截图

测试截图位于 `assets/` 目录下。

运行测试命令：
```bash
go test -v
```

![测试通过截图](./assets/test_result.png)

## 六、跨阶梯计费逻辑难点

### 6.1 难点描述

在实现跨阶梯计费时，最大的难点在于**如何正确处理用电量跨越多个阶梯的情况**。例如，当用电量为500度时，需要：

1. 正确识别用电量跨越了三个阶梯
2. 对每个阶梯内的用电量分别应用对应的电价
3. 确保阶梯边界的计算不重复、不遗漏

### 6.2 解决方案

采用**分段累加**的策略：

```go
func CalculateTieredCharge(usage float64) float64 {
    var charge float64 = 0.0
    
    if usage <= 0 {
        return 0.0
    }
    
    // 第一阶梯
    if usage <= Threshold1 {  // 200度
        return usage * Price1
    }
    charge = Threshold1 * Price1  // 第一阶梯满额
    
    // 第二阶梯
    if usage <= Threshold2 {  // 400度
        charge += (usage - Threshold1) * Price2
        return charge
    }
    charge += (Threshold2 - Threshold1) * Price2  // 第二阶梯满额
    
    // 第三阶梯
    charge += (usage - Threshold2) * Price3
    
    return charge
}
```

### 6.3 关键点总结

1. **边界判断顺序**：从低阶梯到高阶梯依次判断，避免重复计算
2. **满额计算**：当用电量超过某阶梯上限时，该阶梯按满额计算
3. **剩余电量**：最后一个阶梯只计算剩余用电量
4. **边界精确性**：确保200度、400度等边界值的计算精确无误

## 七、使用方法

```bash
# 运行程序
go run billing.go

# 运行测试
go test -v

# 运行测试（显示覆盖率）
go test -cover -v
```

## 八、示例运行

```
请输入用电量（度）：400
请输入用电时段（小时，0-23）：14
-------账单明细-------
当前用电：400.00度
当前时段：14:00点（平时）
基础电费：220.00元
时段倍率：1.0倍
最终电费：220.00元
---------------------
```
