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
					gobble gobble_service_update gobble_up \
					feed_service_update feed_up \
					feed_hmf unzip_hmf

.PHONY: $(RECIPES)

Umf = /service/feed/log/main/current
Burp = /service/gobble/log/main/current
Order = /service/order/log/main/current
Res_T = /service/trade/log/main/current
Res_S = /service/simulate/log/main/current

REMOTE_FEED = mia-hub

# Default recipe: feed {{{1
default_recipe: feed

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

# Run service gobble {{{1
#gobble: gobble_command_install gobble_service_update gobble_up feed
#	@UMF=$(Umf) service/gobble.sh

# Run service gobble, pipe in from remote feed {{{1
gobble: gobble_command_install gobble_service_update gobble_up
	@REMOTE_FEED=$(REMOTE_FEED) service/gobble.sh

# Run service feed {{{1
feed: feed_service_update feed_up
	@echo; echo "  Goals successful: $^"; echo

feed_service_update: $(GOPATH)/bin/feed \
service/feed.sh service/feed_run.sh service/feed_log_run.sh
	@sudo -E service/update.sh feed $(Umf)

$(GOPATH)/bin/feed: feed/feed.go json/umf.go
	@cd feed; go install; echo "- command feed installed in $(GOPATH)/bin"

feed_up:
	@sudo svstat /service/feed

#feed2telete: unzip_hmf feed_hmf {{{1
#	@echo; echo "  Goals successful: $^"; echo

# Unzip historical market feed {{{1
unzip_hmf:
	@rm -rf archive 2>/dev/null; unzip -q archive

# Pipe USD market feed to tbot.sh {{{1
feed_hmf:
	@for f in archive/*.mf; do echo $$f | grep USD | xargs cat; done | ./tbot.sh
