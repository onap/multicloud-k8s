build:
	$(MAKE) -C monitor build
	$(MAKE) -C k8splugin build
	$(MAKE) -C orchestrator build
	$(MAKE) -C ncm build
	$(MAKE) -C clm build
	$(MAKE) -C dcm build
	$(MAKE) -C rsync build
	$(MAKE) -C ovnaction build

deploy:
	$(MAKE) -C monitor deploy
	$(MAKE) -C k8splugin deploy
	$(MAKE) -C orchestrator deploy
	$(MAKE) -C ncm deploy
	$(MAKE) -C clm deploy
	$(MAKE) -C dcm deploy
	$(MAKE) -C rsync deploy
	$(MAKE) -C ovnaction deploy

all:
	$(MAKE) -C monitor all
	$(MAKE) -C k8splugin all
	$(MAKE) -C orchestrator all
	$(MAKE) -C ncm all
	$(MAKE) -C clm all
	$(MAKE) -C dcm all
	$(MAKE) -C rsync all
	$(MAKE) -C ovnaction all

clean:
	$(MAKE) -C monitor clean
	$(MAKE) -C k8splugin clean
	$(MAKE) -C orchestrator clean
	$(MAKE) -C ncm clean
	$(MAKE) -C clm clean
	$(MAKE) -C dcm clean
	$(MAKE) -C rsync clean
	$(MAKE) -C ovnaction clean
