#!/usr/bin/env sh
set -eu

envsubst '${SERVER_NAME} ${IMAGE_BUCKET}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

exec "$@"
