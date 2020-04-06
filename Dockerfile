# download dependencies
FROM golang:1.13 AS base
RUN go get -u github.com/valyala/quicktemplate/qtc
RUN mkdir -p /root/app
WORKDIR /root/app
ADD go.mod .
ADD go.sum .
RUN go mod download

# build app
FROM base AS build
COPY . .
RUN make

# final image
FROM scratch
COPY --from=build /root/app/app.cmd /app.cmd
CMD ["/app.cmd"]