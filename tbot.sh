#!/usr/bin/env bash # {{{1

# Args {{{1
# Locals {{{1
. util.sh

# Quick tests {{{1
go run feed.go | pour umf.json | wc -l
