# Kraken

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

	PUT/POST   : let me catch some prey with body { "url" : <string>, "priority" : <int>, ["method" : <string>, "body" : <string>, "tags" : [<string>, ...] }

```
