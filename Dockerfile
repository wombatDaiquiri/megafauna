FROM golang:1.19

WORKDIR /usr/src/megafauna

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
# COPY go.mod go.sum ./
# RUN go mod download && go mod verify

COPY . .
RUN go build -mod vendor -v -o /usr/local/bin/megafauna

CMD ["megafauna"]
# 30 seconds compile time lmao