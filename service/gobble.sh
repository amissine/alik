#!/usr/bin/env bash

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
  pipe_in_from_remote_feed
  exit 0
fi

if [ -z "$UMF" ]; then # error
  exit 69
fi

kill_pids () {
  pids=' '
  while true; do
    shift 2; if [ ! $? -eq 0 ]; then break; fi
    pids="$pids $1"; shift 4
  done
  echo "- kill_pids: killing pids$pids..."
  sudo kill $pids
}
kill_feeds () {
  echo; echo '- shutting down /service/feed...'; sudo svc -d /service/feed
  kill_pids $(cat)
}
trap "sudo cat /service/feed/syserr | grep 'feed started' | kill_feeds" SIGINT

echo '- hit Ctrl-C to clean up and exit'; echo
sudo -E tail -F $UMF
