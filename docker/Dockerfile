FROM golang:1

WORKDIR /usr/src/app

COPY . .

RUN go build

CMD [ "/usr/src/app/llmwrangler" ]
