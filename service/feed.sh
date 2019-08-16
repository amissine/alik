#!/bin/sh
touch ./sysin; chgrp admin ./sysin; chmod 660 ./sysin;
touch ./syserr; chgrp admin ./syserr; chmod 660 ./syserr;
./feed <./sysin 2>>./syserr
