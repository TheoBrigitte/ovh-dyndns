# OVH DynDns

If you have a domain name at OVH you can use this service as a DynDns-like client.
This service retrieve the public IP adresse or hostname and update DNS record accordingly.

- It relies on [ipify.org][1] in order to get the public IP adresse.
- Hostname is discover by issuing a reverse DNS lookup request.
- OVH API communication is handled by [go-ovh][2] client.

## Install

#### Binary

```
go get -d github.com/TheoBrigitte/ovh-dyndns
go build github.com/TheoBrigitte/ovh-dyndns
sudo mv ovh-dyndns /usr/local/bin
```

#### Configuration

Replace values in configuration files:
* ovh.conf hold OVH API credential. You can get those from [ovh.com][3]. For consumer key run `ovh-dyndns --consumer-key`
* ovh-dyndns.toml hold the DNS record(s) to be updated.

```
cp $GOPATH/github.com/TheoBrigitte/ovh-dyndns/config/ovh-dyndns.toml.sample /etc/ovh-dyndns.toml
cp $GOPATH/github.com/TheoBrigitte/ovh-dyndns/config/ovh.conf.sample /etc/ovh.conf
```

#### Systemd
```
sudo systemctl enable --now $GOPATH/github.com/TheoBrigitte/ovh-dyndns/systemd/ovh-dyndns.service
sudo systemctl enable $GOPATH/github.com/TheoBrigitte/ovh-dyndns/systemd/ovh-dyndns.timer
```

[1]: https://www.ipify.org/
[2]: https://github.com/ovh/go-ovh/
[3]: https://api.ovh.com/g934.first_step_with_api