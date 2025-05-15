# iBoard 后端接口设计文档
## 架构设计
### 分层架构
系统采用经典的三层架构设计：
- Controller 层：处理 HTTP 请求响应
- Service 层：实现业务逻辑
- Model 层：数据模型定义
### 每层职责
#### Controller 层
- 处理 HTTP 请求和响应
- 参数校验和绑定
- 调用 Service 层处理业务逻辑
- 统一响应格式处理
- 错误处理和状态码管理
#### Service 层
- 实现核心业务逻辑
- 调用 Repository 层进行数据操作
- 数据组装和转换
- 事务管理
- 业务规则验证
#### Model 层
- 定义数据结构
- 定义表关系
- 字段验证规则
- 模型关联关系
## 模块划分
系统分为基础模块(base)和关系模块(relationship)两大类：

## SuperAdmin 功能模块
### 1. Super Admin Management(超级管理员账户管理)
  - add(email,password)
  - login(emailmpassword)
  - updateP(id,newPassword)
  - get(pageSize,pageNum)
  - getOne(id)
  - delete(ids[])
### 2. Building Admin Management(楼宇管理员账户管理)
  - add(email,password,status)
  - update(id,!password,!status)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 3. Building Management(楼宇信息管理)
  - add(name,ismartId,password,!remark)
  - delete(ids[])
  - update(name,!ismartId,!password,!remark)
  - get(pageSize,pageNum)
  - getOne(id)
### 4. Advertisement Management(广告管理)
  - add(title,!description,type,status,duration,startTime,endTime,isPublic,path)
  - update(id,!title,!description,!type,!status,!duration,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 5. Notice Management(通知管理)
  - add(title,!description,type,status,startTime,endTime,isPublic,path,fileTy)
  - addMany([{title,type,status,startTime,endTime,isPublic,path,duration}])
  - update(id,!title,!type,!status,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 6. File Management(文件管理)
  - getUploadParams(fileName)
  - uploadCallback(fileName,size)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 7. Building-Advertisement Relationship(楼宇广告关联管理)
  - bind(buildingId,advertisementId)(advertisementIds[],buildingIds[])
  - unbind(buildingId,advertisementId)(advertisementId,buildingIds[])
  - getBuildingAdvertisements(buildingId)
  - getAdvertisementBuildings(advertisementId)
### 8. Building-Notice Relationship(楼宇通知关联管理)
  - bind(buildingId,noticeId)(noticeIds[],buildingIds[])
  - unbind(buildingIds[],noticeId),
  - getBuildingNotices(buildingId)  
  - getNoticeBuildings(noticeId)
### 9. Building-Admin Relationship(楼宇管理员关联管理)
  - bind(buildingId,adminId)
  - unbind(buildingId,adminId)
  - getBuildingAdmins(buildingId)
  - getAdminBuildings(adminId)


## BuildingAdmin 功能模块
## BuildingAdmin login 功能为下面的接口添加token
### 1. Advertisement Management(广告管理)
  - add(title,!description,type,status,duration,startTime,endTime,isPublic,path)
  - update(id,!title,!description,!type,!status,!duration,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 2. Notice Management(通知管理)
  - add(title,!description,type,status,startTime,endTime,isPublic,path,fileTy)
  - addMany([{title,type,status,startTime,endTime,isPublic,path,duration}])
  - update(id,!title,!type,!status,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 3. File Management(文件管理)
  - getUploadParams(fileName)
  - uploadCallback(fileName,size)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)

## building 功能模块
  - login(ismartId,password)
  - get_advertisements_building
  - get_notices_building 


<!-- services:
  backend:
    build: .  # 使用本地Dockerfile构建
    container_name: iboard_http_service
    restart: always
    ports:
      - "10031:10031"
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=1090119your
      - DB_NAME=iboard_db
      - DB_TIMEZONE=Asia/Shanghai

      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=
      - REDIS_DB=0

      # 根据实际情况配置邮件服务
      - SMTP_ADDR=smtp.example.com
      - SMTP_PORT=587
      - SMTP_USER=your-email@example.com
      - SMTP_PASS=your-password
    volumes:
      - ./logs:/app/logs  # 将日志目录挂载到主机
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - iboard-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "365"

  mysql:
    image: mysql:8.0
    container_name: iboard_mysql
    restart: always
    ports:
      - "3308:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=1090119your
      - MYSQL_DATABASE=iboard_db
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password
    networks:
      - iboard-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p1090119your"]
      interval: 10s
      timeout: 5s
      retries: 5
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "3"

  redis:
    image: redis:7.4.1
    container_name: iboard_redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - iboard-network
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "365"

networks:
  iboard-network:
    driver: bridge

volumes:
  mysql_data:
  redis_data: -->