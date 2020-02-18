#!/usr/local/bin/bash # gobble.sh {{{1

. util/common.sh

if [ -n "$REMOTE_FEED" ]; then # {{{1
  log REMOTE_FEED $REMOTE_FEED
  pipe_in_from_remote_feed | ./gobble 2>>./syserr
  log DONE
  exit 0
fi

if [ -n "$HISTORICAL_UMF" ]; then # {{{1
  { cd "$HISTORICAL_UMF"
    cat <(ls | grep '\.s' | xargs cat); echo
  } | go run gobble/main.go 2>>./syserr
  exit 0
fi

{ cd /service/feed/log/main # {{{1
  cat <(ls | grep '\.s' | xargs cat)
  tail -n 999999 -F current
} | ./gobble 2>>./syserr
