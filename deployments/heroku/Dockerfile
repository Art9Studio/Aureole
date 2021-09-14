FROM aureolecloud/aureole:latest

RUN apk add --no-cache bash

RUN mv /aureole /app/

COPY templates app/templates/
COPY config.yaml app/
COPY entrypoint.sh app/

RUN mkdir app/keys

CMD [ ".", "app/entrypoint.sh" ]