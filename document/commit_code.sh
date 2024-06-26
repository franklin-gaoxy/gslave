#!/bin/bash

git add .
hn=`hostname`
us=`whoami`
git commit -m "[${hn}:${us}]:Auto submit"
git push me master
