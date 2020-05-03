#!/bin/sh

docker-compose up -d --force-recreate --build uaa db

until curl -sfk http://localhost:8080/.well-known/openid-configuration > /dev/null; do
    echo "Waiting for UAA to be up and running"
    sleep 2
done

docker-compose up -d --no-deps --force-recreate --build payment_system
