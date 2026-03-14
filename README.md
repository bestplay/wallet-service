# Wallet Service

一个支持 REST 和 gRPC 接口的钱包服务，支持内存和 MySQL 两种存储方式。

## 功能特性

- REST API 和 gRPC 双协议支持
- 支持内存存储和 MySQL 存储
- 支持水平扩展和负载均衡
- 精确的金额计算（使用 decimal）

## 项目结构

```
wallet-service/
├── cmd/
│   └── server/          # 服务入口
├── config/                # 配置文件
│   └── config.yaml
├── deploy/                # 部署配置
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── nginx.conf
├── internal/
│   ├── handler/          # HTTP 和 gRPC 处理器
│   ├── model/             # 数据模型
│   ├── proto/             # Protobuf 定义
│   ├── service/           # 业务逻辑
│   └── store/             # 存储层
└── test/                  # 测试文件
```

## 快速开始

### 本地运行

1. 克隆项目
```bash
git clone <repository-url>
cd wallet-service
```

2. 安装依赖
```bash
go mod download
```

3. 配置存储
编辑 `config/config.yaml`，选择存储类型：
```yaml
store:
  type: "memory"  # 或 "mysql"
```

4. 运行服务
```bash
go run cmd/server/main.go
```

服务将在以下端口启动：
- REST API: http://localhost:8080
- gRPC API: localhost:50051

### Docker 部署
```bash
cd deploy
docker-compose up -d --build
```

这将启动：
- 1 个 Nginx 负载均衡器
- 3 个钱包服务实例
- 1 个 MySQL 数据库

访问地址：
- REST API: http://localhost:8080
- gRPC API: localhost:50051

## API 文档

### REST API

#### 创建钱包
```bash
curl -X POST http://localhost:8080/wallets
```

#### 获取钱包
```bash
curl http://localhost:8080/wallets/{id}
```

#### 转账
```bash
curl -X POST http://localhost:8080/wallets/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "sourceId": "source-wallet-id",
    "destId": "dest-wallet-id",
    "amount": "100.00"
  }'
```

#### 充值
```bash
curl -X POST http://localhost:8080/wallets/recharge \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "wallet-id",
    "amount": "50.00"
  }'
```

### gRPC API

使用 grpcurl 工具测试 gRPC 接口：

```bash
# 安装 grpcurl
brew install grpcurl

# 创建钱包
grpcurl -plaintext -d '{}' localhost:50051 wallet.WalletService/CreateWallet

# 获取钱包
grpcurl -plaintext -d '{"id": "wallet-id"}' localhost:50051 wallet.WalletService/GetWallet

# 转账
grpcurl -plaintext -d '{
  "source_id": "source-id",
  "dest_id": "dest-id",
  "amount": "100.00"
}' localhost:50051 wallet.WalletService/Transfer

# 充值
grpcurl -plaintext -d '{
  "wallet_id": "wallet-id",
  "amount": "50.00"
}' localhost:50051 wallet.WalletService/Recharge
```

## 配置说明

### 存储配置

通过 `config/config.yaml` 或环境变量配置存储类型：

#### 内存存储（默认）
```yaml
store:
  type: "memory"
```

#### MySQL 存储
```yaml
store:
  type: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    user: "root"
    password: "password"
    database: "wallet"
```

或通过环境变量：
```bash
export STORE_TYPE=mysql
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=password
export MYSQL_DATABASE=wallet
```

## 水平扩展

### 架构说明

水平扩展配置包括：

1. **Nginx 负载均衡器**
   - 监听端口 8080 (REST) 和 50051 (gRPC)
   - 使用最小连接算法分发请求

2. **多个服务实例**
   - 默认配置 3 个钱包服务实例
   - 每个实例连接到共享的 MySQL 数据库

3. **共享数据库**
   - 所有服务实例共享同一个 MySQL 数据库
   - 确保数据一致性

### 扩展服务实例

要增加更多服务实例：

1. 编辑 `deploy/nginx.conf`，在 `upstream wallet_services` 中添加新的服务器：
```nginx
upstream wallet_services {
    least_conn;
    server wallet-service-1:8080;
    server wallet-service-2:8080;
    server wallet-service-3:8080;
    server wallet-service-4:8080;  # 添加新实例
}
```

2. 编辑 `deploy/docker-compose.yml`，添加新的服务定义：
```yaml
wallet-service-4:
  build:
    context: ..
    dockerfile: deploy/Dockerfile
  container_name: wallet-service-4
  environment:
    STORE_TYPE: "mysql"
    MYSQL_HOST: "mysql"
    MYSQL_PORT: "3306"
    MYSQL_USER: "walletuser"
    MYSQL_PASSWORD: "walletpassword"
    MYSQL_DATABASE: "wallet"
  depends_on:
    mysql:
      condition: service_healthy
  restart: unless-stopped
  networks:
    - wallet-network
```

3. 重启服务：
```bash
cd deploy
docker-compose down
docker-compose up -d --build
```

### 监控和日志

```bash
# 查看所有服务状态
docker-compose ps

# 查看 Nginx 日志
docker-compose logs nginx

# 查看特定服务日志
docker-compose logs wallet-service-1

# 查看所有服务日志
docker-compose logs -f
```

## 测试

### 集成测试

```bash
# 运行所有测试
go test ./test -v

# 运行特定测试
go test ./test -run TestWalletConcurrency -v
```

### 负载测试

```bash
cd test
k6 run k6-test.js
```

详细测试说明请参考 [test/README.md](./test/README.md)

## 其他注意事项

由于项目题目要求中，钱包初始化余额为 0，可以使用 /recharge 接口充值后，测试转账功能

### 技术栈

## 技术栈

- Go 1.26
- Gin (HTTP 框架)
- gRPC (RPC 框架)
- Protocol Buffers (序列化)
- MySQL 8.0 (数据库)
- Nginx (负载均衡)
- Docker & Docker Compose (容器化)

## 许可证

MIT License
