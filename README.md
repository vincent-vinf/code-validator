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
* 沙箱包装实现
  * 需要cgroup，所以docker run --privileged
* 文件管理
* 多个测试样例



```bash
isolate --run --stderr-to-stdout -o /tmp/o -- /bin/echo "123" > /tmp/123

docker run -it --rm --privileged registry.cn-shanghai.aliyuncs.com/codev/js-executor:0.0.1 bash
```

概念

* sandbox：安全可控的执行用户操作
* pipeline：控制一系列sandbox中的操作，管理对外界文件的可见性
* validator：本质是对pipeline的封装，实现 初始化-运行-验证 流程，对语言进行抽象
* manager：对众多验证器进行管理，提供新建验证任务，查询验证任务等接口



 
