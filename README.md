# cronitor

Crontab based simple website monitoring.
Try to GET a site, measures transfer time and look for a keyword.
Check server.conf.dist for configuring information.

```
$ go get github.com/gleicon/cronitor
$ go install github.com/gleicon/cronitor
```
## 1 minute intervals checks

```
$ crontab -e
```

add:
*/1 * * * * $GOPATH/bin/cronitor -c /etc/server.conf

save and exit. 

