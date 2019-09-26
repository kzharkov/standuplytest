FROM nginx:alpine

RUN groupadd --gid 2001 nginx \
  && useradd --uid 2001 --gid nginx --shell /bin/bash --create-home nginx

USER 2001

COPY nginx.conf /etc/nginx/sites-available/default

EXPOSE 2083