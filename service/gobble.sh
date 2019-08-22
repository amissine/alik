#!/usr/bin/env bash

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
sudo -E tail -F $1
