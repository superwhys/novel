#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-${VERSION:-dev}}"
SSH_HOST="${SSH_HOST:-ali-prod}"
REMOTE_DIR="${REMOTE_DIR:-/opt/novel}"
REMOTE_FRONTEND_DIR="${REMOTE_FRONTEND_DIR:-/etc/nginx/www/novel.superwhys.top}"
SERVICE_NAME="${SERVICE_NAME:-novel}"
RELEASE_NAME="novel-${VERSION}-linux-amd64"
ARCHIVE="${ROOT_DIR}/dist/${RELEASE_NAME}.tar.gz"
REMOTE_ARCHIVE="/tmp/${RELEASE_NAME}.tar.gz"

if [[ ! "${VERSION}" =~ ^[a-zA-Z0-9][a-zA-Z0-9._-]*$ ]]; then
	echo "错误：版本号只能包含字母、数字、点、下划线和短横线。" >&2
	exit 1
fi

echo "[1/4] 构建 ${RELEASE_NAME}"
make -C "${ROOT_DIR}" package VERSION="${VERSION}"

echo "[2/4] 上传到 ${SSH_HOST}:${REMOTE_ARCHIVE}"
scp "${ARCHIVE}" "${SSH_HOST}:${REMOTE_ARCHIVE}"

printf -v REMOTE_COMMAND 'bash -s -- %q %q %q %q %q' \
	"${REMOTE_ARCHIVE}" "${RELEASE_NAME}" "${REMOTE_DIR}" \
	"${REMOTE_FRONTEND_DIR}" "${SERVICE_NAME}"

echo "[3/4] 替换后端 ${REMOTE_DIR} 和前端 ${REMOTE_FRONTEND_DIR}"
ssh "${SSH_HOST}" "${REMOTE_COMMAND}" <<'REMOTE_SCRIPT'
set -Eeuo pipefail

ARCHIVE="$1"
RELEASE_NAME="$2"
REMOTE_DIR="$3"
REMOTE_FRONTEND_DIR="$4"
SERVICE_NAME="$5"
REMOTE_PARENT="$(dirname -- "${REMOTE_DIR}")"
REMOTE_BASENAME="$(basename -- "${REMOTE_DIR}")"
FRONTEND_PARENT="$(dirname -- "${REMOTE_FRONTEND_DIR}")"
FRONTEND_BASENAME="$(basename -- "${REMOTE_FRONTEND_DIR}")"
EXTRACT_DIR="/tmp/${RELEASE_NAME}-extract-$$"
BACKEND_STAGING_DIR="${REMOTE_PARENT}/.${REMOTE_BASENAME}-deploy-$$"
BACKEND_BACKUP_DIR="${REMOTE_PARENT}/.${REMOTE_BASENAME}-backup-$$"
FRONTEND_STAGING_DIR="${FRONTEND_PARENT}/.${FRONTEND_BASENAME}-deploy-$$"
FRONTEND_BACKUP_DIR="${FRONTEND_PARENT}/.${FRONTEND_BASENAME}-backup-$$"

if [[ "$(id -u)" -eq 0 ]]; then
	SUDO=()
else
	SUDO=(sudo -n)
	if ! "${SUDO[@]}" true; then
		echo "错误：远程用户需要免密 sudo 权限。" >&2
		exit 1
	fi
fi

cleanup() {
	"${SUDO[@]}" rm -rf "${EXTRACT_DIR}" "${BACKEND_STAGING_DIR}" \
		"${FRONTEND_STAGING_DIR}"
	"${SUDO[@]}" rm -f "${ARCHIVE}"
}
trap cleanup EXIT

if ! "${SUDO[@]}" test -f "${REMOTE_DIR}/config.json"; then
	echo "错误：${REMOTE_DIR}/config.json 不存在，已取消部署。" >&2
	exit 1
fi

if ! "${SUDO[@]}" test -d "${REMOTE_FRONTEND_DIR}"; then
	echo "错误：前端目录 ${REMOTE_FRONTEND_DIR} 不存在，已取消部署。" >&2
	exit 1
fi

"${SUDO[@]}" mkdir -p "${EXTRACT_DIR}" "${BACKEND_STAGING_DIR}/docs/story" \
	"${FRONTEND_STAGING_DIR}"
"${SUDO[@]}" tar --no-same-owner --strip-components=1 \
	-xzf "${ARCHIVE}" -C "${EXTRACT_DIR}"

