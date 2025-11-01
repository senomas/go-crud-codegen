# Guard to avoid double-loading
ifndef __COMMON_MK__
__COMMON_MK__ := 1

DC := docker compose --env-file .env --env-file .secret.env --env-file .local.env  -f compose.yml

define prop-get
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/prop.py get build.properties $(1)
endef

define prop-set
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/prop.py set build.properties $(1) $(2)
endef

define env-set
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/prop.py set .env $(1) $(2)
endef

define build-args
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/prop.py build-args build.properties
endef

define envs
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/prop.py envs build.properties
endef

define clean-images
	docker images "$(1)" \
	  --format '{{.Repository}}:{{.Tag}} {{.CreatedAt}}' \
		| docker run --rm -i --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work \
			${DOCKER_REGISTRY}/python:3.12-slim \
			python3 build-util/docker-keep-last.py 1 \
		| xargs -r -n1 sh -c 'docker rmi "$0" >/dev/null 2>&1 || true'
endef

define bump-ver
	docker run --rm \
		-v "$$(pwd)":/work \
		-w /work \
		${DOCKER_REGISTRY}/python:3.12-slim \
		python3 build-util/bump_ver.py $(1)
endef
	
define docker-pull
	@echo -e "Processing $(2) - $(1)" && \
	( \
		[ "$(REPULL)" == "" ] && docker image ls --format "{{.Repository}}:{{.Tag}}" | grep -q "^${DOCKER_REGISTRY}/$(2)$$" && \
			echo "Docker image $(2) already exists locally." || \
			( \
				docker pull $(1) && \
				docker tag $(1) ${DOCKER_REGISTRY}/$(2) && \
				docker push ${DOCKER_REGISTRY}/$(2) \
			) \
	)
endef

define start-service
	for svc in $(1); do \
		echo -e "\n\nStarting $$svc"; \
		$(DC) up -d --remove-orphans $$svc || ( $(DC) logs $$svc ; exit 1 ); \
	done; \
	timeout=${HEALTHY_TIMEOUT}; \
	if [ "$$timeout" == "" ]; then \
		timeout=300; \
	fi; \
	for svc in $(1); do \
		echo -ne "\r ⏳Waiting for $$svc to be healthy (0/$${timeout}s)..."; \
		SECS=0; \
		while [ "$$SECS" -lt $${timeout} ]; do \
			STATUS=$$($(DC) ps -q $$svc | xargs -r docker inspect -f '{{.State.Health.Status}}'); \
			if [ "$$STATUS" = "healthy" ]; then \
				echo -e "\r \e[32m\xE2\x9C\x94\e[0m Container $$svc is \e[32mhealthy\e[0m                                                        "; \
				break; \
			fi; \
			sleep 5; \
			SECS=$$((SECS+5)); \
			echo -ne "\r ⏳Waiting for $$svc to be healthy ($$SECS/$${timeout}s)..."; \
		done; \
		if [ "$$SECS" -ge $${timeout} ]; then \
			echo -e "\r ❌Container $$svc failed to become healthy within $${timeout}s        "; \
			$(DC) logs $$svc; \
			exit 1; \
		fi; \
	done
endef

define wait-healthy
	@docker compose up -d $(1)
	@echo -e "\033[48;5;202;38;5;15mWaiting $(1) to be healthy...\033[0m"
	@until [ $$(docker compose ps -q $(1) \
		| xargs docker inspect -f '{{.State.Health.Status}}') = "healthy" ]; do \
		sleep 1; \
	done
endef

define docker-build
	@echo "Checking git status for $(1)..."; \
	dirty=$$(git status --porcelain -- $(1)); \
	if [ -n "$$dirty" ]; then \
    echo "ERROR: Dirty, please commit/clean manually."; \
		git status --porcelain -- $(1); \
    exit 1; \
  fi; \
	app_ver=$$(git log -1 --date=format-local:%Y%m%d-%H%M%S --format=%cd -- $(1))-$$(git log -1 --format=%h -- $(1)); \
	prop_ver=$$($(call prop-get,$(2)_VER)); \
	if [ "$$app_ver" != "$$prop_ver" ]; then \
		echo "$(2): Changes detected, updating codegen and rebuilding..."; \
		echo "Bumping version [$$prop_ver] -> [$$app_ver]"; \
	fi; \
	( \
		docker image ls --format "{{.Repository}}:{{.Tag}}" | grep -q "^${DOCKER_REGISTRY}/$3:$$app_ver$$" && \
		echo "Docker image ${DOCKER_REGISTRY}/$3:$$app_ver already exists locally." || \
		docker pull ${DOCKER_REGISTRY}/$3:$$app_ver 2>/dev/null || \
		( \
			echo Building docker image ${DOCKER_REGISTRY}/$3:$$app_ver ...; \
			bargs=$$($(call build-args)); \
			if [ "$(5)" == "" ]; then \
				docker build --progress=plain \
					$$bargs -t $3:$$app_ver $1; \
			else \
				docker build -f $(4) --progress=plain \
					$$bargs -t $3:$$app_ver $5; \
			fi; \
			if [ "${DOCKER_PUSH}" != "" ]; then \
				docker tag $3:$$app_ver ${DOCKER_PUSH}/$3:$$app_ver; \
	      docker push ${DOCKER_PUSH}/$3:$$app_ver; \
				echo -e "\e[32m\xE2\x9C\x94\e[0m  Build complete, new version is ${DOCKER_PUSH}/$3:$$app_ver"
			else \
				docker tag $3:$$app_ver ${DOCKER_REGISTRY}/$3:$$app_ver; \
				echo -e "\e[32m\xE2\x9C\x94\e[0m  Build complete, new version is ${DOCKER_REGISTRY}/$3:$$app_ver"
			fi; \
		) \
	); \
	if [ "$$app_ver" != "$$prop_ver" ]; then \
    echo "Updating version in build.properties: $(2)_VER = $$app_ver"; \
		$(call prop-set,$2_VER,$$app_ver); \
	fi
endef

define docker-upgrade
	@ver=$$(. .env; printf '%s' "$$UTIL_VER"); \
	docker run --rm \
	  -u $(UID):$(GID) \
	  --group-add $(DOCKER_GID) \
	  -e HOME=/tmp \
	  -e DOCKER_HOST=unix:///var/run/docker.sock \
		--env-file .env \
		--env-file .local.env \
	  -v "$(DOCKER_SOCK)":/var/run/docker.sock \
		-v "$(PWD)":/w \
		-w /w ${DOCKER_REGISTRY}/test-suite-util:${UTIL_VER}\
	  python3 util/apt-upgrade.py -f $(1)
endef

endif  # __COMMON_MK__
