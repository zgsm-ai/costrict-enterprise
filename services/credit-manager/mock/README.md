# 维护公告 mock

运行 `npm run dev:mock` 后，GET `/costrict-static/announcement/credit-maintain.json` 返回 `announcement.ts` 中的原始 JSON（无 code/data 包装）。

- `maintain`：维护总开关。
- `start_time` / `end_time`：北京时间，格式 `YYYY-MM-DD HH:mm:ss`。开始时刻包含，结束时刻不包含；无效时间段不启用维护。
- `title`：公告标题。
- `impacts.quota.affected` / `description`：配额影响开关和说明；开启后用量统计不可用。
- `impacts.order.affected` / `description`：订单影响开关和说明；开启后订阅及购买页面不可用。

只有总开关开启且处于维护时间段内才显示公告、应用限制。两个影响开关均关闭时仍显示公告，页面保持可用。

示例保留 2026-09-09 16:30:00 至 17:00:00。联调时将时间改为覆盖当前北京时间，分别切换影响开关，验证菜单点击、直接 URL 访问、停留页面进入维护、结束后恢复。

客户端每秒判断时间，每 30 秒刷新公告，窗口重新聚焦时也刷新。404 表示无维护公告，会清除之前的公告及限制，不弹错误提示。JSON 无法解析、字段缺失或类型错误、空标题、受影响项缺少有效说明、日期不存在或时间段无效时，同样清除公告及限制，不向用户弹错误提示。其他错误首次获取失败时允许使用；后续失败保留上次公告，仍按结束时间解除限制。

普通开发模式通过 `PROXY_TARGET_STATIC` 配置静态服务代理。生产环境需在同源 `/costrict-static/announcement/credit-maintain.json` 提供该文件；mock 不打入生产包。
