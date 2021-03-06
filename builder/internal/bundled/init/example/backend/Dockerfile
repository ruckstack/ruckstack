FROM node:15

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm ci --only=production

Copy . .

EXPOSE 8080

CMD [ "node", "server.js" ]