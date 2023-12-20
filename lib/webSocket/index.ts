import {WebSocket, WebSocketServer } from 'ws';
import * as http from 'http';

interface Client {
  ws: WebSocket;
  topic: string;
}

const clients: Client[] = [];

const wss = new WebSocketServer({ port: 80 });

wss.on('connection', (ws: WebSocket) => {
  const client: Client = { ws, topic: '' };
  clients.push(client);

  ws.on('message', (message: string) => {
    console.log('受信メッセージ:', message.toString());

    try {
      const data = JSON.parse(message);

        client.topic = data.topic;

      // 同じトピックに属するクライアントにメッセージを送信
      if (client.topic) {
        clients.forEach((c) => {
          if (c.topic === client.topic && c.ws !== ws && data.message) {
            c.ws.send(data.message);
          }
        });
      }
    } catch (error) {
      console.error('メッセージの解析に失敗:', error);
    }
  });

  ws.on('close', () => {
    // クライアントの接続が閉じられたらリストから削除
    const index = clients.indexOf(client);
    if (index !== -1) {
      clients.splice(index, 1);
    }
  });
});
//HealthCheckでWebSocketはサポートされないので、HTTPサーバーを立てる
const httpServer = http.createServer((req, res) => {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Healthy');
});

httpServer.listen(8000);
console.log('WebSocketサーバーがポート80で稼働中');

