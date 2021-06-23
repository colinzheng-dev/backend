from nginx:1.17-alpine

COPY services/image-proxy/nginx.conf /etc/nginx/nginx.conf
COPY services/image-proxy/default.conf.template /etc/nginx/conf.d/default.conf.template
COPY services/image-proxy/docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
