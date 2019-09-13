#!/bin/bash

set -e
set +x
echo -e $DEPLOY_KEY > ~/.ssh/id_deploy
chmod 600 ~/.ssh/id_deploy
set -x

scp -p 6022 smib root@space.somakeit.org.uk:/usr/local/bin/smib
sed -i "s/<TOKEN>/$SLACK_TOKEN/" init-script
scp -p 6022 smib root@space.somakeit.org.uk:/etc/init.d/slacker-smib
ssh -p 6022 root@space.somakeit.org.uk service slacker-smib stop
sleep 5
ssh -p 6022 root@space.somakeit.org.uk service slacker-smib start