FROM debian:stretch-slim
MAINTAINER libsgh
ENV GIN_MODE=release
ENV TZ=Asia/Shanghai
RUN apt-get update
RUN apt-get install -y curl
WORKDIR /app
COPY getPanIndex.sh /app
RUN chmod +x /app/getPanIndex.sh
RUN sh /app/getPanIndex.sh
CMD ["/app/PanIndex", "-config=/app/config.json"]