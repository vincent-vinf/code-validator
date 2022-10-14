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

### 问题
* 多文件支持？
* 包安装

### TODO
* 沙箱包装实现