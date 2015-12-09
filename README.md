Kraken
======

Kraken helps you to fetch a lot of URLs.

Hello I am KRAKEN - URLs are my prey:

```

/status

	GET: get the status of this kraken


/tentacle/<name>

	PUT / POST : create or overwrite a new tentacle with body {"bandwidth": <int>, "retry": <int>}
	PATCH      : patch the tentacle change it bandwidth and number of retries with body  {"bandwidth": <int>, "retry": <int>}
	GET        : get the status of an existing tentacle
	DELETE     : get rid of the tentacle


/tentacle/<name>/<preyId>

	PUT/POST   : let me catch some prey with body { "url" : <string>, "priority" : <int>, ["lock" : <string>, "method" : <string>, "body" : <string>, "tags" : [<string>, ...]] }

```

With "lock" you can define a resource name. Kraken will try to lock this resource before running a prey. This helps you to prevent running preys with the same resource in parallel.

curl
----

Some curl examples for locale development. Please consider NOT to run kraken in insecure mode on production systems!

start kraken

```bash
./kraken-linux -address "127.0.0.1:8080" -insecure
```

create tentacle

```bash
curl -k -H "Content-Type: application/json" -X PUT -d '{"retry":3, "bandwidth":2}' 127.0.0.1:8080/tentacle/foo
```

add a prey

```bash
curl -k -H "Content-Type: application/json" -X PUT -d '{ "url" : "https://www.google.com", "priority" : 100}' 127.0.0.1:8080/tentacle/foo/prey1
```

get status of a tentacle

```bash
curl -k 127.0.0.1:8080/tentacle/foo
```

delete tentacle

```bash
curl -k -X DELETE 127.0.0.1:8080/tentacle/foo
```
