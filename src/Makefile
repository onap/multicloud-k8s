build:
	$(MAKE) -C monitor build
	$(MAKE) -C k8splugin build
	$(MAKE) -C orchestrator build

deploy:
	$(MAKE) -C monitor deploy
	$(MAKE) -C k8splugin deploy
	$(MAKE) -C orchestrator deploy

all:
	$(MAKE) -C monitor all
	$(MAKE) -C k8splugin all
	$(MAKE) -C orchestrator all

clean:
	$(MAKE) -C monitor clean
	$(MAKE) -C k8splugin clean
	$(MAKE) -C orchestrator clean
