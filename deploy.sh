#!/bin/bash

go build -tags 'sqlite_fts5' .
ssh cal@frontdoor 'sudo /bin/systemctl stop golinks && /bin/systemctl status golinks'
scp golinks index.html cal@frontdoor:/home/cal/golinks/
ssh cal@frontdoor 'sudo /bin/systemctl start golinks && /bin/systemctl status golinks'