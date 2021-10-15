FROM arm64v8/debian:10-slim
MAINTAINER libsgh
ENV TZ=Asia/Shanghai
RUN apt-get update
RUN apt-get install -y curl
WORKDIR /app
COPY getPanIndex.sh /app
RUN chmod +x /app/getPanIndex.sh
RUN sh /app/getPanIndex.sh
CMD ["/app/PanIndex"]