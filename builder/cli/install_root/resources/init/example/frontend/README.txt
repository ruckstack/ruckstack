docker build -t local/backend .
docker run --rm -v "$(pwd)":/usr/share/app -p 8080:8080 -e POSTGRES_PASSWORD=${pwd} local/backend

docker run --rm -v "$(pwd)/src":/usr/share/nginx/html -p 8080:80 nginx:latest