# Pre-requisite: Docker
sudo docker network create --subnet=172.18.0.0/16 stg-net
./e2e 60 #time in minutes for which chaos needs to be run

