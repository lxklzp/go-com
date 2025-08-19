#!/bin/sh

# host：在宿主机中编译，docker：在docker中编译，jenkins：用于jenkins
BUILD_ENV=$1
# 编译内核：amd64 arm64
BUILD_CORE=$2
# 项目版本号
PROJECT_VERSION=$3
# build：只执行编译，pack：只执行生成镜像和打包，all：执行全部步骤
BUILD_STEP=$4

# go版本号
GO_VERSION="1.24.3"
# yes：编译main_web，no：不编译main_web
BUILD_WEB="no"
# 项目根目录名称
PROJECT_NAME="auth-fast"
# 项目的git仓库ssh地址
PROJECT_GIT_URL="git@192.168.2.211:phoenix/auth-fast.git"
# jenkins目录
PROJECT_JENKINS_DIR="/var/data"

if [ $BUILD_ENV = "jenkins" ]
then
  chmod -R 777 /var/data/$PROJECT_NAME
  docker run -v $PROJECT_JENKINS_DIR/$PROJECT_NAME:/var/data/$PROJECT_NAME go$GO_VERSION.ubuntu-$BUILD_CORE /var/data/$PROJECT_NAME/data/$PROJECT_NAME/build.sh docker $BUILD_CORE $PROJECT_VERSION build
fi

if [ $BUILD_STEP = "all" ]
then
  # 初始化文件
  mkdir -p /var/data
  cd /var/data
  rm -rf $PROJECT_NAME
  git clone -b $PROJECT_VERSION $PROJECT_GIT_URL
  chmod -R 777 $PROJECT_NAME

  if [ $BUILD_ENV = "docker" ]
  then
    docker run -v /var/data/$PROJECT_NAME:/var/data/$PROJECT_NAME go$GO_VERSION.ubuntu-$BUILD_CORE /var/data/$PROJECT_NAME/data/$PROJECT_NAME/build.sh docker $BUILD_CORE $PROJECT_VERSION build
  fi
fi

if ( [ $BUILD_ENV = "host" ] && [ $BUILD_STEP = "all" ] ) || ( [ $BUILD_ENV = "docker" ] && [ $BUILD_STEP = "build" ] )
then

  # 步骤1 编译 build
  cd /var/data/$PROJECT_NAME

  mkdir -p runtime/$PROJECT_NAME-$PROJECT_VERSION/mysql
  cp -R /var/data/$PROJECT_NAME/data/$PROJECT_NAME/mysql/* runtime/$PROJECT_NAME-$PROJECT_VERSION/mysql

  mkdir -p runtime/$PROJECT_NAME-$PROJECT_VERSION/pgsql
  cp -R /var/data/$PROJECT_NAME/data/$PROJECT_NAME/pgsql/* runtime/$PROJECT_NAME-$PROJECT_VERSION/pgsql

  cp -R /var/data/$PROJECT_NAME/data/$PROJECT_NAME/release.md runtime/$PROJECT_NAME-$PROJECT_VERSION/release.md

  # 编译
  cd /var/data/$PROJECT_NAME
  go mod tidy
  go run build_app.go
  if [ $BUILD_WEB = "yes" ]
  then
    go run build_web.go
  fi

  rm -rf runtime/$PROJECT_NAME-$PROJECT_VERSION/$PROJECT_NAME
  mkdir -p runtime/$PROJECT_NAME-$PROJECT_VERSION/$PROJECT_NAME
  cp -R runtime/build_app/* runtime/$PROJECT_NAME-$PROJECT_VERSION/$PROJECT_NAME
  touch runtime/$PROJECT_NAME-$PROJECT_VERSION/$PROJECT_NAME/runtime/empty.txt

fi

# 步骤2 生成镜像和打包 pack
if [ $BUILD_STEP = "all" ]
then
  # docker镜像 main_app

  docker stop $(docker ps -a | grep "$PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE" | awk '{print $1 }')
  docker rm $(docker ps -a | grep "$PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE" | awk '{print $1 }')
  docker image rm $PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE

  cd /var/data/$PROJECT_NAME/runtime/$PROJECT_NAME-$PROJECT_VERSION
  chmod +x $PROJECT_NAME/main_app
  if [ $BUILD_WEB = "yes" ]
  then
    mv $PROJECT_NAME/main_web main_web
  fi
  docker build -f /var/data/$PROJECT_NAME/data/$PROJECT_NAME/dockerfile_app -t $PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE .
  if [ $BUILD_WEB = "yes" ]
  then
    mv main_web $PROJECT_NAME/main_web
  fi

  if [ $BUILD_WEB = "yes" ]
  then
    # docker镜像 main_web

    docker stop $(docker ps -a | grep "$PROJECT_NAME-web-$PROJECT_VERSION-$BUILD_CORE" | awk '{print $1 }')
    docker rm $(docker ps -a | grep "$PROJECT_NAME-web-$PROJECT_VERSION-$BUILD_CORE" | awk '{print $1 }')
    docker image rm $PROJECT_NAME-web-$PROJECT_VERSION-$BUILD_CORE

    cd /var/data/$PROJECT_NAME/runtime/$PROJECT_NAME-$PROJECT_VERSION
    chmod +x $PROJECT_NAME/main_web
    mv $PROJECT_NAME/main_app main_app
    docker build -f /var/data/$PROJECT_NAME/data/$PROJECT_NAME/dockerfile_web -t $PROJECT_NAME-web-$PROJECT_VERSION-$BUILD_CORE .
    mv main_app $PROJECT_NAME/main_app
  fi

  # 导出docker镜像
  if [ $BUILD_WEB = "yes" ]
  then
    docker save -o $PROJECT_NAME-$PROJECT_VERSION-$BUILD_CORE-images.tar $PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE:latest $PROJECT_NAME-web-$PROJECT_VERSION-$BUILD_CORE:latest
  else
    docker save -o $PROJECT_NAME-$PROJECT_VERSION-$BUILD_CORE-images.tar $PROJECT_NAME-app-$PROJECT_VERSION-$BUILD_CORE:latest
  fi

  # 生成压缩包

  cd /var/data/$PROJECT_NAME/runtime

  zip -q -r $PROJECT_NAME-$PROJECT_VERSION-$BUILD_CORE.zip $PROJECT_NAME-$PROJECT_VERSION
fi