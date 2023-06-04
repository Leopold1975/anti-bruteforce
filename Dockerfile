FROM ubuntu:20.04

WORKDIR /abf

COPY ./bin .
COPY ./configs ./configs

ENTRYPOINT [ "/abf/anti-bruteforce" ]