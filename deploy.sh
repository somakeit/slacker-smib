#!/bin/bash

set -e
set +x
echo -e $DEPLOY_KEY | base64 -d > ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa
set -x

ssh -o StrictHostKeyChecking=no -p 6022 root@space.somakeit.org.uk service slacker-smib stop
scp -o StrictHostKeyChecking=no -P 6022 smib root@space.somakeit.org.uk:/usr/local/bin
cp init-script slacker-smib
set +x
sed -i "s/<TOKEN>/$SLACK_TOKEN/" slacker-smib
set -x
scp -o StrictHostKeyChecking=no -P 6022 slacker-smib root@space.somakeit.org.uk:/etc/init.d
ssh -o StrictHostKeyChecking=no -p 6022 root@space.somakeit.org.uk service slacker-smib start