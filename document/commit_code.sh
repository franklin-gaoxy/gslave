#!/bin/bash

git add .
hn=`hostname`
us=`whoami`
git commit -m "[${hn}:${us}]:Auto submit"
if [[ "$us" == "wb.gaoxiuyang01" ]];then
    git push
else 
    git push me master
fi

