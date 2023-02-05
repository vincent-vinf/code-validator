### Doc
* minio: https://min.io/docs/minio/kubernetes/upstream/index.html

### 实现
* rlimit
* isolate
  * http://www.ucw.cz/moe/isolate.1.html
  * https://github.com/judge0/judge0/blob/5b70e8a0ca480fd77f77136c287535e8e69bc5a7/app/jobs/isolate_job.rb

### 编译

#### isolate

1. 下载
   * curl -L -o isolate.zip https://github.com/ioi/isolate/archive/refs/heads/master.zip
   * git clone https://github.com/ioi/isolate.git
2. 安装依赖
   * yum install -y libcap-devel
   * apt update && apt install -y libcap-dev
3. make install
4. 清理
   * rm -rf /var/lib/apt/lists/*

### 问题
* 多文件支持？
* 包安装

### TODO
- [x] 沙箱包装实现
- [x] 文件管理
- [x] 多个测试样例
- [ ] 编译支持
- [ ] 解压打包的代码
  - [ ] zip

- [ ] 自定义 初始化和验证步骤
- [ ] 自动扩缩容
- [ ] 多语言
  - [x] JavaScript
  - [x] Python
  - [ ] C++


```bash
docker run -it --rm --privileged registry.cn-shanghai.aliyuncs.com/codev/js-executor:0.0.1 bash
```

概念

* sandbox：安全可控的执行用户操作
* pipeline：控制一系列sandbox中的操作，管理对外界文件的可见性
* performer：本质是对pipeline的封装，实现 初始化-运行-验证 流程，对语言进行抽象，循环处理测试用例
* manager：对众多performer进行管理，提供新建验证任务，查询验证任务等接口



### 需求

1. 编排了一个js语言的作业验证流程
   * 自定义
   * 使用提供的模板
2. 上传必要的文件（后台会解压并解析文件，排除基础的格式错误，上传的文件被存储到oss）
   * n个测试样例
     * 输入
     * 输出
   * m个待评测的程序源码
3. 点击运行，batch消息被存储到数据库
   1. 发送m条任务消息到消息队列
   2. worker接受到消息，开始运行任务
   3. 收集的任务结果会被存储到数据库
   4. 任务日志会被存储到oss
4. 支持的查看操作
   1. 查看batch的统计图表
   2. 查看batch的验证任务列表，支持一些排序和筛选
   3. 查看batch的某个验证任务的结果，包括耗时、内存占用、测试样例通过情况，支持在线查看代码文件
5. 重试机制，如果是内部错误，则尝试重新运行task

> 支持其他人将测试文件上传到这个batch吗？

一个缓存库，从minio到本地

使用rabbitmq[路由模式](https://www.rabbitmq.com/tutorials/tutorial-four-go.html)，管理不同语言实例

多语言支持

模板支持，处于配置文件

传统校验*测试样例数量

* 可选init
* 可选解压
* 执行代码
* 验证结果

额外校验

* 可选init
* 可选解压
* 验证



测试用端口号

* dispatcher：8001
* user：8002
* result：8003

概念关系

* batch：验证任务，包括多个验证项
* verification：验证项
* task：对于一个验证任务创建的任务实例，包含一份用户输入的代码文件，并引用了batch
* subtask：对验证项的实例化，归属与task，引用了verification。是执行器执行的最小单元，保存了执行状态和执行结果

oss路径规则

* verification包含的文件



已完成内容

* 中间件对接（100%）mysql、minio、rabbitmq
* 各个微服务（40%）完成用户服务，正在编写核心的分发和执行服务，执行服务基本完成
* 前端（5%）完成框架选择和基本的代码执行demo UI
* 整体k8s部署（0%）

总体进度估计为（30%）

下一个阶段目标

* 完成执行和分发微服务（预计2月17开学前完成）
* UI（开学后开始）

当前问题：验证任务的执行结构需要优化一下，使得代码验证和自定义验证的结构统一





文件解压缩问题

