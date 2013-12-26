#!/bin/bash
# if you're paranoid, you can use this to backup your tmsu db before every pixie run

cp ~/.tmsu/default.db ~/.tmsu/default.db.$(date +%F-%T)
go run pixie.go
