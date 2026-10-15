# credit-manager

The backend management system for CoStrict

## 统一登录接入

登录状态只检查浏览器中的 `zgsmAdminToken` Cookie。每次业务请求读取最新 Cookie，
放入 `Authorization: Bearer <cookie值>`。入口 `/credit/manager?state=...&tab=usage`
保持兼容，`state` 不再用于兑换凭证，也不清理原地址；认证过程不请求 token、用户信息或 session 接口。
未登录显示登录页，点击登录进入 `/login-service/api/v1/web/login?return_to=...`，
登录完成回到原始完整地址。业务接口返回 401 时只提示“账号认证异常”，不删除 Cookie 或跳转登录。
基本信息从 Cookie 的 JWT 载荷读取：用户名 `display_name`、用户 ID `sub`、
配额 ID `universal_id`、工号 `employee_number`（缺失时显示 `-`）。不展示或请求邀请码。
载荷解码仅用于展示，不替代服务端认证；无法解码时显示占位符，不删除 Cookie。

部署时将 `/login-service` 代理到 loginservice（开发环境使用 `PROXY_TARGET_LOGIN`，
默认 `http://127.0.0.1:8080`），也可通过构建变量 `VITE_LOGIN_SERVICE_BASE_URL` 指定登录服务 origin。
loginservice 的 `web.return_to_allowlist` 需允许业务地址，例如
`https://业务域名/credit/manager*`，并使其 Cookie 的 domain/path 覆盖业务页面、
`http_only: false`，以便前端读取。开发环境需相应配置对外 origin、回调与回跳白名单；
Vite 代理不会自动改写服务端的这些配置。

奖励计划页的登录链接同样使用 loginservice。邀请码保留在回跳地址中；当前 loginservice
网页登录协议仅接收 `return_to`，尚不能替代旧登录接口的 `inviter_code` 注册绑定能力。
