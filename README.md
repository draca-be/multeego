# Multeego

Multeego (*pronunciation: multi go*) is a tool for developers of multi-service applications. By writing a very simple
configuration file you can quickly set up a local development instance of your cloud application.

## Usage

When your production application is built out of multiple processes, whether they are running in Kubernetes, serverless
functions or a simple docker instance, usually there is some kind of proxying webserver like Traefik or nginx in front
of it to redirect requests to the correct process.

If you want to recreate this locally often this involves multiple open terminals, docker-compose and Dockerfiles,
multiple running watchers or process managers, a local webserver, ...

Multeego seeks to eliminate this configuration headache so that developers can focus on developing again.

## Quick start

The fastest way to see how Multeego works is by example. This example is also in the [example](example/) directory.

You need a multeego.yaml file
```yaml
port: 8080
services:
  - path: /red/
    cwd: ./red-src
    port: 8060
    command: go
    arguments: run main.go -port ${PORT}

  - path: /blue/
    cwd: ./blue-src
    port: 8070
    command: fresh

  - path: /
    cwd: ./static
```

You can try it yourself by checking out this repository, cd into the example directory and running multeego. You'll need
to have fresh installed for the example to work (or adapt the blue process to also use go run like the red process).

```shell
$ go get github.com/pilu/fresh
$ cd example
$ go run ../
```

In another terminal, try to curl some URL's

```shell
$ curl http://localhost:8080/blue/

Hello from blue! I was called with URL /
```

The request to the `/blue/` path was redirected to the blue process.

```shell
$ curl http://localhost:8080/blue/subpath

Hello from blue! I was called with URL /subpath
```

Any sub-paths are automatically redirected to the relevant sub-path on the blue process, stripping the base path from 
the URL by default.

```shell
$ curl http://localhost:8080/red/

Hello from red! I was called with URL /
```

The request to the `/red/` path was redirected to the red process.

```shell
$ curl http://localhost:8080/
Hello from static content.
```

You can also host some static content if you want.

## multeego.yaml

This is the basic format of `multeego.yaml`

```yaml
port: <listenPort>
services:
  - path: <servicePath>
    cwd: <serviceCWD>
    port: <servicePort>
    command: <serviceCommand>
    arguments: <serviceArguments>
  - ...
```

- `<listenPort>`
  - Defines the port the main process will listen on.
- `<servicePath>`
  - Is the URL path that will be used to match the service
  - Since the path patterns are directly passed into the Go http library, the rules follow what is described in the
    [ServeMux](https://pkg.go.dev/net/http#ServeMux) documentation.
- `<serviceBase>`
  - Specifies the base path that will be stripped from the request URL before passing on to the service.
  - If not set, it defaults to `<servicePath>` so any requests to `/<servicePath>/<target>` are seen by the service as
    requests to `/<target>`.
- `<serviceCWD>`
  - Is the working directory where the service will be executed. Typically, this is your project folder.
  - If this is a relative path, it is relative to the location of `multeego.yaml`.
- `<servicePort>`
  - The port where the service will be listening on.
  - This port is also exported to the PORT environment variable.
  - If left out or set to 0, the files in `<serviceCWD>` will be served as static content.
- `<serviceCommand>`
  - The executable that Multeego should start.
  - This is a single parameter, do not pass your arguments here.
  - If `<servicePort>` is not set, this parameter is ignored.
- `<serviceArguments>`
  - The arguments to pass to the command.
  - If you want to specify the port number in the arguments you can use the template ${PORT}, it will be replaced by the
    actual port number on execution.
  - If `<servicePort>` is not set, this parameter is ignored.
  - If `<serviceCommand>` is not set, this parameter is ignored.

### Static file hosting

As stated above, if `<servicePort>` is not set or 0, Multeego will switch to serving files in `<serviceCWD>` as static
content. This is useful if you have some static content to host that isn't a part of the rest of your services.

Note that as long as the target URL has a matching file, Multeego will just dump that browser to the browser. While it
does some URL cleaning, under no circumstances assume that this is a safe implementation. If the file is not found, a
404 is returned as expected.

To make the browser experience a little more enjoyable, if the target URL is not a file but a directory Multeego will 
expand the path to `/<directory>/index.html`.

### On port numbering

Multeego makes no assumptions on your configuration, meaning that it will also not try to stop you if you write crazy
configurations. If you reuse a port number, it will happily forward the requests for you.

- If you do not specify a `<serviceCommand>`, you can for example retarget specific paths to specific paths of one of
  the other services.
- You can redirect to ports not managed by Multeego, for example that one process you want to start in your IDE for
  some hard-core debugging while you need a bunch of other services to be online.

## Technologies

Multeego is written and tested using Go 1.17 on a Gentoo system.

It will probably work on any Linux distribution but has (not) yet been tested on other platforms

## FAQ

#### This is cool, can I use it in production environments?

Short answer: no, you shouldn't.

Long answer: no you shouldn't. Multeego does very limited cleaning of the URL but that's it. It is designed with 
development in mind where you want to run your application on your own OS, with access to any and all resources. There
is no containerization (although you could run it in a container if you want), no segregation of data, no nothing.

#### I need to expose my service for an external integration test. What about SSL?

Just put a nginx, a Traefik, a ngrok tunnel or anything else that can handle this in front of it.

## TODO

Some day
- Other platforms
- Hot-reloading to not depend on other packages
- Be more robust for process failures
- Properly handle process output
- Add configuration of extra environment variables

# Disclaimer

USE AT YOUR OWN RISK

- The code is very alpha right now, meaning it can cause hanging processes, random crashes or data loss.
- This should go without saying but **DO NOT USE MULTEEGO IN PRODUCTION ENVIRONMENTS**

# License

Multeego is licensed under the GNU General Public License v3.0