if ! "${SUDO[@]}" test -f "${EXTRACT_DIR}/novel" || \
	! "${SUDO[@]}" test -d "${EXTRACT_DIR}/docs/story/content" || \
	! "${SUDO[@]}" test -f "${EXTRACT_DIR}/frontend/dist/index.html" || \
	! "${SUDO[@]}" test -d "${EXTRACT_DIR}/frontend/dist/assets"; then
	echo "错误：发布包内容不完整，已取消部署。" >&2
	exit 1
fi

"${SUDO[@]}" cp "${EXTRACT_DIR}/novel" "${BACKEND_STAGING_DIR}/novel"
"${SUDO[@]}" cp -R "${EXTRACT_DIR}/docs/story/content" \
	"${BACKEND_STAGING_DIR}/docs/story/content"
"${SUDO[@]}" cp "${REMOTE_DIR}/config.json" "${BACKEND_STAGING_DIR}/config.json"
"${SUDO[@]}" chmod 755 "${BACKEND_STAGING_DIR}/novel"
"${SUDO[@]}" cp "${EXTRACT_DIR}/frontend/dist/index.html" \
	"${FRONTEND_STAGING_DIR}/index.html"
"${SUDO[@]}" cp -R "${EXTRACT_DIR}/frontend/dist/assets" \
	"${FRONTEND_STAGING_DIR}/assets"

"${SUDO[@]}" systemctl stop "${SERVICE_NAME}"

if ! "${SUDO[@]}" mv "${REMOTE_DIR}" "${BACKEND_BACKUP_DIR}"; then
	"${SUDO[@]}" systemctl start "${SERVICE_NAME}" || true
	echo "错误：无法备份后端目录，已取消部署。" >&2
	exit 1
fi

if ! "${SUDO[@]}" mv "${REMOTE_FRONTEND_DIR}" "${FRONTEND_BACKUP_DIR}"; then
	"${SUDO[@]}" mv "${BACKEND_BACKUP_DIR}" "${REMOTE_DIR}"
	"${SUDO[@]}" systemctl start "${SERVICE_NAME}" || true
	echo "错误：无法备份前端目录，已恢复旧版本。" >&2
	exit 1
fi

if ! "${SUDO[@]}" mv "${BACKEND_STAGING_DIR}" "${REMOTE_DIR}"; then
	"${SUDO[@]}" mv "${FRONTEND_BACKUP_DIR}" "${REMOTE_FRONTEND_DIR}"
	"${SUDO[@]}" mv "${BACKEND_BACKUP_DIR}" "${REMOTE_DIR}"
	"${SUDO[@]}" systemctl start "${SERVICE_NAME}" || true
	echo "错误：无法替换后端目录，已恢复旧版本。" >&2
	exit 1
fi

if ! "${SUDO[@]}" mv "${FRONTEND_STAGING_DIR}" "${REMOTE_FRONTEND_DIR}"; then
	"${SUDO[@]}" rm -rf "${REMOTE_DIR}"
	"${SUDO[@]}" mv "${FRONTEND_BACKUP_DIR}" "${REMOTE_FRONTEND_DIR}"
	"${SUDO[@]}" mv "${BACKEND_BACKUP_DIR}" "${REMOTE_DIR}"
	"${SUDO[@]}" systemctl start "${SERVICE_NAME}" || true
	echo "错误：无法替换前端目录，已恢复旧版本。" >&2
	exit 1
fi

if "${SUDO[@]}" systemctl start "${SERVICE_NAME}" && \
	"${SUDO[@]}" systemctl is-active --quiet "${SERVICE_NAME}"; then
	"${SUDO[@]}" rm -rf "${BACKEND_BACKUP_DIR}" "${FRONTEND_BACKUP_DIR}"
	"${SUDO[@]}" systemctl status "${SERVICE_NAME}" --no-pager --lines=5
else
	echo "错误：新版本启动失败，正在恢复旧版本。" >&2
	"${SUDO[@]}" systemctl stop "${SERVICE_NAME}" || true
	"${SUDO[@]}" rm -rf "${REMOTE_DIR}" "${REMOTE_FRONTEND_DIR}"
	"${SUDO[@]}" mv "${BACKEND_BACKUP_DIR}" "${REMOTE_DIR}"
	"${SUDO[@]}" mv "${FRONTEND_BACKUP_DIR}" "${REMOTE_FRONTEND_DIR}"
	"${SUDO[@]}" systemctl start "${SERVICE_NAME}" || true
	"${SUDO[@]}" journalctl -u "${SERVICE_NAME}" --no-pager -n 50 || true
	exit 1
fi
REMOTE_SCRIPT

echo "[4/4] 部署完成"
echo "后端：${SSH_HOST}:${REMOTE_DIR}"
echo "前端：${SSH_HOST}:${REMOTE_FRONTEND_DIR}"
