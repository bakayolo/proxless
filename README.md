# Proxless

> Reduce your kubernetes cost by making all your deployments **on-demand** with Proxless

Proxless is an **open-source proxy for Kubernetes**.

It fronts the services you configure and makes the deployment associated **on-demand**.  
Basically, it scales up and down your deployments based on the demand (service requested or idle).

**No need a CRD, no need a huge stack, the proxless deployment is the only thing u need.**

## Disclaimer

Proxless is provided in alpha mode.  
Using it on your production cluster is done at your own risks.

## Installation

A helm chart is available [here](helm)

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