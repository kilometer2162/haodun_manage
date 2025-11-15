# 后端管理系统

## 技术栈
- Go 1.21+
- Gin Web框架
- GORM ORM
- MySQL 数据库
- Redis 缓存
- JWT 认证

## 环境配置

### 方式一：使用.env文件（推荐）

1. 复制环境变量模板文件：
```bash
cd backend
cp .env.example .env
```

2. 编辑 `.env` 文件，修改为你自己的配置：
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=haodun_manage
REDIS_ADDR=localhost:6379
REDIS_PWD=
JWT_SECRET=your-secret-key-change-in-production
```

### 方式二：使用环境变量

如果不想使用.env文件，可以直接设置环境变量：
```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=haodun_manage
export REDIS_ADDR=localhost:6379
export REDIS_PWD=
export JWT_SECRET=your-secret-key-change-in-production
```

**注意**：环境变量的优先级高于.env文件中的配置。

## 安装依赖
```bash
cd backend
go mod download
```

## 数据库初始化

在启动项目之前，需要先初始化数据库：

```bash
# 使用MySQL命令行工具执行SQL脚本
mysql -u root -p < init.sql

# 或者使用source命令
mysql -u root -p
source init.sql
```

## 运行
```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动

## API 文档

### 认证
- POST `/api/auth/login` - 登录（需要验证码）
- GET `/api/auth/captcha` - 获取验证码
- GET `/api/auth/me` - 获取当前用户信息（需要认证）

### 用户管理
- GET `/api/users` - 获取用户列表
- GET `/api/users/:id` - 获取用户详情
- POST `/api/users` - 创建用户
- PUT `/api/users/:id` - 更新用户
- DELETE `/api/users/:id` - 删除用户

### 角色管理
- GET `/api/roles` - 获取角色列表
- GET `/api/roles/:id` - 获取角色详情
- POST `/api/roles` - 创建角色
- PUT `/api/roles/:id` - 更新角色
- DELETE `/api/roles/:id` - 删除角色
- POST `/api/roles/:id/permissions` - 分配权限

### 权限管理
- GET `/api/permissions` - 获取权限列表
- GET `/api/permissions/:id` - 获取权限详情
- POST `/api/permissions` - 创建权限
- PUT `/api/permissions/:id` - 更新权限
- DELETE `/api/permissions/:id` - 删除权限

### 资源管理
- GET `/api/resources` - 获取资源列表
- GET `/api/resources/:id` - 获取资源详情
- POST `/api/resources` - 创建资源
- PUT `/api/resources/:id` - 更新资源
- DELETE `/api/resources/:id` - 删除资源

### 日志管理
- GET `/api/logs` - 获取日志列表（支持搜索）
- GET `/api/logs/:id` - 获取日志详情

### IP访问统计
- GET `/api/ip-accesses` - 获取IP访问记录
- GET `/api/ip-statistics` - 获取IP统计信息

## 默认账号
- 用户名: `admin`
- 密码: `admin123`

