LOG_SIZE=s1000000
LOG_NUM=n44

log () { # {{{1
  echo `date +%s` $BASHPID $@ >>./syserr
}
