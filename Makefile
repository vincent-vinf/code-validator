#TAG = test-$(shell git log -1 --format=%h)
TAG = 0.0.1
WORK_DIR = .
REGISTRY = registry.cn-shanghai.aliyuncs.com/codev

NAMESPACE=code-validator

all_image: access_image

build_push: all_image push

access_image:
	docker build --target access -t $(REGISTRY)/spike-access-service:$(TAG) $(WORK_DIR)

js-executor:
	docker build -t $(REGISTRY)/js-executor:$(TAG) -f ./dockerfile/js-executor.dockerfile $(WORK_DIR)

push:
	docker push $(REGISTRY)/spike-access-service:$(TAG)

tar_chart:
	tar -zcvf spike-chart-$(TAG).tar.gz -C ./helm .

echo:
	echo $(TAG)

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
