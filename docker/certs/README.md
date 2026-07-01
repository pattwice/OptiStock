# Self-signed TLS certs for LAN HTTPS (optional).
# Generate before first prod deploy:
#
#   openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
#     -keyout docker/certs/privkey.pem \
#     -out docker/certs/fullchain.pem \
#     -subj "/CN=optistock.local"
#
# Mount path is configured in docker-compose.prod.yml and docker/nginx.conf.
