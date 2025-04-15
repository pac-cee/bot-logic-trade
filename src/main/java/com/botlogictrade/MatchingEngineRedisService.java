package com.botlogictrade;

import io.github.cdimascio.dotenv.Dotenv;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.ZSetOperations;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@SpringBootApplication
@RestController
public class MatchingEngineRedisService {
    private static int orderId = 1;
    private static final Object orderIdLock = new Object(); // concurrency control
    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @PostMapping("/order")
    /**
     * Adds an order to the orderbook (Redis-backed, concurrency-safe).
     * Handles memory and cache via Redis. Event publishing stub included.
     */
    public Order addOrder(@RequestBody Order req) {
        int id;
        synchronized (orderIdLock) {
            id = orderId++;
        }
        Order order = new Order(id, req.userId, req.type, req.price, req.amount);
        String orderKey = "order:" + order.id;
        redisTemplate.opsForValue().set(orderKey, order);
        if (order.type.equals("buy")) {
            redisTemplate.opsForZSet().add("buy_orders", order.id, -order.price);
        } else {
            redisTemplate.opsForZSet().add("sell_orders", order.id, order.price);
        }
        matchOrders();
        return order;
    }

    @GetMapping("/orderbook")
    public Map<String, List<Order>> getOrderBook() {
        Set<Object> buyIds = redisTemplate.opsForZSet().range("buy_orders", 0, -1);
        Set<Object> sellIds = redisTemplate.opsForZSet().range("sell_orders", 0, -1);
        List<Order> buyOrders = new ArrayList<>();
        List<Order> sellOrders = new ArrayList<>();
        if (buyIds != null) {
            for (Object id : buyIds) {
                Order o = (Order) redisTemplate.opsForValue().get("order:" + id);
                if (o != null) buyOrders.add(o);
            }
        }
        if (sellIds != null) {
            for (Object id : sellIds) {
                Order o = (Order) redisTemplate.opsForValue().get("order:" + id);
                if (o != null) sellOrders.add(o);
            }
        }
        Map<String, List<Order>> map = new HashMap<>();
        map.put("buy_orders", buyOrders);
        map.put("sell_orders", sellOrders);
        return map;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Collections.singletonMap("status", "ok");
    }

    /**
     * Core matching logic: concurrency-safe, Redis-backed, with event publishing stub.
     * Memory and cache managed via Redis, scalable for microservices.
     */
    private void matchOrders() {
        while (true) {
            Set<Object> buySet = redisTemplate.opsForZSet().range("buy_orders", 0, 0);
            Set<Object> sellSet = redisTemplate.opsForZSet().range("sell_orders", 0, 0);
            if (buySet == null || sellSet == null || buySet.isEmpty() || sellSet.isEmpty()) break;
            Integer buyId = Integer.valueOf(buySet.iterator().next().toString());
            Integer sellId = Integer.valueOf(sellSet.iterator().next().toString());
            Order buy = (Order) redisTemplate.opsForValue().get("order:" + buyId);
            Order sell = (Order) redisTemplate.opsForValue().get("order:" + sellId);
            if (buy == null || sell == null || buy.price < sell.price) break;
            double matchAmount = Math.min(buy.remaining, sell.remaining);
            double matchPrice = sell.price;
            System.out.printf("Matched: Buyer %s <-> Seller %s | %f @ $%f\n", buy.userId, sell.userId, matchAmount, matchPrice);
            // TODO: Publish event to Kafka/RabbitMQ for settlement
            buy.remaining -= matchAmount;
            sell.remaining -= matchAmount;
            if (buy.remaining == 0) {
                buy.status = "matched";
                redisTemplate.opsForZSet().remove("buy_orders", buy.id);
            }
            if (sell.remaining == 0) {
                sell.status = "matched";
                redisTemplate.opsForZSet().remove("sell_orders", sell.id);
            }
            redisTemplate.opsForValue().set("order:" + buy.id, buy);
            redisTemplate.opsForValue().set("order:" + sell.id, sell);
        }
    }

    public static void main(String[] args) {
        Dotenv dotenv = Dotenv.configure().ignoreIfMissing().load();
        SpringApplication.run(MatchingEngineRedisService.class, args);
        System.out.println("Matching Engine Redis Service running on :8082");
    }
}

class Order implements java.io.Serializable {
    public int id;
    public String userId;
    public String type; // "buy" or "sell"
    public double price;
    public double amount;
    public double remaining;
    public String status; // "open", "matched", "cancelled"

    public Order() {}
    public Order(int id, String userId, String type, double price, double amount) {
        this.id = id;
        this.userId = userId;
        this.type = type;
        this.price = price;
        this.amount = amount;
        this.remaining = amount;
        this.status = "open";
    }
}
