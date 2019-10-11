#!/usr/local/bin/bash

. util/common.sh

if [ -n "$REMOTE_FEED" ]; then
  log REMOTE_FEED $REMOTE_FEED
  pipe_in_from_remote_feed | go run gobble/main.go
  log DONE
  exit 0
fi

if [ -z "$UMF" ]; then # error, empty file name
  exit 69
fi

{ cd /service/feed/log/main
  cat <(ls | grep '.s' | xargs cat)
  cat $UMF 
} | go run gobble/main.go # wc -l
