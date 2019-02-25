FROM yyoshiki41/radigo AS radigo

FROM ruby:2.6-alpine

# Set timezone
ENV TZ "Asia/Tokyo"
# Set default output dir
VOLUME ["/output"]

RUN apk add --no-cache ca-certificates ffmpeg rtmpdump tzdata

COPY --from=radigo /bin/radigo /bin/radigo

ENV RADIGO_PATH=/bin/radigo

WORKDIR /usr/src/app

COPY Gemfile Gemfile
COPY Gemfile.lock Gemfile.lock

RUN bundle install

COPY . .