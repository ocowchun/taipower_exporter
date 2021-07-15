VERSION = 0.0.1
TAG = $(VERSION)
PREFIX = ocowchun/taipower_exporter

.PHONY: container
container:
	docker build  -t $(PREFIX):$(TAG) .

.PHONY: push
push: container
	docker push $(PREFIX):$(TAG)