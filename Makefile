# Usage examples {{{1
#
# See also:
# - https://docs.google.com/document/d/11oG00Nvn6vcFC2AemFmSkZNp0trEFrUHxL0IrkGR45c

# Check if 'make' runs from the directory where this Makefile resides. {{{1
$(if $(findstring /,$(MAKEFILE_LIST)),$(error Please only invoke this Makefile from the directory it resides in))
# Run all shell commands with bash. {{{1
SHELL := bash
# Locals {{{1

RECIPES = simulate trade order_s order_t \
					gobble_historical_umf gobble_local gobble_remote \
					gobble gobble_service_update gobble_up \
					feed feed_service_update feed_up \
					feed_hmf unzip_hmf

.PHONY: $(RECIPES)

Umf = /service/feed/log/main/current
Burp = /service/gobble/log/main/current
Order = /service/order/log/main/current
Res_T = /service/trade/log/main/current
Res_S = /service/simulate/log/main/current

HISTORICAL_UMF = historical_umf/main
REMOTE_FEED = mia-hub

# Default recipe: gobble_local {{{1
default_recipe: gobble_local

# Run service simulate {{{1
simulate: order_s
	@service/simulate.sh $(Order) $(Umf)

# Run service trade {{{1
trade: order_t
	@service/trade.sh $(Order)

# Run service order alongside service simulate {{{1
order_s: gobble
	@service/order.sh $(Burp) $(Umf) $(Res_S)

# Run service order alongside service trade {{{1
order_t: gobble
	@service/order.sh $(Burp) $(Umf) $(Res_T)

# Run service gobble locally {{{1
gobble_local:
	@sudo -E service/gobble.sh

# Run service gobble, pipe in from remote feed {{{1
gobble_remote:
	@REMOTE_FEED=$(REMOTE_FEED) service/gobble.sh

# Run service gobble, pipe in from local historical UMF data {{{1
gobble_historical_umf:
	@HISTORICAL_UMF=$(HISTORICAL_UMF) service/gobble.sh
# Run service feed {{{1
feed: feed_service_update feed_up
	@echo; echo "  Goals successful: $^"; echo

feed_service_update: $(GOPATH)/bin/feed \
service/feed.sh service/feed_run.sh service/feed_log_run.sh
	@sudo -E service/update.sh feed $(Umf)

$(GOPATH)/bin/feed: feed/feed.go json/umf.go
	@cd feed; go install; echo "- command feed installed in $(GOPATH)/bin"

feed_up:
	@sudo svstat /service/feed #; sudo tail -F $(Umf)

# Run service gobble {{{1
gobble: gobble_service_update gobble_up
	@echo; echo "  Goals successful: $^"; echo

gobble_service_update: $(GOPATH)/bin/gobble \
service/gobble.sh service/gobble_run.sh service/gobble_log_run.sh
	@sudo -E service/update.sh gobble $(Burp)

$(GOPATH)/bin/gobble: gobble/main.go json/umf.go
	@cd gobble; go install; echo "- command gobble installed in $(GOPATH)/bin"

gobble_up:
	@sudo svstat /service/gobble #; sudo tail -F $(Burp)


#feed2telete: unzip_hmf feed_hmf {{{1
#	@echo; echo "  Goals successful: $^"; echo

# Unzip historical market feed {{{1
unzip_hmf:
	@rm -rf archive 2>/dev/null; unzip -q archive

# Pipe USD market feed to tbot.sh {{{1
feed_hmf:
	@for f in archive/*.mf; do echo $$f | grep USD | xargs cat; done | ./tbot.sh
