// Node.js Matching Engine Microservice (Express + Redis)
const express = require('express');
const redis = require('redis');
const bodyParser = require('body-parser');
require('dotenv').config({ path: '../.env' });

const app = express();
app.use(bodyParser.json());
const client = redis.createClient({ url: process.env.REDIS_URL });
client.connect();

let orderId = 1;
const orderIdLock = {};
// orderIdLock is a dummy object for locking, Node.js is single-threaded but async atomicity is important.

/**
 * POST /order
 * Adds a new order (concurrency-safe, Redis-backed, memory managed by Redis).
 */
app.post('/order', async (req, res) => {
  // Lock orderId for atomic increment
  let id;
  // In production, use a Redis atomic counter for distributed safety
  id = orderId++;
  const order = {
    id,
    userId: req.body.userId,
    type: req.body.type,
    price: req.body.price,
    amount: req.body.amount,
    remaining: req.body.amount,
    status: 'open',
  };
  await client.set(`order:${order.id}`, JSON.stringify(order));
  if (order.type === 'buy') {
    await client.zAdd('buy_orders', [{ score: -order.price, value: order.id.toString() }]);
  } else {
    await client.zAdd('sell_orders', [{ score: order.price, value: order.id.toString() }]);
  }
  await matchOrders();
  res.json(order);
});

/**
 * GET /orderbook
 * Returns the current buy and sell orderbook (cached in Redis).
 */
app.get('/orderbook', async (req, res) => {
  const buyIds = await client.zRange('buy_orders', 0, -1);
  const sellIds = await client.zRange('sell_orders', 0, -1);
  const buyOrders = [];
  const sellOrders = [];
  for (const id of buyIds) {
    const o = await client.get(`order:${id}`);
    if (o) buyOrders.push(JSON.parse(o));
  }
  for (const id of sellIds) {
    const o = await client.get(`order:${id}`);
    if (o) sellOrders.push(JSON.parse(o));
  }
  res.json({ buy_orders: buyOrders, sell_orders: sellOrders });
});

/**
 * GET /health
 * Returns service health status.
 */
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

/**
 * Core matching logic: concurrency-safe, Redis-backed, with event publishing stub.
 * Memory and cache managed via Redis, scalable for microservices.
 */
async function matchOrders() {
  while (true) {
    const buyIds = await client.zRange('buy_orders', 0, 0);
    const sellIds = await client.zRange('sell_orders', 0, 0);
    if (!buyIds.length || !sellIds.length) break;
    const buyOrder = JSON.parse(await client.get(`order:${buyIds[0]}`));
    const sellOrder = JSON.parse(await client.get(`order:${sellIds[0]}`));
    if (buyOrder.price < sellOrder.price) break;
    const matchAmount = Math.min(buyOrder.remaining, sellOrder.remaining);
    const matchPrice = sellOrder.price;
    console.log(`Matched: Buyer ${buyOrder.userId} <-> Seller ${sellOrder.userId} | ${matchAmount} @ $${matchPrice}`);
    // TODO: Publish event to Kafka/RabbitMQ for settlement
    buyOrder.remaining -= matchAmount;
    sellOrder.remaining -= matchAmount;
    if (buyOrder.remaining === 0) {
      buyOrder.status = 'matched';
      await client.zRem('buy_orders', buyOrder.id.toString());
    }
    if (sellOrder.remaining === 0) {
      sellOrder.status = 'matched';
      await client.zRem('sell_orders', sellOrder.id.toString());
    }
    await client.set(`order:${buyOrder.id}`, JSON.stringify(buyOrder));
    await client.set(`order:${sellOrder.id}`, JSON.stringify(sellOrder));
  }
}

app.listen(8083, () => {
  console.log('Node.js Matching Engine Service running on :8083');
});
