# 基本イメージの選択
FROM node:latest

WORKDIR /usr/src/app

COPY package*.json ./
RUN npm install
COPY . .

EXPOSE 80 8000

CMD ["node", "index.js"]
