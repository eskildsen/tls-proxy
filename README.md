# Dynamic TLS Reverse Proxy

This service listens on a single port, and forwards traffics to other ports on the same host based on the TLS hostname. The destination ports are continuously loaded from a file.

## Setup

**Certificates**

For supporting multiple hostnames, ensure to have a wildcard certificate, e.g. `*.domain.com`. Save it as `certs/certificate.pem`. Put the corresponding key in `certs/priv.key` or adjust command line accordingly.

**Destination Ports**

Create the file targets.txt and put in lines in the format `<hostname>:<port>`, e.g.
```
danger.domain.com:10000
db.domain.com:3306
```

While running this file will be monitored for changes and new hosts will be immediately available.

## Usage

```sh
$ ./tls-proxy -help
Usage of ./tls-proxy:
  -cert string
        the certificate to use for incoming TLS (default "certs/certificate.pem")
  -dest string
        a file containing hostnames and corresponding local ports for end destination (default "targets.txt")
  -host string
        host to listen on (default "0.0.0.0")
  -key string
        the corresponding certificate key (default "certs/priv.key")
  -m string
        filename for storing traffic metrics
  -p int
        port to listen on (default 1337)
```
