REGISTRY?=integratedcloudnative
TAG?=latest
IMAGES=iavf-driver-installer qat-driver-installer

.PHONY: all $(IMAGES)

all: $(IMAGES)

$(IMAGES):
	docker build -t $(REGISTRY)/$@:$(TAG) -f Dockerfile.$@ .
