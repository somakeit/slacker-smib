#!/bin/bash

set -e
set +x
echo -e $DEPLOY_KEY | base64 -d > ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa
set -x

scp -p 6022 smib root@space.somakeit.org.uk:/usr/local/bin
cp init-script slacker-smib
sed -i "s/<TOKEN>/$SLACK_TOKEN/" slacker-smib
scp -p 6022 slacker-smib root@space.somakeit.org.uk:/etc/init.d
ssh -p 6022 root@space.somakeit.org.uk service slacker-smib stop
sleep 5
ssh -p 6022 root@space.somakeit.org.uk service slacker-smib start