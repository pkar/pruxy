# pruxy

A simple reverse proxy that is configured from etcd or environment variables.


### run

```bash
go get github.com/pkar/pruxy

export LOCAL_IP=127.0.0.1
go run $GOPATH/src/github.com/pkar/pruxy/cmd/main.go -port=6000 -prefix=pruxy -etcd=$LOCAL_IP:4001,$LOCAL_IP:4002

# or using environment variables

PRUXY_1="admin.dev.local=$127.0.0.1:8080,$127.0.0.1:8081" pruxy -prefix=PRUXY_
```

## Environment

Leave option -etcd option empty

Format of environment variables should be

```
# PREFIX_VAR="{Host}=upstream1:port1,upstream2:port2"

PRUXY_1="admin.dev.local=$127.0.0.1:8080,$127.0.0.1:8081" PRUXY_2="www.dev.local=$127.0.0.1:80" pruxy -prefix=PRUXY_
```

## Etcd

### setup etcd keys
etcd keys are stored and watched on the given prefix name

```
curl $LOCAL_IP:4001/v2/keys/pruxy/{hostname}/{upstream}

curl -L $LOCAL_IP:4001/v2/keys/pruxy/blog.example.com/$LOCAL_IP:5000 -XPUT -d value='1'
curl -L $LOCAL_IP:4001/v2/keys/pruxy/blog.example.com/$LOCAL_IP:5001 -XPUT -d value='1'
curl -L $LOCAL_IP:4001/v2/keys/pruxy/dir.example.com/$LOCAL_IP:5002 -XPUT -d value='1'
```

#### run some simple upstream servers

```
python -m SimpleHTTPServer 5000
python -m SimpleHTTPServer 5001
python -m SimpleHTTPServer 5002
```

### test out reverse proxy

```
curl -H "Host: blog.example.com" $LOCAL_IP:6000
curl -H "Host: dir.example.com" $LOCAL_IP:6000
```

### etcd containers

If you don't have etcd somewhere, try this from https://coreos.com/blog/Running-etcd-in-Containers/

```bash
export PUBLIC_IP=192.168.59.103
docker run -d -p 7001:7001 -p 4001:4001 --name etcd1 coreos/etcd -peer-addr ${PUBLIC_IP}:7001 -addr ${PUBLIC_IP}:4001 -peers ${PUBLIC_IP}:7002,${PUBLIC_IP}:7003
docker run -d -p 7002:7002 -p 4002:4002 --name etcd2 coreos/etcd -peer-addr ${PUBLIC_IP}:7002 -addr ${PUBLIC_IP}:4002 -peers ${PUBLIC_IP}:7001,${PUBLIC_IP}:7003
docker run -d -p 7003:7003 -p 4003:4003 --name etcd3 coreos/etcd -peer-addr ${PUBLIC_IP}:7003 -addr ${PUBLIC_IP}:4003 -peers ${PUBLIC_IP}:7001,${PUBLIC_IP}:7002

curl -L $PUBLIC_IP:4001/v2/stats/leader
```

### docker build

```bash
docker build -t pkar/pruxy .
docker run pkar/pruxy -port=6000 -prefix=pruxy -etcd=$LOCAL_IP:4001,$LOCAL_IP:4002,$LOCAL_IP:4003
```
