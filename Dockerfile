FROM golang:1.8-alpine

#expose ports


ADD . /var/www/asteroid-lb

RUN apk add --update \
    git

RUN /var/www/asteroid-lb/install.sh

EXPOSE 8080

CMD ./run.sh