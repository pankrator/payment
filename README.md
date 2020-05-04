# Payment system


## How to run locally

Run the script in local dev folder:

```sh
cd local_dev
./run.sh
```

## How to use

Open your browser and load `localhost:8000`.

### Users

Users can be found in users.csv file
There is one admin and two merchants

Admin user: **admin** with password **secret**

Merchant **ivan** with password **1234**

Merchant **koko** with password **1234**

## Run tests

Before running all tests, there must be UAA and Postgre containers running. Use the following command:

```sh
docker-compose up -d --force-recreate --build uaa db
```