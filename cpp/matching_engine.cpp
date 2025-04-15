// C++ Matching Engine Microservice (Crow + Redis)
#include "crow_all.h"
#include <sw/redis++/redis++.h>
#include <nlohmann/json.hpp>
#include <iostream>

struct Order {
    int id;
    std::string userId;
    std::string type;
    double price;
    double amount;
    double remaining;
    std::string status;
};

#include <mutex>
#include <sstream>
#include <vector>
#include <optional>

std::mutex order_id_mutex;
int order_id = 1; // Concurrency-safe order ID

// Serialize Order to JSON
nlohmann::json order_to_json(const Order& o) {
    return nlohmann::json{{"id", o.id}, {"userId", o.userId}, {"type", o.type}, {"price", o.price}, {"amount", o.amount}, {"remaining", o.remaining}, {"status", o.status}};
}

Order json_to_order(const nlohmann::json& j) {
    return Order{j["id"], j["userId"], j["type"], j["price"], j["amount"], j["remaining"], j["status"]};
}

void match_orders(sw::redis::Redis& redis) {
    while (true) {
        std::vector<std::string> buy_ids, sell_ids;
        redis.zrange("buy_orders", 0, 0, std::back_inserter(buy_ids));
        redis.zrange("sell_orders", 0, 0, std::back_inserter(sell_ids));
        if (buy_ids.empty() || sell_ids.empty()) break;
        auto buy_val = redis.get("order:" + buy_ids[0]);
        auto sell_val = redis.get("order:" + sell_ids[0]);
        if (!buy_val || !sell_val) break;
        Order buy = json_to_order(nlohmann::json::parse(*buy_val));
        Order sell = json_to_order(nlohmann::json::parse(*sell_val));
        if (buy.price < sell.price) break;
        double match_amount = std::min(buy.remaining, sell.remaining);
        double match_price = sell.price;
        std::cout << "Matched: Buyer " << buy.userId << " <-> Seller " << sell.userId << " | " << match_amount << " @ $" << match_price << std::endl;
        // TODO: Publish event to Kafka/RabbitMQ for settlement
        buy.remaining -= match_amount;
        sell.remaining -= match_amount;
        if (buy.remaining == 0) {
            buy.status = "matched";
            redis.zrem("buy_orders", buy.id);
        }
        if (sell.remaining == 0) {
            sell.status = "matched";
            redis.zrem("sell_orders", sell.id);
        }
        redis.set("order:" + std::to_string(buy.id), order_to_json(buy).dump());
        redis.set("order:" + std::to_string(sell.id), order_to_json(sell).dump());
    }
}

int main() {
    crow::SimpleApp app;
    sw::redis::Redis redis("tcp://127.0.0.1:6379");
    std::cout << "C++ Matching Engine Service running on :8085" << std::endl;

    // POST /order: Add new order (concurrency-safe, Redis-backed)
    CROW_ROUTE(app, "/order").methods("POST"_method)([&redis](const crow::request& req){
        auto body = nlohmann::json::parse(req.body);
        int id;
        {
            std::lock_guard<std::mutex> lock(order_id_mutex);
            id = order_id++;
        }
        Order o{id, body["userId"], body["type"], body["price"], body["amount"], body["amount"], "open"};
        std::string key = "order:" + std::to_string(o.id);
        redis.set(key, order_to_json(o).dump());
        if (o.type == "buy") {
            redis.zadd("buy_orders", o.id, -o.price);
        } else {
            redis.zadd("sell_orders", o.id, o.price);
        }
        match_orders(redis);
        return crow::response(200, order_to_json(o).dump());
    });

    // GET /orderbook: Return buy/sell orderbook
    CROW_ROUTE(app, "/orderbook")([&redis](){
        std::vector<std::string> buy_ids, sell_ids;
        redis.zrange("buy_orders", 0, -1, std::back_inserter(buy_ids));
        redis.zrange("sell_orders", 0, -1, std::back_inserter(sell_ids));
        std::vector<nlohmann::json> buy_orders, sell_orders;
        for (auto& id : buy_ids) {
            auto val = redis.get("order:" + id);
            if (val) buy_orders.push_back(nlohmann::json::parse(*val));
        }
        for (auto& id : sell_ids) {
            auto val = redis.get("order:" + id);
            if (val) sell_orders.push_back(nlohmann::json::parse(*val));
        }
        nlohmann::json res = { {"buy_orders", buy_orders}, {"sell_orders", sell_orders} };
        return crow::response(200, res.dump());
    });

    // GET /health: Health check
    CROW_ROUTE(app, "/health")([](){
        return crow::response(200, nlohmann::json{{"status", "ok"}}.dump());
    });

    app.port(8085).multithreaded().run();
}
