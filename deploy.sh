#!/bin/bash

rsync . -a --exclude ".*" --progress cal@frontdoor:/home/cal/golinksgo/
ssh cal@frontdoor 'cd /home/cal/golinksgo && ./build_and_run.sh'
