# Proxless

> Reduce your kubernetes cost by making all your deployments serverless with Proxless

Proxless is an open-source proxy made for Kubernetes.
Every deployment/service that are fronted by Proxless will be made serverless.
Proxless scale to 0 your deployments after N seconds idle and scale them back to 1 after a call to the URL.

## Installation

A helm chart will be available soon.

## Development Setup

Duplicate the `.env.example` file into a `.env` file and modify the variables accordingly

Then run

```shell script
$ go run cmd/main.go
```

## Meta

Benjamin APPREDERISSE - [@benhazard42](https://twitter.com/benhazard42)

Distributed under the MIT license. See ``LICENSE`` for more information.

## Contributing

1. Fork it (<https://github.com/bappr/kube-proxless/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request