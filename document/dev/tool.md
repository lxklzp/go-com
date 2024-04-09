**gentool导出成struct**
```
# pgsql
gentool -db postgres -dsn "host=127.0.0.1 user=dolphinscheduler password=mypass dbname=resources port=29005 sslmode=disable" -onlyModel -tables="c_linkbundle"

# mysql
gentool -dsn "user:pwd@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local" -onlyModel -tables="c_linkbundle"
```

**编译xx.proto文件**
```
cd /internal/grpcs/proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative xx.proto
```



