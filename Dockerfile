# Build stage
FROM golang:1.14.4
# Set up build dir
WORKDIR /build
# Copy in source files
COPY ./ ./
# Build the binary
RUN make linux_amd64

# Container stage
FROM gcr.io/distroless/static
WORKDIR /app/
COPY --from=0 /build/dist/linux_amd64 .
USER nonroot:nonroot

ENV PORT="8000"
ENV ADDRESS="0.0.0.0"

CMD [ "/app/jamjar-relay-server" ]
