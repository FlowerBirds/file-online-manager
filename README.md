# file-online-manager
文件在线管理工具，可查看、删除、上传服务器上的文件

## 程序结构
### 前端
 - 页面布局为left-center布局，left展示文件夹树，center展示选中的文件列表
 - 文件展示列表
 - 左侧展示文件夹树
 - 右侧展示当前选中的文件夹下的文件列表
 - 文件列表展示的每个文件上面显示删除、重命名操作、复制、解压等按钮
 - 左侧文件夹树上面有个工具栏，工具栏上面有新增文件夹、上传文件按钮

### 后端
 - 支持展示静态页面，类似与NGINX的功能
 - 支持API网关功能，增加权限校验，基于用户名和密码
 - 实现文件和文件夹的展示、删除、重命名、上传、复制等操作

### API设计
 - 文件删除：/api/manager/file/delete
 - 文件重命名：/api/manager/file/rename
 - 文件列表：/api/manager/file/list
 - 文件复制：/api/manager/file/copy
 - 文件加压：/api/manager/file/unzip
 - 文件上传：/api/manager/file/upload
 - 文件夹列表：/api/manager/folder/list
 - 文件夹删除：/api/manager/folder/delete
 - 文件夹重命名：/api/manager/folder/rename
 - 文件夹复制：/api/manager/folder/copy

### 架构设计
- 前后端分离
- 基于http basic auth的权限验证方式
- 支持自定义contextPath


## ChatGPT Prompt
### 后端
基于GO实现一个文件在线管理系统，该系统可以查看部署当前程序服务器的本地文件。系统后端基于http basic auth的权限验证方式，
所有请求必须通过用户民和密码登录成功后方可访问。本系统是前后端分离，系统可以处理静态网页，当前端访问
html时，系统能正确返回，html、png、js、css等静态文件位于static文件夹。系统实现文件和文件夹的展示、删除、重命名、上传、复制等操作，
涉及的API有如下几个：
- 文件删除：/api/manager/file/delete
- 文件重命名：/api/manager/file/rename
- 文件列表：/api/manager/file/list
- 文件复制：/api/manager/file/copy
- 文件夹列表：/api/manager/folder/list
- 文件夹删除：/api/manager/folder/delete
- 文件夹重命名：/api/manager/folder/rename
- 文件夹复制：/api/manager/folder/copy
- 文件上传：/api/manager/file/upload
----
生成main方法，并实现上述功能。

### 前端
代码编写一个html页面，采用vue + element-ui框架实现，页面布局采用left+center布局。left部分中采用
top+center布局，top里面实现一个工具栏，有一排按钮，高度20px，center部分中是一个树形结构，用来展示
文件夹树，树为异步加载，每点击一个文件夹，展开一层。工具栏中有上传按钮、文件夹重命名按钮、文件夹删除按钮、
文件夹复制按钮。left+center布局中的center部分，为一个文件列表展示，根左侧的文件夹树进行联动，点击文件夹树上的某一个
节点，文件列表中立刻展示对应文件夹下的文件列表，当双击文件列表中的文件夹时，文件列表展示对应文件夹下的文件列表，同时左侧
文件夹树选中对应文件夹。文件列表中可以对文件进行操作，包括删除、重命名、复制等。涉及操作的相关API如下：
- 文件删除：/api/manager/file/delete
- 文件重命名：/api/manager/file/rename
- 文件列表：/api/manager/file/list
- 文件复制：/api/manager/file/copy
- 文件夹列表：/api/manager/folder/list
- 文件夹删除：/api/manager/folder/delete
- 文件夹重命名：/api/manager/folder/rename
- 文件夹复制：/api/manager/folder/copy
- 文件上传：/api/manager/file/upload
------
实现以上功能。

### 部署YAML
基于镜像manage:latest，编写一个在k8s中部署manage服务的yaml文件，其中映射出的访问端口为8080，挂载的文件夹分别
是/app/apps、/app/file、/app/resource-home，挂载的目录均为本地路径，部署成功后，浏览器可访问其暴漏的8080端口访问系统并进行操作。


## 安全模式
当设置环境变量`MANAGE_SECURITY`值为true或者不设置（默认）时，则开启安全模式，该模式下登录凭证每月更新一次，为随机值，需从服务日志中获取。
```bash
export MANAGE_SECURITY=true
```

当设置MANAGE_SECURITY为false时，禁用自动更新，通过设置`MANAGE_USERNAME`和`MANAGE_PASSWORD`来指定登录凭证。
```bash
export MANAGE_SECURITY=false
export MANAGE_USERNAME=admin
export MANAGE_PASSWORD=1Fx98ksOa23GHapo0
```


## TODO list
- [ ] 支持大文件分片上传
- [x] 支持压缩文件解压（zip、tar）
- [ ] 支持文件夹压缩
- [ ] 支持文件下载
- [x] 支持设置安全模式，可以定时更新登录token
