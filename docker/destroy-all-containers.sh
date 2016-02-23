#!/bin/bash

set -x

cd `dirname $0`

destroys=`ls *-destroy-container.sh`

for destroy in $destroys
do
  bash $destroy
done
