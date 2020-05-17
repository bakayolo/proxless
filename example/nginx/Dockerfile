FROM nginx:1.15.11-alpine

WORKDIR /proxless

RUN apk update && apk add --no-cache apache2-utils gettext

ENV BACKEND_HOST="http://hello-proxless"
ENV PORT=80

COPY nginx.conf .

ENTRYPOINT [ "sh", "-c" ]

CMD [ "envsubst < /proxless/nginx.conf > /etc/nginx/conf.d/default.conf && nginx -g \"daemon off;\"" ]