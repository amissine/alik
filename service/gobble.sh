#!/usr/local/bin/bash

pipe_in_from_remote_feed () {
  local ssh_client="$HOME/.ssh/client_session_$REMOTE_FEED"
  local client=$(hostname)
  local ssh_server=".ssh/server_session_$client"

  ( sleep 3; scp -q $ssh_client $REMOTE_FEED:$ssh_server ) &
  ssh -R 0:127.0.0.1:22 $REMOTE_FEED '{ cd /service/feed/log/main; \
    cat <(ls | grep '.s' | xargs cat); \
    echo "{\"current\":\"=============================================\"}"; \
    tail -n 999999 -F current; \
  }' 2>$ssh_client
}

if [ -n "$REMOTE_FEED" ]; then
  pipe_in_from_remote_feed | go run gobble/main.go
  exit 0
fi

if [ -z "$UMF" ]; then # error, empty file name
  exit 69
fi

{ cd /service/feed/log/main
  cat <(ls | grep '.s' | xargs cat)
  #echo "{\"current\":\"=============================================\"}"
  cat $UMF 
} | go run gobble/main.go # wc -l
