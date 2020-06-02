FROM ruby:2.7.1

ENV RAILS_ENV=development

RUN gem install bundler

WORKDIR /app
EXPOSE 3000

COPY Gemfile .
COPY Gemfile.lock .

RUN bundle install

COPY . .

CMD ["bundle", "exec", "rails", "server", "-b", "0.0.0.0"]
