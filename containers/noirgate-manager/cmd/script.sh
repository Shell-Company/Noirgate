# ~/projects/discodns/build/bin

./certbot-auto certonly --manual --agree-tos -d *.malwareroulette.io,*.noirgate.malwareroulette.io

service nginx stop
cp /etc/letsencrypt/live/malwareroulette.io/fullchain.pem /etc/nginx/cert.pem
cp /etc/letsencrypt/live/malwareroulette.io/privkey.pem /etc/nginx/key.pem
service nginx start

# start disco dns 
~/projects/discodns/build/bin/discodns -l 127.0.0.1 -p 9053 --etcd="http://127.0.0.1:2379" -m 0 -v&

# cleanup containers

docker stop $(docker ps -f name=noir* -q) && docker container prune

# Extract container contents 
sudo docker container cp <cont>:/tmp/ .  