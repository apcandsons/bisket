# bisqit

Tired of waiting for deployment? Let’s enjoy the ***blazing-fast deployment*** in a bitesize!!

## Motivation

Container technology is great, but the deploy experience is just horrible.

Image creation is time-consuming, difficult to optimize, and lacks scalability. After it's available in the repository, it needs to be pulled, mounted, and then there's the additional overhead involved in running a new instance and updating traffic. This process consumes a significant amount of time for engineers daily. In my opinion, this technology is over-engineered, resulting in the wastage of substantial networking and computing resources every day.

bisqit is a lightweight application switcher, designed to be used in a critical applications where turnover speed matters. It provides a seamless way to switch between different applications versions without consuming your time, enhancing productivity while ensuring a reliable and smooth deployment experience.

## tl;dr

Install bisqit

```bash
$ brew install bisqit
```

Initialize using `bisq init` command

```bash
$ bisq init
```

This will create `bisqit.yaml`, use your editor to modify however needed.

```yaml
bisqit_server_port: 8001
run:
  - dist/server -p $BISQIT_PORT
repository:
	git:
	  repository: https://github.com/smithee/amazing_app
		secret: <git-api-key>
```

<aside>
💡 Note that your application needs to either read the environment variable of `BISQIT_PORT` or be passed as an argument to change the port on which the application runs.

</aside>

You can actually run bisqit server on your local device to test. So, just do this:

```bash
$ bisq server start
Bisqit server running on 8001
Reading https://github.com/smithee/amazing_app
Detected latest version: @1.0.0
Pulling latest version: @1.0.0
Running amazing_app@1.0.0 on 18000 # some arbitrary that's not in use
Ready to route amazing_app@1.0.0 on 18000
Routing tcp:8001 -> 18000
```

Test your app on [`http://localhost:8001`](http://localhost:8001) and see what happens!

## Update Version Instance

Let’s make an update to your application.

Keep the server running, and open a separate terminal, and do this:

```bash
$ sed -i 's/Hello, world!/Hello, Japan!/g' src/main.go
$ make build # whatever it takes to build
$ git commit -am "Saying hello to Japan"
$ git tag @1.0.1
$ git git push origin @1.0.1
```

Now see what happens on your bisq server:

```bash
$ bisq server start
Bisqit server running on 8001
...
Routing tcp:8001 -> 18000
Detected latest version: @1.0.1
Pulling latest version: @1.0.1 ... Done
Running amazing_app@1.0.1 on 18001 ... Done
Waiting for healthy state on amazing_app@1.0.1 on 18001 ... Done
Routing tcp:8001 -> 18001
Ready to destroy previous version: @1.0.0
Destroying previous version: @1.0.0 ... Done
```

```bash
$ bisq server start
Bisqit server running on 8001
...
Routing tcp:8001 -> 18000
Detected latest version: @1.0.1
Pulling latest version: @1.0.1 ... Done in 3.2s
Running amazing_app@1.0.1 on 18001 ... Done in 0.2s
Waiting for healthy state on amazing_app@1.0.1 on 18001 ... Done in 10.5s
Routing tcp:8001 -> 18001
Ready to destroy previous version: @1.0.0
Destroying previous version: @1.0.0 ... Done
```

Ah! There we go. Now, reload your browser and see the changes on `https://localhost:8001`

## Rollback Version Instance

Let’s say your boss tells you to rollback. Okay, let’s delete the tag and see what happens.

```bash
git push origin --delete @1.0.1
```

```bash
$ bisq server start
Bisqit server running on 8001
...
Routing tcp:8001 -> 18001
Ready to destroy previous version: @1.0.0
Destroying previous version: @1.0.0Routing tcp:8001 -> 18001
Detected latest version: @1.0.0
Pulling latest version: @1.0.0 ... Done
Running amazing_app@1.0.0 on 18002 ... Done
Ready to route amazing_app@1.0.0
```

## SSL Support

Security is important!

## Preserve Version Instances

By default, bisqit server auto destroys previous version to save resources on your computing environment. However, you can add `preserve_previous_generations` attribute so that it preserves past instances.

```yaml
bisqit_server_port: 8001
preserve_previous_generations: 2 # this would allow you to
```

Let’s see what happens when you update from version `@1.0.0` to `1.0.1` .

```bash
$ bisq server start
Bisqit server running on 8001
...
Routing tcp:8001 -> 18000
Detected latest version: @1.0.1
Pulling latest version: @1.0.1 ... Done
Running amazing_app@1.0.1 on 18001 ... Done
Waiting for healthy state on amazing_app@1.0.1 on 18001 ... Done
Routing tcp:8001 -> 18001
Ready to destroy previous version: @1.0.0
```

Now delete the latest version:

```yaml
git push origin --delete @1.0.1
```

```bash
$ bisq server start
Bisqit server running on 8001
...
Detected latest version: @1.0.0
Detected running instance of amazing_app@1.0.0 on 18000
Routing tcp:8001 -> 18000
Ready to destroy invalidated version: @1.0.1
Destroying invalidated version: @1.0.1 ... Done
```

You can observe that the rollback speed is blazing fast.

## The Preview Mode

One of the advantages on using bisqit is that it provides a mechanism to run multiple instances of applications for testing purposes.  And way you control which application version to connect, is through providing a HTTP Headers (Defaults to `x-bisq-preview`)

To enable preview mode, set `preview_mode: true` in your `bisqit.yaml`

```bash
bisqit_server_port: 8001
preview_mode: true
```

To take an advantage of preview mode, you can use the `@preview/1.x.x` tag.

```yaml
$ sed -i 's/Hello, world!/Hello, Preview!/g' src/main.go
$ make build # whatever it takes to build
$ git commit -am "Saying hello to Preview audience"
$ git tag @preview/1.0.1
$ git git push origin @preview/1.0.1
```

```bash
$ bisq server start
Bisqit server running on 8001
Preview mode: on
...
Routing tcp:8001 -> 18000
Detected preview version: @preview/1.0.1
Running instance of amazing_app@preview/1.0.1 on 18001
Routing tcp:8001 -> 18001 as preview (add header `x-bisq-preview: 1.0.1`)
```

Now you’ll see that the preview version is detected, and now running on 18001. When you make a call to the server with the header, it behaves differently.

```bash
$ curl http://localhost:8001/
Hello, World!
$ curl -H "x-bisq-preview: 1.0.1" http://localhost:8001/
Hello, Preview!
```

Ah, isn’t this great.

## Change the Log Level

To change the log level, use `log_level` attribute in the config.

```bash
bisqit_server_port: 8001
log_level: info
```

There are five levels of `log_level`:

- `debug`
- `info`
- `warn`
- `error`
- `fatal`

By default, it is set to warn. However, `info` would allow you to see per-request events.

```bash
$ bisq server start
Bisqit server running on 8001
...
Routing tcp:8001 -> 18001
2024-01-01T12:34:56+0900 [INFO] a0fjem HTTP/1.0 connected from 127.0.0.1
2024-01-01T12:34:56+0900 [INFO] a0fjem HTTP/1.0 disconnected sent 50 bytes
```
