#!/usr/bin/env bash # {{{1

# Args {{{1
# Locals {{{1
. util.sh

# Quick tests {{{1
#go run feed.go | cat -u >> umf0.json; wc -l umf0.json
#f=umf.json; rm $f; touch $f; cat -u umf0.json | { tail -n 999999 -F $f & pid=$!; cat -u >> $f; sleep 0.2; kill $pid; } | wc -l; diff umf0.json $f; echo $?
f0=umf0.json; f=umf.json; cat -u $f0 | pour $f | wc -l; diff $f0 $f; echo $?
