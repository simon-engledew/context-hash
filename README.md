# docker-context-hash

Generate a repeatable hash of a docker context.

```shell script
$ docker-context-hash ./example
92fb07f5612143976c8a6420f2f022ce094bb3573d826f063bc4ff18f508c829
```

This is quick, only considers the context and no work is performed in Docker.

If your Docker image does not guarantee reproducible dependencies you can still end up with two 
different images sharing the same hash:

```Dockerfile
FROM alpine:latest

ENTRYPOINT ["/bin/sh"]
```

> One month this builds with alpine:3.12.0, another month it builds with alpine:3.11.1
>
> They share the same hash.

`docker-context-hash` is useful as a component in a caching mechanism in a CI/CD system, or an alternative to manually incrementing version numbers.

```shell script
docker build ./image -t image:$(docker-context-hash ./image)
```