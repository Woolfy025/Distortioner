FROM golang:1.17-bullseye as build
WORKDIR /go/src/distortioner
COPY . .
RUN go build

FROM ghcr.io/graynk/ffmpegim as release

RUN mkdir app
WORKDIR app
COPY --from=build /go/src/distortioner/distortioner distortioner

ENTRYPOINT ["./distortioner"]