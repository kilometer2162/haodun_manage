# 管理系统

一个完整的前后端分离管理系统，包含RBAC权限管理、日志记录、验证码登录、IP访问统计等功能。

## 项目结构

```
haodun_manage/
├── backend/          # 后端（Golang）
│   ├── config/       # 配置
│   ├── controllers/  # 控制器
│   ├── database/     # 数据库
│   ├── middleware/   # 中间件
│   ├── models/       # 数据模型
│   ├── router/       # 路由
│   ├── utils/        # 工具函数
│   └── main.go       # 入口文件
└── frontend/         # 前端（Vue3 + ElementUI）
    ├── src/
    │   ├── layouts/  # 布局组件
    │   ├── views/    # 页面组件
    │   ├── router/   # 路由配置
    │   ├── stores/   # 状态管理
    │   └── utils/    # 工具函数
    └── package.json
```

## 功能特性

### 1. RBAC权限管理
- ✅ 用户管理（增删改查）
- ✅ 角色管理（角色创建、权限分配）
- ✅ 权限管理（权限配置）
- ✅ 资源管理（API资源、菜单、按钮）

### 2. 日志系统
- ✅ 操作日志记录
- ✅ 日志搜索（多条件筛选）
- ✅ 日志详情查看

### 3. 验证码登录
- ✅ 验证码生成（Redis存储）
- ✅ 验证码验证
- ✅ JWT Token认证

### 4. IP访问统计
- ✅ IP访问记录
- ✅ 按日期统计
- ✅ 按地区统计
- ✅ IP信息搜索

## 快速开始

### 后端启动

1. 安装依赖
```bash
cd backend
go mod download
```

2. 配置环境变量

   方式一：使用.env文件（推荐）
   ```bash
   cd backend
   cp .env.example .env
   # 然后编辑.env文件，修改为你自己的配置
   ```

   方式二：直接设置环境变量
   ```bash
   export DB_HOST=localhost
   export DB_PORT=3306
   export DB_USER=root
   export DB_PASSWORD=your_password
   export DB_NAME=haodun_manage
   export REDIS_ADDR=localhost:6379
   export REDIS_PWD=
   export JWT_SECRET=your-secret-key
   ```

3. 初始化数据库
   ```bash
   cd backend
   mysql -u root -p < init.sql
   ```

4. 启动服务
   ```bash
   go run main.go
   ```

后端服务将在 `http://localhost:8080` 启动

### 前端启动

1. 安装依赖
```bash
cd frontend
npm install
```

2. 启动开发服务器
```bash
npm run dev
```

前端应用将在 `http://localhost:5173` 启动

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

## 技术栈

### 后端
- Go 1.21+
- Gin Web框架
- GORM ORM
- MySQL 数据库
- Redis 缓存
- JWT 认证

### 前端
- Vue 3
- Element Plus
- Vue Router
- Pinia
- Axios
- Vite

## API文档

详细API文档请参考 [backend/README.md](./backend/README.md)

## 开发说明

### 数据库初始化
系统启动时会自动创建数据库表并初始化默认管理员账号。

### 验证码
验证码通过Redis存储，5分钟过期。生产环境建议将验证码生成图片而不是直接返回文本。

### IP地区信息
当前IP地区信息默认为"未知"，可以集成IP查询API（如ipip.net、ip2region等）来获取真实的地理位置信息。

## 许可证

MIT

