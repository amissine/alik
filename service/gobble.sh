#!/usr/local/bin/bash

. util/common.sh

if [ -n "$REMOTE_FEED" ]; then
  log REMOTE_FEED $REMOTE_FEED
  pipe_in_from_remote_feed | go run gobble/main.go
  log DONE
  exit 0
fi

{ cd /service/feed/log/main
  cat <(ls | grep '.s' | xargs cat)
  tail -n 999999 -F current
} | go run gobble/main.go 2>>./syserr
