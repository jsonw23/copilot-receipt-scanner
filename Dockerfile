FROM golang:1.18 AS go-builder

WORKDIR /usr/src/app

# pre-fetch the dependencies for the go app in a separate layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# copy in the go codebase and build
COPY . .
RUN go build -v -o /usr/local/bin/app ./


FROM node AS node-builder

WORKDIR /usr/src/app

# pre-fetch the dependencies for the web app in a separate layer
COPY ./web/package.json ./web/package-lock.json ./
RUN npm i

# copy the rest of the web app
COPY ./web .
RUN npm run build


FROM golang:1.18 AS prod

ENV STAGE=prod

# copy the compiled binary from go-builder
COPY --from=go-builder /usr/local/bin/app /usr/local/bin/app

# copy the React app build from node-builder
COPY --from=node-builder /usr/src/app/build /var/www/html

# open the default port and run the executable
EXPOSE 8080
CMD ["app"]