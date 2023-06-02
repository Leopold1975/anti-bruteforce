FROM golang:1.19

WORKDIR /abf

COPY ./bin .
COPY ./configs ./configs

ENTRYPOINT [ "/abf/anti-bruteforce" ]