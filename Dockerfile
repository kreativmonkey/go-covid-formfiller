FROM golang:1.16.4-buster as builder

LABEL maintainer="kreativmonkey@calyrium.org"
LABEL version="0.1"
LABEL description="This Docker is for creating Covid-Test PDF."

# Disable Prompt During Packages Installation
ARG DEBIAN_FRONTEND=noninteractive
ARG CONFIGFILE=config.yml

ENV APP_HOME /go/src/covidformer
ENV APP_USER app

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER && \
    mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

USER $APP_USER
WORKDIR $APP_HOME

COPY src/ .

RUN go mod download && go mod verify && go build -o covidformer

FROM debian:buster
FROM golang:1.16.4-buster

ENV APP_USER app
ENV APP_HOME /go/src/covidformer

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_HOME && \
    mkdir -p /data 
WORKDIR $APP_HOME
COPY src/conf/ conf/
COPY src/views/ views/
COPY src/formular.pdf .
COPY src/config.yml .
COPY --chown=0:0 --from=builder $APP_HOME/covidformer $APP_HOME
EXPOSE 8888
USER $APP_USER
CMD ["./covidformer"]