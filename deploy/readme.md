# 部署说明

## 首次配置

将`build`目录移动到你的执行目录。 配置`config/config.yaml`

### 依赖安装
参考`depend`目录

### 配置目录权限
```
chmod -R 777 /你的执行目录/runtime
```

## 增量配置
替换可执行文件：`main_app`、`main_web`，根据开发提供的信息修改`config/config.yaml`

## 启动

### 启动应用服务器
```
/你的执行目录/main_app -id=1
```

### 启动web服务器
```
/你的执行目录/main_web -id=1001
```

### 定时任务
```
crontab -e

* * * * * /你的执行目录/cron_redis2db
```

## 关闭
```
kill pid
```