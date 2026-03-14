# Wallet Service - 测试说明

## 目录

1. [集成测试和并发测试](#集成测试和并发测试)
2. [k6 负载测试](#k6-负载测试)

---

## 集成测试和并发测试

使用 Go 原生测试框架编写的并发测试，用于验证系统在高并发场景下的正确性。

### 运行测试

```bash
# 运行所有测试
go test ./test -v

# 运行特定测试
go test ./test -run TestWalletConcurrency -v
go test ./test -run TestRechargeConcurrency -v
go test ./test -run TestTransferConcurrency -v
```

### 测试用例说明

1. **TestWalletConcurrency**
   - 测试 100 个并发创建钱包操作
   - 验证所有钱包创建成功且获得唯一 ID

2. **TestRechargeConcurrency**
   - 测试对同一个钱包进行 100 次并发充值
   - 每次充值 10.0 单位
   - 验证最终余额准确等于 1000.0

3. **TestTransferConcurrency**
   - 测试从源钱包到目标钱包 100 次并发转账
   - 每次转账 10.0 单位
   - 源钱包初始余额 10000.0，目标钱包初始余额 0
   - 验证源钱包最终余额 9000.0，目标钱包最终余额 1000.0

---

## k6 负载测试

使用 k6 工具进行负载和并发测试，用于评估系统在不同负载下的性能表现。

### 前置条件

1. 安装 k6：
   ```bash
   # macOS
   brew install k6
   
   # 或参考官方文档：https://k6.io/docs/getting-started/installation/
   ```

2. 启动钱包服务：
   ```bash
   go run ./cmd/server
   ```

### 运行测试

```bash
# 启动服务后
cd test
k6 run k6-test.js
```

### 测试配置说明

测试分为 3 个阶段：

1. **渐增阶段** (30秒)：从 0 增加到 20 个虚拟用户
2. **稳定阶段** (1分钟)：保持 50 个虚拟用户
3. **渐减阶段** (30秒)：从 50 减少到 0 个虚拟用户

### 性能阈值

- **响应时间**：95% 的请求响应时间 < 500ms
- **错误率**：失败请求率 < 1%

### 测试场景

每个虚拟用户随机执行以下操作之一：

1. **创建钱包** (30% 概率)
2. **查询钱包** (50% 概率)
3. **转账** (20% 概率)

### 自定义测试配置

可以修改 `k6-test.js` 中的 `options` 对象来自定义测试：

```javascript
export const options = {
  stages: [
    { duration: '30s', target: 20 },    // 30秒内增加到20个用户
    { duration: '1m', target: 50 },     // 1分钟内保持50个用户
    { duration: '30s', target: 0 },      // 30秒内减少到0个用户
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],    // 95%请求<500ms
    http_req_failed: ['rate<0.01'],       // 失败率<1%
  },
};
```
