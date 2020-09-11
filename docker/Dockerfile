FROM golang:1.14
MAINTAINER libsgh
ENV GIN_MODE=release \
    PORT=8080
WORKDIR /app
COPY getPanIndex.sh /app
RUN chmod +x /app/getPanIndex.sh
RUN sh /app/getPanIndex.sh
EXPOSE 8080
CMD ["/app/PanIndex", "-config.path=/app/config.json"]