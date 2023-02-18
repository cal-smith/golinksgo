#!/bin/bash

go build -v -tags 'sqlite_fts5' .
sudo /bin/systemctl stop golinks && /bin/systemctl status golinks
mv golinks index.html /home/cal/golinks/
sudo /bin/systemctl start golinks && /bin/systemctl status golinks