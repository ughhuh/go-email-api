# syntax=docker/dockerfile:1

FROM golang:1.22

# Set destination for copying
WORKDIR /app

# create log dir
RUN mkdir -p ./logs

# copy config data and download go modules
COPY go.mod go.sum .env config.json ./
RUN go mod download

# copy source files
COPY ./cmd/app/*.go ./

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-email-api

# expose port
EXPOSE 8080

# Run
ENTRYPOINT [ "/docker-email-api" ]
