# Payment system


## How to run locally

* Run postgres DB in Docker container

```sh
docker run -e POSTGRES_PASSWORD=payment -e POSTGRES_USER=payment -e POSTGRES_DB=payment -d -p5432:5432 postgres
```