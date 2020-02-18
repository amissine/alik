# Locals {{{1
LOG_SIZE=s10000000
LOG_NUM=n44

log () { # {{{1
  echo `date +%s` $BASHPID $@ >>./syserr
}

pipe_in_from_local_current () { # {{{1
  log 'pipe_in_from_local_current started'
  cd /service/feed/log/main
  tail -n 999999 -F current
}

pipe_in_from_remote_current () { # {{{1
  local ssh_client="$HOME/.ssh/client_session_$REMOTE_FEED"
  local client=$(hostname)
  local ssh_server=".ssh/server_session_$client"

# TODO sudo chmod 755 /service/feed/log/main/

  ( sleep 3; scp -q $ssh_client $REMOTE_FEED:$ssh_server ) &
  ssh -R 0:127.0.0.1:22 $REMOTE_FEED '{ cd /service/feed/log/main; \
    tail -n 99999 -F current; \
  }' 2>$ssh_client
}

pipe_in_from_remote_feed () { # {{{1
  local ssh_client="$HOME/.ssh/client_session_$REMOTE_FEED"
  local client=$(hostname)
  local ssh_server=".ssh/server_session_$client"

# TODO sudo chmod 755 /service/feed/log/main/

  ( sleep 3; scp -q $ssh_client $REMOTE_FEED:$ssh_server ) &
  ssh -R 0:127.0.0.1:22 $REMOTE_FEED '{ cd /service/feed/log/main; \
    cat <(ls | grep '\.s' | xargs cat); \
    tail -n 999999 -F current; \
  }' 2>$ssh_client
}
