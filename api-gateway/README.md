mkdir -p api-gateway/nginx/certs \
touch api-gateway/nginx/certs/key.pem \
touch api-gateway/nginx/certs/cert.pem \
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
-keyout api-gateway/nginx/certs/key.pem \
-out api-gateway/nginx/certs/cert.pem \
-subj "/C=/ST=/L=/O=/OU=/CN="