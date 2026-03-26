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

**Systemd**

For a more robust setup, use Systemd to monitor the binary and restart upon failure etc. Sample file to be placed at `/etc/systemd/system/tls-proxy.service`:
```
[Unit]
Description=Challenge Proxy
Requires=network-online.target
After=network-online.target

[Service]
ExecStart=/srv/proxy/tls-proxy --cert /srv/proxy/certs/certificate.pem --key /srv/proxy/certs/certificate.key --dest /var/run/proxy.lst -m /srv/proxy/traffic.stats --host 0.0.0.0 -p 1337
Restart=on-failure
RestartSec=10

[Install]
WantedBy = multi-user.target
```

Install and enable using commands:
```sh
systemctl daemon-reload
systemctl enable tls-proxy
systemctl start tls-proxy
```

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
