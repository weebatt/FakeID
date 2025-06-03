sudo mkdir -p api-gateway/nginx/certs \
sudo touch api-gateway/nginx/certs/key.pem \
sudo touch api-gateway/nginx/certs/cert.pem \
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
-keyout api-gateway/nginx/certs/key.pem \
-out api-gateway/nginx/certs/cert.pem \
-subj "/C=/ST=/L=/O=/OU=/CN="