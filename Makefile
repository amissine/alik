# Usage examples {{{1
#
# Check if 'make' runs from the directory where this Makefile resides. {{{1
$(if $(findstring /,$(MAKEFILE_LIST)),$(error Please only invoke this Makefile from the directory it resides in))
# Run all shell commands with bash. {{{1
SHELL := bash
# Locals {{{1
RECIPES = gobble feed feed_hmf unzip_hmf
UMF = /service/feed/log/main/current

.PHONY: $(RECIPES)

# Default recipe: gobble {{{1
default_recipe: gobble

# Run service gobble {{{1
gobble: feed
	@service/gobble.sh $(UMF)

# Run service feed {{{1
feed: feed/feed.go service/feed.sh
	@echo '- updating service $@...'; sudo -E service/update_feed.sh $(UMF)
	@sudo -E cat umf0.json >> /service/feed/sysin

#feed2telete: unzip_hmf feed_hmf {{{1
#	@echo; echo "  Goals successful: $^"; echo

# Unzip historical market feed {{{1
unzip_hmf:
	@rm -rf archive 2>/dev/null; unzip -q archive

# Pipe USD market feed to tbot.sh {{{1
feed_hmf:
	@for f in archive/*.mf; do echo $$f | grep USD | xargs cat; done | ./tbot.sh
