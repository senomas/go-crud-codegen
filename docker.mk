# Guard to avoid double-loading
ifndef __COMMON_MK__
__COMMON_MK__ := 1

DC := docker compose --env-file .env --env-file .secret.env --env-file .local.env  -f compose.yml

define prop-get
docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work python:3.12-slim \
		python3 build-util/prop.py get build.properties $(1)
endef
	
define prop-set
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work python:3.12-slim \
		python3 build-util/prop.py set build.properties $(1) $(2)
endef
	
define env-set
	docker run --rm --user $$(id -u):$$(id -g) -v "$$(pwd)":/work -w /work python:3.12-slim \
		python3 build-util/prop.py set .env $(1) $(2)
endef
	
define bump-ver
	docker run --rm \
		-v "$$(pwd)":/work \
		-w /work \
		python:3.12-slim \
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
	for svc in $(1); do \
		echo "⏳ Waiting for $$svc to be healthy (timeout 120s)..."; \
		SECS=0; \
		while [ "$$SECS" -lt 120 ]; do \
			STATUS=$$($(DC) ps -q $$svc | xargs -r docker inspect -f '{{.State.Health.Status}}'); \
			if [ "$$STATUS" = "healthy" ]; then \
				echo "✅ $$svc is healthy"; \
				break; \
			fi; \
			sleep 5; \
			SECS=$$((SECS+5)); \
			echo "…still waiting for $$svc ($$SECS/120s)"; \
		done; \
		if [ "$$SECS" -ge 120 ]; then \
			echo "❌ $$svc failed to become healthy within 120s"; \
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
	@echo -e "Processing $(2) - $(1)"; \
	app_ver=$$(git log -1 --date=format-local:%Y%m%d-%H%M%S --format=%cd -- $(1))-$$(git log -1 --format=%h -- $(1)); \
	echo "Computed version: $$app_ver"; \
	prop_ver=$$($(call prop-get,$(2)_VER)); \
	if [ "$$app_ver" != "$$prop_hash" ]; then \
		echo "$(2): Changes detected, updating codegen and rebuilding..."; \
		echo "Bumping version $$prop_ver -> $$app_ver"; \
	fi; \
	( \
		docker image ls --format "{{.Repository}}:{{.Tag}}" | grep -q "^${DOCKER_REGISTRY}/$3:$$app_ver$$" && \
		echo "Docker image ${DOCKER_REGISTRY}/$3:$$app_ver already exists locally." || \
		docker pull ${DOCKER_REGISTRY}/$3:$$app_ver || \
		( \
			dirty=$$(git status --porcelain -- $(1)); \
			last_msg=$$(git log -1 --pretty=%s -- $1 2>/dev/null || true); \:
			if [ -n "$$dirty" ]; then \
			  if printf "%s" "$$last_msg" | grep -qiE '^dev$$'; then \
			    echo "Working tree $1 dirty and last commit is 'dev' — amending changes..."; \
			    git add -A; \
			    git commit --amend --no-edit; \
			  else \
					echo "$(2): ERROR: Working tree is dirty, but last commit message is not exactly 'dev'."; \
			    echo "       Please commit/clean manually."; \
			    exit 1; \
			  fi; \
			fi; \
			ARGS=$$(cat .build.env | grep -v '^#' | xargs -d '\n' -I{} echo --build-arg {}) ; \
			docker build --build-arg DOCKER_REGISTRY=${DOCKER_REGISTRY} $$ARGS \
				-t ${DOCKER_REGISTRY}/$3:$$app_ver $1 && \
			docker push ${DOCKER_REGISTRY}/$3:$$app_ver \
		) \
	); \
	$(call prop-set,$2_VER,$$app_ver); \
	$(call env-set,$2_VER,$$app_ver); \
	echo "✅ Build complete, new version is $$app_ver"
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
