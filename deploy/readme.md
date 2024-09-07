# 部署说明

## 全量部署

将`build`目录移动到你的执行目录。 配置`config/config.yaml`

### 依赖安装

参考`depend`目录

### 配置目录权限
```
chmod -R 777 /你的执行目录/runtime
```

## 增量配置

替换可执行文件：`main_app`、`main_web`，根据开发提供的信息修改`config/config.yaml`

对比服务器线上版本`config`目录和你要更新的版本的`build/config`目录下的`config.yaml`，对相关配置项进行新增、替换、删除操作。

## 启动

### 启动应用服务器
```
/你的执行目录/main_app -id=1 -config=config.yaml文件完整路径
```

### supervisor配置示例
```
[program:auth-fast-app]
command=/var/data/auth-fast/main_app -id=1 -config=/var/data/auth-fast/config/config.yaml
directory=/var/data/auth-fast                                                                                                                                                                                      
autostart=true
autorestart=true
stderr_logfile=NONE
stdout_logfile=NONE
startsecs=30
startretries=60
```

## 编译
包含confluent-kafka-go的静态编译：

安装musl：
```
wget https://musl.libc.org/releases/musl-latest.tar.gz
tar zxvf musl-latest.tar.gz
cd musl-*
./configure --prefix=/usr/local/musl
make && sudo make install
/usr/local/musl/bin/musl-gcc -v
```

进入项目目录，编译：
```
CC=/usr/local/musl/bin/musl-gcc go build --ldflags '-linkmode external -extldflags "-static"' -tags musl -o /var/data/your-app/main_app /var/data/your-app/main_app.go
```