.ONESHELL:

GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

USER := $(shell whoami)


# # # # #
# Slack #
# # # # #

start-simnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.hub` deploy started...", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

end-simnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.hub` deploy ended", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

start-testnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.hub` deploy started...", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

end-testnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.hub` deploy ended", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

start-mainnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`mainnet.hub` deploy started...", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

end-mainnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`mainnet.hub` deploy ended", \
		"username": "$(USER)"}' \
		 $(SLACK_HOOK)

# # # # # # # # # #
# Docker machine  #
# # # # # # # # # #

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
simnet-build-compose:
		@$(call print, "Activating simnet.connector.bitlum.io machine && building...")

		eval `docker-machine env simnet.connector.bitlum.io` && \
		cd ./docker/simnet/ && \
		docker-compose up --build -d

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
testnet-build-compose:
		@$(call print, "Activating testnet.connector.bitlum.io machine && building...")

		eval `docker-machine env testnet.connector.bitlum.io` && \
		cd ./docker/testnet/ && \
		docker-compose up --build -d
		
# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
mainnet-build-compose:
		@$(call print, "Activating mainnet.connector.bitlum machine && building...")

		eval `docker-machine env mainnet.connector.bitlum.io` && \
		cd ./docker/mainnet/ && \
		docker-compose up --build -d

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
simnet-logs:
		@$(call print, "Activating simnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/simnet/docker-compose.yml logs --tail=1000 -f

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
testnet-logs:
		@$(call print, "Activating testnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker-compose -f ./docker/testnet/docker-compose.yml logs --tail=1000 -f

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
mainnet-logs:
		@$(call print, "Activating mainnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env mainnet.connector.bitlum.io` && \
		docker-compose -f ./docker/mainnet/docker-compose.yml logs --tail=1000 -f

# # # # # # # # #
# Golang build  #
# # # # # # # # #



CCPATH := "/usr/local/gcc-4.8.1-for-linux64/bin/x86_64-pc-linux-gcc"
GOBUILD := GOOS=linux GOARCH=amd64 CC=$(CCPATH) CGO_ENABLED=1 go build -v -i -o

simnet-clean:
		@$(call print, "Removing build simnet hub binaries...")
		rm -rf ./docker/simnet/hub/bin/

testnet-clean:
		@$(call print, "Removing build testnet hub binaries...")
		rm -rf ./docker/testnet/hub/bin/

simnet-build:
		@$(call print, "Building simnet hub...")
		$(GOBUILD) ./docker/simnet/hub/bin/hub  .
		$(GOBUILD) ./docker/simnet/hub/bin/hubcli ./cmd/hubcli




testnet-build:
		@$(call print, "Building testnet hub...")
		$(GOBUILD) ./docker/testnet/hub/bin/hub .
		$(GOBUILD) ./docker/testnet/hub/bin/hubcli ./cmd/hubcli



ifeq ($(SLACK_HOOK),)
simnet-deploy:
		@$(call print, "You forgot specify SLACK_HOOK!")
else
simnet-deploy: \
		start-simnet-notification \
		simnet-build \
		simnet-build-compose \
		simnet-clean \
		end-simnet-notification
endif

ifeq ($(SLACK_HOOK),)
testnet-deploy:
		@$(call print, "You forgot specify SLACK_HOOK!")
else
testnet-deploy: \
		start-testnet-notification \
		testnet-build \
		testnet-build-compose \
		testnet-clean \
		end-testnet-notification
endif

ifeq ($(SLACK_HOOK),)
mainnet-deploy:
		@$(call print, "You forgot specify SLACK_HOOK!")
else
mainnet-deploy: \
		start-mainnet-notification \
		mainnet-build-compose \
		end-mainnet-notification
endif

.PHONY: simnet-deploy \
		simnet-logs \
		testnet-deploy \
		testnet-logs \
		mainnet-deploy \
		build \
		clean
