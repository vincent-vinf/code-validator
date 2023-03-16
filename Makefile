#TAG = test-$(shell git log -1 --format=%h)
TAG = 0.0.1
WORK_DIR = .
REGISTRY = registry.cn-shanghai.aliyuncs.com/codev

NAMESPACE=code-validator

all_image: dispatcher result user

build_push: all_image push

dispatcher:
	docker build --target dispatcher -t $(REGISTRY)/dispatcher:$(TAG) -f ./dockerfile/code-validator.dockerfile $(WORK_DIR)

result:
	docker build --target result -t $(REGISTRY)/result:$(TAG) -f ./dockerfile/code-validator.dockerfile $(WORK_DIR)

user:
	docker build --target user -t $(REGISTRY)/user:$(TAG) -f ./dockerfile/code-validator.dockerfile $(WORK_DIR)

js-actuator:
	docker build -t $(REGISTRY)/js-actuator:$(TAG) -f ./dockerfile/js-actuator.dockerfile $(WORK_DIR)

python-actuator:
	docker build -t $(REGISTRY)/python-actuator:$(TAG) -f ./dockerfile/python-actuator.dockerfile $(WORK_DIR)

push:
	docker push $(REGISTRY)/spike-access-service:$(TAG)

tar_chart:
	tar -zcvf spike-chart-$(TAG).tar.gz -C ./helm .

dependencies_install:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm install cv-minio bitnami/minio -n $(NAMESPACE) --create-namespace
	helm install -f chart/mysql/values.yaml cv-mysql bitnami/mysql -n $(NAMESPACE) --create-namespace
	helm install cv-rabbitmq --set auth.erlangCookie=secretcookie bitnami/rabbitmq -n $(NAMESPACE) --create-namespace
	helm install cv-redis --set architecture=standalone bitnami/redis -n $(NAMESPACE) --create-namespace

mc_install:
	curl https://dl.min.io/client/mc/release/linux-amd64/mc --create-dirs -o $HOME/bin/mc
	chmod +x "$HOME/bin/mc"
	mc --autocompletion
