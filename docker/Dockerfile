FROM alpine
MAINTAINER libsgh
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --update curl && rm -rf /var/cache/apk/*
ENV GIN_MODE=release \
    PORT=8080
WORKDIR /app
COPY getPanIndex.sh /app
RUN chmod +x /app/getPanIndex.sh
RUN sh /app/getPanIndex.sh
EXPOSE 8080
CMD ["/app/PanIndex"]