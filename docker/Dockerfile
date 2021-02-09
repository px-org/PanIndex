FROM golang:1.14
MAINTAINER libsgh
ENV GIN_MODE=release
WORKDIR /app
COPY getPanIndex.sh /app
RUN chmod +x /app/getPanIndex.sh
RUN sh /app/getPanIndex.sh
CMD ["/app/PanIndex", "-config=/app/config.json"]