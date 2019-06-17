# build stage
FROM golang:alpine AS build-env
RUN apk add git
ADD . /src
RUN cd /src && go build -o mongo_exerciser
# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/mongo_exerciser /app/
CMD ["/app/mongo_exerciser"]