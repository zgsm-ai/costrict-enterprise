# 模拟 GitHub Star Webhook

`github_star_webhook.py` 使用 Python 3 标准库构造 GitHub `Stars` 事件，计算 HMAC-SHA256 签名并发送请求，无需安装依赖。请在项目根目录执行以下命令。

## 服务端准备

- 启用 `syncStar.webhook.enabled`，配置非空的 `syncStar.webhook.secret`。
- 配置目标 `syncStar.owner` 和 `syncStar.repo`，请求中的仓库必须与其一致。
- 确保 `/oidc-auth/api/v1/webhooks/github` 可以访问，且 POST 请求转发到 oidc-auth。
- 选择本地数据库中已有的测试用户，使用其 `github_id`（GitHub 数字 ID，不是用户名或本地 UUID）。

模拟请求会真实修改目标服务的数据库，但不会在 GitHub 上执行加星或取消操作。定时同步开启时，下一轮会按 GitHub 的真实状态覆盖模拟结果。

## 加星

替换地址和用户 ID；`owner`、`repo` 默认分别为 `zgsm-ai`、`costrict`：

```bash
python3 scripts/github_star_webhook.py \
  --url http://127.0.0.1:8080/oidc-auth/api/v1/webhooks/github \
  --owner zgsm-ai \
  --repo costrict \
  --user-id 123456 \
  --action created
```

脚本会提示输入 Webhook Secret，输入不回显。请使用与服务端完全一致的密钥。`--login` 可选，仅用于服务端日志，不影响数据库用户匹配。

## 取消加星

使用相同参数，将 `--action created` 改为 `--action deleted` 后重新运行。

## 环境变量方式

支持通过环境变量提供参数，命令行参数优先。下面的密钥输入命令适用于 Bash：

```bash
export WEBHOOK_URL='https://<线上认证服务域名>/oidc-auth/api/v1/webhooks/github'
export GITHUB_OWNER='zgsm-ai'
export GITHUB_REPO='costrict'
export GITHUB_USER_ID='123456'
export STAR_ACTION='created'
read -rs -p 'Webhook Secret: ' WEBHOOK_SECRET
echo
export WEBHOOK_SECRET

python3 scripts/github_star_webhook.py
python3 scripts/github_star_webhook.py --action deleted

unset WEBHOOK_SECRET
```

`WEBHOOK_SECRET` 是本脚本读取的变量，服务端对应变量是 `SYNCSTAR_WEBHOOK_SECRET`。非交互执行时必须设置 `WEBHOOK_SECRET`。脚本不会自动重试或跟随重定向；请求超时为 15 秒。

## 检查结果

| HTTP 状态 | 含义 |
| --- | --- |
| `204` | 已处理；没有匹配用户时也会返回此状态 |
| `401` | Secret 或签名不正确 |
| `400` | 仓库不匹配、用户 ID 或请求体不合法 |
| `404` | Webhook 未启用，或地址、网关路由不正确 |
| `500` | 数据库更新失败 |

脚本仅在收到 `204` 时返回退出码 `0`；其他 HTTP 响应或网络错误返回 `1`，参数错误返回 `2`。查看参数说明：

```bash
python3 scripts/github_star_webhook.py --help
```

在默认表名配置下查询数据库，替换为实际测试用户的 GitHub ID：

```sql
SELECT id, github_id, github_star
FROM auth_users
WHERE github_id = '123456';
```

- 加星后 `github_star` 应为 `zgsm-ai.costrict`（格式为 `owner.repo`）。
- 取消后，仅当原字段等于目标仓库标记时才会清空。
- 没有匹配用户时不会创建用户；可结合服务端日志中的 `rows` 判断实际更新数量。
- 当前 `userinfo` 接口不返回 `isStar`，请通过数据库或服务端日志检查。
