# iBoard 后端接口设计文档
## 架构设计
### 分层架构
系统采用经典的四层架构设计：
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
+---> 从Redis获取缓存的建筑物信息 (getCachedBuilding)
  |
  +---> 从Redis获取缓存的通知ID列表 (getCachedNoticeIDs)
  |
  +---> 从数据库获取现有的iSmart通知 (is_ismart_notice=true)
  |
  +---> 调用旧系统API获取通知列表
  |
  ---> 对比每一个旧系统的notice
  |
  +---> 创建现有通知MD5映射,现有的md5映射,比如[12,23],然后比对旧系统的md5[12.24],然后其中24 就是需要同步的notice,然后因为23 在旧系统中已经无了,就需要删除delete
  |
  +---> 计算处理通知的最佳并发数 (calculateWorkerCount)
  |
  |     +-------------+   通知1   +--------------+
  +---> | 并发处理通知 | --------> | 下载、MD5比对、处理 |
  |     +-------------+          +--------------+
  |           |         通知2          |
  |           +-----------------> ... 
  |           |         通知3          |
  |           +-----------------> ...
  |           |          ...          |
  |           +-----------------> ...
  |
  +---> 收集处理结果
  |     |
  |     +---> 已同步的通知 (hasSyncedCount)
  |     |
  |     +---> 新同步的通知 (successCount)
  |     |
  |     +---> 同步失败的通知 (failedNotices)
  |     |
  |     +---> 需要删除的通知 (deleteCount)
  |
  +---> [如果有变更通知] 处理需删除的通知 (processDeletedNotices)
   +---> 遍历现有通知
  |     |
  |     +---> [通知存在于旧系统] 保留(在md5映射中判断)
  |     |
  |     +---> [通知不存在于旧系统] 开始事务(根据md5映射判断)
  |            |
  |            +---> 解除通知与建筑物的绑定
  |            |
  |            +---> 检查通知是否还绑定到其他建筑物
  |            |     |
  |            |     +---> [有其他绑定] 保留通知
  |            |     |
  |            |     +---> [无其他绑定] 删除通知
  |            |            |
  |            |            +---> 检查文件是否还被其他通知引用
  |            |                  |
  |            |                  +---> [有其他引用] 保留文件
  |            |                  |
  |            |                  +---> [无其他引用] 删除文件
  |            |
  |            +---> 提交事务
  |            |
  |            +---> 增加删除计数
  |
  +---> 返回删除的通知数量
  |
  +---> 更新Redis缓存
  |     |
  |     +---> 更新通知ID列表 (setCachedNoticeIDs)
  |     |
  |     +---> 更新通知计数 (updateCachedNoticeCount)
  |
  +---> 返回同步结果统计