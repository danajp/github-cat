FROM ruby:2.6.3

WORKDIR /app

COPY . /app/
RUN bundle install --local --deployment

ENTRYPOINT ["bundle", "exec", "/app/github-cat"]
