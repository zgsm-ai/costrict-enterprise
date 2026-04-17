# 登录后重定向到目标页面 — 设计文档

**日期：** 2026-04-15  
**分支：** feature/optimize-site-v1

---

## 背景

用户直接访问受保护路由（如 `/credits`、`/subscribe`）时，当前实现会将其重定向到 `/login`，登录完成后始终回到首页 `/`，丢失了用户的原始访问意图。

---

## 目标

用户访问需要登录的页面时，完成登录后自动跳转回原始目标页面。

---

## 范围

- **受保护路由**：除 `PUBLIC_ROUTES`（`/credit-reward-plan`、`/credit-md-preview`）和 `/login` 之外的所有路由
- **公开路由**：不受影响，无需登录即可访问

---

## 方案：localStorage 暂存目标路径

### 选型理由

当前登录流程是完整的页面外跳（`window.location.href = loginUrl`），OIDC 登录完成后回到应用入口时，URL query 参数已丢失。使用 `localStorage` 持久化目标路径，跨页面跳转后仍可读取，改动范围最小。

---

## 详细设计

### localStorage key

```
key:   'auth_redirect'
value: 完整的目标路径，含 query 和 hash，例如 '/credits?tab=history'
```

### 变更一：认证守卫（`src/router/guards/auth.ts`）

**触发条件：** 用户访问受保护路由，且 `isAuthenticated()` 返回 `false`，且后台认证也失败。

**行为：**
1. 将 `to.fullPath` 写入 `localStorage['auth_redirect']`
2. 执行 `router.replace('/login')`

**不写入的情况：**
- 目标路径是 `/login` 本身
- 目标路径在 `PUBLIC_ROUTES` 中
- 用户已登录（直接放行）

### 变更二：认证服务（`src/services/auth.ts`）

**触发条件：** `authenticate()` 执行成功后。

**行为：**
1. 读取 `localStorage['auth_redirect']`
2. 若存在：清除该 key，调用 `router.replace(redirectPath)` 跳转
3. 若不存在：保持现有逻辑不变

**注意：** `router` 实例通过参数注入或 `import` 方式传入，避免循环依赖。

---

## 数据流

```
用户访问 /credits
    ↓
auth guard: isAuthenticated() → false
    ↓
localStorage.setItem('auth_redirect', '/credits')
    ↓
router.replace('/login')
    ↓
用户点击登录按钮 → window.location.href = OIDC URL
    ↓
OIDC 登录完成 → 回调到应用（携带 token hash）
    ↓
authService.authenticate() 成功
    ↓
读取 localStorage['auth_redirect'] → '/credits'
清除 key
router.replace('/credits')
```

---

## 边界情况

| 场景 | 处理方式 |
|------|----------|
| 用户直接访问 `/login` | 不写入 `auth_redirect` |
| 用户访问公开路由 | 不写入 `auth_redirect` |
| 登录成功但无 `auth_redirect` | 保持现有逻辑（跳首页或当前页） |
| 多 tab 同时登录 | 最后一次写入的路径生效（可接受） |
| `auth_redirect` 值为非法路径 | Vue Router 会处理，跳转失败则留在当前页 |

---

## 涉及文件

| 文件 | 改动 |
|------|------|
| `src/router/guards/auth.ts` | 认证失败时写入 `localStorage['auth_redirect']` |
| `src/services/auth.ts` | 认证成功后读取并清除 `localStorage['auth_redirect']`，执行跳转 |

`src/views/Login/login-page.vue` 无需改动。
