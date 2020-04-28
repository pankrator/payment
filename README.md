# Payment system


## How to run locally

* Run postgres DB in Docker container

```sh
docker run -e POSTGRES_PASSWORD=payment -e POSTGRES_USER=payment -e POSTGRES_DB=payment -d -p5432:5432 postgres
```

* Build UAA Server

```sh
docker build -t payment_uaa -f local_dev/uaa/Dockerfile local_dev/uaa/.
```

* Run UAA Server

```sh
docker run -d -p8080:8080 --name uaa payment_uaa
```