# syntax=docker/dockerfile:1

FROM golang:1.22 AS build

# Set destination for copying
WORKDIR /app

# download project dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy source files
COPY ./cmd/app/*.go ./

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-email-api

FROM scratch AS run

# copy build from the previous stage
COPY --from=build /docker-email-api /
COPY config.json /

# expose port
EXPOSE 8080

# Run
ENTRYPOINT [ "/docker-email-api" ]
