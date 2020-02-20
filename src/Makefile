build:
	$(MAKE) -C monitor build
	$(MAKE) -C k8splugin build
	$(MAKE) -C orchestrator build
	$(MAKE) -C hpacontroller build

deploy:
	$(MAKE) -C monitor deploy
	$(MAKE) -C k8splugin deploy
	$(MAKE) -C orchestrator deploy
	$(MAKE) -C hpacontroller deploy

all:
	$(MAKE) -C monitor all
	$(MAKE) -C k8splugin all
	$(MAKE) -C orchestrator all
	$(MAKE) -C hpacontroller all

clean:
	$(MAKE) -C monitor clean
	$(MAKE) -C k8splugin clean
	$(MAKE) -C orchestrator clean
	$(MAKE) -C hpacontroller clean
