#!/usr/bin/env bash
# 接口冒烟测试：自动启动服务、依次调用全部接口、结束前关闭服务。
#
# 用法：bash scripts/smoke_test.sh
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${PORT:-18080}"
BASE="http://127.0.0.1:${PORT}/api"
BIN="$(mktemp -d)/blog-server"
LOG="$(mktemp)"

export JWT_SECRET="smoke-test-secret"
export BLOG_ADDR=":${PORT}"
export BLOG_DB="$(mktemp -d)/smoke.db"

cleanup() {
  [ -n "${SERVER_PID:-}" ] && kill "$SERVER_PID" 2>/dev/null
}
trap cleanup EXIT

LAST_BODY=""

call() {
  # call <说明> <期望状态码> <curl 参数...>
  local title="$1" expect="$2"
  shift 2
  printf '\n=== %s ===\n' "$title"
  local body code
  body="$(curl -s -w '\n%{http_code}' "$@")"
  code="$(printf '%s' "$body" | tail -n1)"
  body="$(printf '%s' "$body" | sed '$d')"
  printf 'HTTP %s (期望 %s)\n%s\n' "$code" "$expect" "$body"
  LAST_BODY="$body"
}

extract_token() {
  printf '%s' "$LAST_BODY" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
}

printf '编译服务...\n'
( cd "$ROOT" && go build -o "$BIN" . ) || exit 1

printf '启动服务 (端口 %s)...\n' "$PORT"
"$BIN" >"$LOG" 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 60); do
  curl -s -o /dev/null "$BASE/posts" && break
  sleep 0.5
done

USER_A="alice_$$"
USER_B="bob_$$"
PASSWORD="secret123"

# ---------- 用户认证 ----------

call "1. 注册用户 A（201）" 201 -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_A\",\"password\":\"$PASSWORD\",\"email\":\"$USER_A@example.com\"}"

call "2. 重复注册同一用户名（409）" 409 -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_A\",\"password\":\"$PASSWORD\",\"email\":\"other@example.com\"}"

call "3. 注册参数不合法：密码过短（400）" 400 -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"username":"shortpwd","password":"123","email":"shortpwd@example.com"}'

call "4. 登录成功，取得 token（200）" 200 -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_A\",\"password\":\"$PASSWORD\"}"
TOKEN_A="$(extract_token)"

call "5. 登录密码错误（401）" 401 -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_A\",\"password\":\"wrong-password\"}"

# ---------- 文章读写 ----------

call "6. 未携带 token 创建文章（401）" 401 -X POST "$BASE/posts" \
  -H 'Content-Type: application/json' \
  -d '{"title":"未认证","content":"应当被拒绝"}'

call "7. token 格式错误（401）" 401 -X POST "$BASE/posts" \
  -H 'Authorization: Bearer not-a-real-token' \
  -H 'Content-Type: application/json' \
  -d '{"title":"伪造 token","content":"应当被拒绝"}'

call "8. 用户 A 创建文章（201）" 201 -X POST "$BASE/posts" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H 'Content-Type: application/json' \
  -d '{"title":"第一篇文章","content":"Gin + GORM + JWT 示例"}'
POST_ID="$(printf '%s' "$LAST_BODY" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')"

call "9. 获取文章列表（200）" 200 "$BASE/posts?page=1&page_size=10"

call "10. 获取单篇文章详情（200）" 200 "$BASE/posts/$POST_ID"

call "11. 作者本人更新文章（200）" 200 -X PUT "$BASE/posts/$POST_ID" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H 'Content-Type: application/json' \
  -d '{"title":"第一篇文章（已更新）","content":"更新后的内容"}'

# ---------- 越权校验 ----------

call "12. 注册并登录用户 B（200）" 200 -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_B\",\"password\":\"$PASSWORD\",\"email\":\"$USER_B@example.com\"}"
call "12b. 用户 B 登录（200）" 200 -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER_B\",\"password\":\"$PASSWORD\"}"
TOKEN_B="$(extract_token)"

call "13. 用户 B 更新他人文章（403）" 403 -X PUT "$BASE/posts/$POST_ID" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"title":"越权修改","content":"应当被拒绝"}'

call "14. 用户 B 删除他人文章（403）" 403 -X DELETE "$BASE/posts/$POST_ID" \
  -H "Authorization: Bearer $TOKEN_B"

# ---------- 评论 ----------

call "15. 用户 B 对文章发表评论（201）" 201 -X POST "$BASE/posts/$POST_ID/comments" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"content":"这是用户 B 的评论"}'

call "16. 获取评论列表（200）" 200 "$BASE/posts/$POST_ID/comments"

call "17. 对不存在的文章发表评论（404）" 404 -X POST "$BASE/posts/999999/comments" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"content":"评论不存在的文章"}'

# ---------- 删除与收尾 ----------

call "18. 获取不存在的文章（404）" 404 "$BASE/posts/999999"

call "19. 作者本人删除文章（200）" 200 -X DELETE "$BASE/posts/$POST_ID" \
  -H "Authorization: Bearer $TOKEN_A"

call "20. 删除后再查（404）" 404 "$BASE/posts/$POST_ID"

printf '\n=== 服务端日志（末尾 12 行）===\n'
tail -n 12 "$LOG"

printf '\n全部用例执行完毕。\n'
