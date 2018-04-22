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
		 $(HOOK)

end-simnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.hub` deploy ended", \
		"username": "$(USER)"}' \
		 $(HOOK)

start-testnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.hub` deploy started...", \
		"username": "$(USER)"}' \
		 $(HOOK)

end-testnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.hub` deploy ended", \
		"username": "$(USER)"}' \
		 $(HOOK)

# # # # # # # # # #
# Docker machine  #
# # # # # # # # # #

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
simnet-build-compose:
		@$(call print, "Activating simnet.connector.bitlum.io machine && building...")

		cp ./docker/hub/configs/simnet/hub.conf ./docker/hub/hub.conf
		cp ./docker/hub/configs/simnet/lnd-tls.cert  ./docker/hub/lnd-tls.cert

		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml up --build -d

		rm ./docker/hub/hub.conf
		rm ./docker/hub/lnd-tls.cert

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
testnet-build-compose:
		@$(call print, "Activating testnet.connector.bitlum.io machine && building...")

		cp ./docker/hub/configs/testnet/hub.conf ./docker/hub/hub.conf
		cp ./docker/hub/configs/testnet/lnd-tls.cert  ./docker/hub/lnd-tls.cert

		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml up --build -d

		rm ./docker/hub/hub.conf
		rm ./docker/hub/lnd-tls.cert

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
simnet-logs:
		@$(call print, "Activating simnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml logs --tail=1000 -f

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
testnet-logs:
		@$(call print, "Activating testnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml logs --tail=1000 -f

# # # # # # # # #
# Golang build  #
# # # # # # # # #

clean:
		@$(call print, "Removing build hub binaries...")
		rm ./docker/hub/hub

build:
		@$(call print, "Building hub...")
		GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/hub/hub

ifeq ($(HOOK),)
simnet-deploy:
		@$(call print, "You forgot specify hook!")
else
simnet-deploy: \
		start-simnet-notification \
		build \
		simnet-build-compose \
		clean \
		end-simnet-notification
endif

ifeq ($(HOOK),)
testnet-deploy:
		@$(call print, "You forgot specify hook!")
else
testnet-deploy: \
		start-testnet-notification \
		build \
		testnet-build-compose \
		clean \
		end-testnet-notification
endif

.PHONY: \
		simnet-deploy \
		testnet-deploy \
		simnet-logs \
		testnet-logs \
		build \
		clean
