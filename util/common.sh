LOG_SIZE=s1000000
LOG_NUM=n44

log () { # {{{1
  echo `date +%s` $BASHPID $@ >>./syserr
}

pipe_in_from_remote_feed () {
  local ssh_client="$HOME/.ssh/client_session_$REMOTE_FEED"
  local client=$(hostname)
  local ssh_server=".ssh/server_session_$client"

# TODO sudo chmod 755 /service/feed/log/main/

  ( sleep 3; scp -q $ssh_client $REMOTE_FEED:$ssh_server ) &
  ssh -R 0:127.0.0.1:22 $REMOTE_FEED '{ cd /service/feed/log/main; \
    cat <(ls | grep '.s' | xargs cat); \
    tail -n 999999 -F current; \
  }' 2>$ssh_client
}
