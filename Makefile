# Usage examples {{{1
#
# Check if 'make' runs from the directory where this Makefile resides. {{{1
$(if $(findstring /,$(MAKEFILE_LIST)),$(error Please only invoke this Makefile from the directory it resides in))
# Run all shell commands with bash. {{{1
SHELL := bash
# Locals {{{1
RECIPES = feed feed_hmf unzip_hmf

.PHONY: $(RECIPES)

# Default recipe: feed {{{1
default_recipe: feed

# Run feed {{{1
feed: unzip_hmf feed_hmf
	@echo; echo "  Goals successful: $^"; echo

# Unzip historical market feed {{{1
unzip_hmf:
	@rm -rf archive 2>/dev/null; unzip -q archive

# Pipe USD market feed to tbot.sh {{{1
feed_hmf:
	@for f in archive/*.mf; do echo $$f | grep USD | xargs cat; done | ./tbot.sh
