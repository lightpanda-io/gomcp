# Variables
# ---------
OP_VAULT := "mcp"

# Utilies
# ---------

# increment the parameter.
incr = $(shell echo $$(($(1)+1)))

# extract a list of fields secrets values from a 1password item
# usage: $(call op_secrets,my_vault,my_item,my_fields)
op_secrets = $(shell op item get --reveal --vault=$(1) --fields=$(shell echo $(3)| sed "s/ /,/g") -- "$(2)")
# extract one secret from secrets returned by op_secrets (space separated)
# usage: $(call op_secret,my_secrets,1)
op_secret = $(shell echo $(1) |cut -d, -f$(2))
# return KEY=VALUE envs list.
op_env = $(eval secrets=$(2)) \
		 $(eval i=1) \
		 $(foreach k,$(1), \
			$(k)="$(shell echo $(secrets)|cut -d',' -f$(i))" \
			$(eval i=$(call incr,($i))))

cdp_keys = MCP_CDP
cdp_fields = password
cdp_secrets = $(call op_secrets,$(OP_VAULT),CDP ADDRESS,$(cdp_fields))
cdp_env = $(call op_env,$(cdp_keys),$(cdp_secrets))

# Infos
# -----
.PHONY: help

## Display this help screen
help:
	@printf "\e[36m%-35s %s\e[0m\n" "Command" "Usage"
	@sed -n -e '/^## /{'\
		-e 's/## //g;'\
		-e 'h;'\
		-e 'n;'\
		-e 's/:.*//g;'\
		-e 'G;'\
		-e 's/\n/ /g;'\
		-e 'p;}' Makefile | awk '{printf "\033[33m%-35s\033[0m%s\n", $$1, substr($$0,length($$1)+1)}'

.PHONY: check-op
check-op:
	@op vault get $(OP_VAULT) > /dev/null

.PHONY: launch
launch:
	@fly launch --org Lightpanda

.PHONY: secrets
secrets: check-op
	@fly secrets set $(cdp_env)

.PHONY: deploy
deploy:
	fly deploy .

