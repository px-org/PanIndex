ARG APP_NAME
ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine as builder
LABEL stage=go-builder
ARG VERSION
ARG GO_VERSION
ARG TYPE
ARG APP_NAME
ENV GITHUB_REF=$VERSION
WORKDIR /app/
COPY ./ ./
RUN apk add --no-cache bash git curl gcc musl-dev; \
    curl -s -O 'https://raw.githubusercontent.com/px-org/build-action/main/build.sh'; \
    bash build.sh ${TYPE} ${APP_NAME}

FROM alpine:edge
MAINTAINER libsgh
ARG APP_NAME
ENV APP_NAME_ENV="${APP_NAME}"
WORKDIR /app
COPY --from=builder /app/bin/${APP_NAME} ./
CMD /app/$APP_NAME_ENV