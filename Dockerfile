FROM alpine

ADD bin/linux_amd64/pruxy /usr/bin/

CMD pruxy -port=6000 -prefix=PRUXY_
