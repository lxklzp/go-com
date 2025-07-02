# 初始化文件
cd /var/data/go-com

rm -rf runtime/go-com-v1.0.0/mysql
mkdir -p runtime/go-com-v1.0.0/mysql
cp -R /var/data/go-com/data/go-com/mysql/* runtime/go-com-v1.0.0/mysql

rm -rf runtime/go-com-v1.0.0/pgsql
mkdir -p runtime/go-com-v1.0.0/pgsql
cp -R /var/data/go-com/data/go-com/pgsql/* runtime/go-com-v1.0.0/pgsql

# 编译
cd /var/data/go-com
go run build_app.go
go run build_web.go

rm -rf runtime/go-com-v1.0.0/go-com
mkdir -p runtime/go-com-v1.0.0/go-com
cp -R runtime/build_app/* runtime/go-com-v1.0.0/go-com
touch runtime/go-com-v1.0.0/go-com/runtime/empty.txt

# docker镜像 main_app

docker stop $(docker ps -a | grep "go-com-app-v1.0.0" | awk '{print $1 }')
docker rm $(docker ps -a | grep "go-com-app-v1.0.0" | awk '{print $1 }')
docker image rm go-com-app-v1.0.0

cd /var/data/go-com/runtime/go-com-v1.0.0
chmod +x go-com/main_app
mv go-com/main_web main_web
docker build -f /var/data/go-com/data/go-com/dockerfile_app -t go-com-app-v1.0.0 .
mv main_web go-com/main_web

# docker镜像 main_web

docker stop $(docker ps -a | grep "go-com-web-v1.0.0" | awk '{print $1 }')
docker rm $(docker ps -a | grep "go-com-web-v1.0.0" | awk '{print $1 }')
docker image rm go-com-web-v1.0.0

cd /var/data/go-com/runtime/go-com-v1.0.0
chmod +x go-com/main_web
mv go-com/main_app main_app
docker build -f /var/data/go-com/data/go-com/dockerfile_web -t go-com-web-v1.0.0 .
mv main_app go-com/main_app

# 导出docker镜像

docker save -o go-com-v1.0.0-images.tar go-com-app-v1.0.0:latest go-com-web-v1.0.0:latest