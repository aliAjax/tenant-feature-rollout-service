# Tenant Feature Rollout Service

纯Go特性开关与渐进式发布控制面，提供租户隔离的项目/环境/开关管理、规则匹配、确定性百分比分流、发布/暂停/归档/回滚、批量评估、变更游标和订阅接口。默认使用内存仓储以便单节点开发；`Repository`、`Publisher`、`Hasher`、`Clock`均为接口，可替换为PostgreSQL、Redis、签名凭据提供器和生产消息系统。

## 快速启动

```bash
go test ./...
go vet ./...
ROLLOUT_ADDR=:8087 go run ./cmd/rolloutd
```

### 创建并发布一个开关

```bash
curl -s -X POST localhost:8087/v1/projects -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"name":"checkout"}'
# 复制返回的 project id
curl -s -X POST localhost:8087/v1/environments -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"project_id":"prj_xxx","name":"production"}'
# 复制返回的 environment id
curl -s -X POST localhost:8087/v1/flags -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"project_id":"prj_xxx","environment_id":"env_xxx","key":"new_nav","type":"boolean","default":false}'
curl -s -X POST localhost:8087/v1/flags/prj_xxx/keys/new_nav/rules -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"actor":"release-bot","reason":"pilot","rule":{"priority":1,"percentage":25,"salt":"new-nav-v1","serve":true,"conditions":[{"attribute":"region","operator":"in","value":["JP","US"]}]}}'
curl -s -X POST localhost:8087/v1/flags/prj_xxx/keys/new_nav/publish -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"actor":"release-bot","reason":"production rollout"}'
curl -s -X POST localhost:8087/v1/evaluate -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"project_id":"prj_xxx","key":"new_nav","user_id":"user-42","attributes":{"region":"JP"}}'
```

评估返回`value`、强类型`type`、版本号、匹配规则ID和原因。相同用户、盐值和规则版本会得到稳定桶位；未匹配时返回默认值。`keys`数组可用于批量评估。暂停开关会立即返回默认值，`rollback`请求携带历史版本号并生成新的不可变版本。

## 端点

- `GET /healthz`、`GET /readyz`、`GET /metrics`
- `POST /v1/projects`、`POST /v1/environments`
- `POST /v1/flags`
- `POST /v1/flags/{project}/keys/{key}/rules`
- `POST /v1/flags/{project}/keys/{key}/publish|pause|archive|rollback`
- `POST /v1/evaluate`（单开关或`keys`批量）
- `GET /v1/changes?cursor=N`（版本游标增量）

所有请求支持`X-Request-ID`（缺失时自动生成）和`X-Tenant-ID`；响应为统一JSON错误格式。HTTP中间件包含CORS边界、超时、请求体限制、恢复和请求日志。服务收到SIGINT/SIGTERM后会优雅停机。

## 持久化与部署

`migrations/001_init.sql`和`002_audit.sql`定义项目、环境、开关版本、变更事件和客户端游标表；`deploy/docker-compose.yml`提供PostgreSQL和Redis开发依赖。生产部署应实现`Repository`和`Publisher`适配器，并将凭据只通过一次性响应或轮换接口返回，禁止写入日志。配置样例位于`configs/config.yaml`，`ROLLOUT_ADDR`可覆盖监听地址。

## 代码规模

非测试 Go 源码超过 3000 行，共 48 个功能文件；测试文件、迁移 SQL、配置、依赖和构建产物不计入。代码按项目、环境、开关、目标、发布、评估、分发、快照、治理和配额领域拆分，每个领域包含 domain、application、adapter、infrastructure 层。
