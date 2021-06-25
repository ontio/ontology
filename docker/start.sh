#!/bin/bash

set -e
# ensure the ontology wallet.dat
if [ ! -f /app/${NODE}.dat ]; then
    if [[ "$NODE" == "NODE1" ]]; then
        for i in {1..7}
        do
            ./ontology account add -d -w NODE${i}.dat <<EOF
server
server
EOF
            ADDR=`eval cat NODE${i}.dat|jq .accounts |jq '.[0].address' |sed 's/"//g'`
            PUB=`eval cat NODE${i}.dat|jq .accounts |jq '.[0].publicKey' |sed 's/"//g'`
            sed -i "s/__NODE${i}ADDR__/$ADDR/g" config.json
            sed -i "s/__NODE${i}PUB__/$PUB/g" config.json

            echo "CHANGE config.json done"
        done
        cp /app/config.json /data/config.json
        cp /app/${NODE}.dat /data/wallet.dat

        # backgroud file server to serve config.json
        nohup caddy file-server -listen :9999 -root /app > /dev/null 2>&1 &
    else
        echo "node1 generating wallet, please wait..."
        sleep 50
        wget -c node1:9999/config.json -O /data/config.json
        wget -c node1:9999/${NODE}.dat -O /data/wallet.dat
    fi

fi

# log to docker stdout
echo "server" | ./ontology --config=/data/config.json --wallet /data/wallet.dat --data-dir /data/Chain --enable-consensus --loglevel=${LOG_LEVEL} --networkid=${NETWORKID} --disable-log-file
