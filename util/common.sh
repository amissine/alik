LOG_SIZE=s500000
LOG_NUM=n22

log () { # {{{1
  echo `date +%s` $BASHPID $@ >>./syserr
}
