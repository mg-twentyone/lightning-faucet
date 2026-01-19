FROM golang:1.23-alpine

# install git and curl for Air
RUN apk add --no-cache git curl

WORKDIR /app

# install Air binary
RUN curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

COPY . .

# use the vendor folder
ENV GOFLAGS="-mod=vendor"

CMD ["air", "-c", ".air.toml"]