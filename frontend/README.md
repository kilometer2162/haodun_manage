# 前端管理系统

## 技术栈
- Vue 3
- Element Plus
- Vue Router
- Pinia
- Axios
- Vite

## 安装依赖
```bash
cd frontend
npm install
```

## 开发
```bash
npm run dev
```

应用将在 `http://localhost:5173` 启动

## 构建
```bash
npm run build
```

## 功能模块

### 1. 用户认证
- 验证码登录
- JWT Token 认证
- 用户信息管理

### 2. RBAC权限管理
- 用户管理：增删改查
- 角色管理：角色创建、权限分配
- 权限管理：权限配置
- 资源管理：API资源、菜单、按钮

### 3. 日志管理
- 日志查看
- 多条件搜索（用户名、操作类型、模块、IP、状态、日期范围）
- 日志详情查看

### 4. IP访问统计
- IP访问记录查看
- 按日期统计访问情况
- 按地区统计访问情况
- 多条件搜索

## 默认账号
- 用户名: `admin`
- 密码: `admin123`

