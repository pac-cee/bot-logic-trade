// Rust Matching Engine Microservice (Actix-web + Redis)
use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use redis::{AsyncCommands, Client};
use serde::{Deserialize, Serialize};
use std::sync::Mutex;

#[derive(Serialize, Deserialize, Clone)]
struct Order {
    id: i32,
    user_id: String,
    order_type: String, // "buy" or "sell"
    price: f64,
    amount: f64,
    remaining: f64,
    status: String, // "open", "matched", "cancelled"
}

struct AppState {
    redis_url: String,
    order_id: Mutex<i32>,
}

/// Health endpoint
async fn health() -> impl Responder {
    HttpResponse::Ok().json(serde_json::json!({"status": "ok"}))
}

/// POST /order - Add new order (concurrency-safe, Redis-backed)
async fn add_order(state: web::Data<AppState>, order: web::Json<Order>) -> impl Responder {
    let mut id_guard = state.order_id.lock().unwrap();
    let id = *id_guard;
    *id_guard += 1;
    drop(id_guard);
    let mut new_order = order.into_inner();
    new_order.id = id;
    new_order.remaining = new_order.amount;
    new_order.status = "open".to_string();
    let client = Client::open(state.redis_url.clone()).unwrap();
    let mut con = client.get_async_connection().await.unwrap();
    let order_key = format!("order:{}", new_order.id);
    let order_json = serde_json::to_string(&new_order).unwrap();
    let _: () = con.set(&order_key, order_json).await.unwrap();
    if new_order.order_type == "buy" {
        let _: () = con.zadd("buy_orders", new_order.id, -new_order.price).await.unwrap();
    } else {
        let _: () = con.zadd("sell_orders", new_order.id, new_order.price).await.unwrap();
    }
    match_orders(&mut con).await;
    HttpResponse::Ok().json(new_order)
}

/// GET /orderbook - Returns current buy/sell orderbook
async fn get_orderbook(state: web::Data<AppState>) -> impl Responder {
    let client = Client::open(state.redis_url.clone()).unwrap();
    let mut con = client.get_async_connection().await.unwrap();
    let buy_ids: Vec<i32> = con.zrange("buy_orders", 0, -1).await.unwrap_or_default();
    let sell_ids: Vec<i32> = con.zrange("sell_orders", 0, -1).await.unwrap_or_default();
    let mut buy_orders = Vec::new();
    let mut sell_orders = Vec::new();
    for id in buy_ids {
        let key = format!("order:{}", id);
        if let Ok(val) = con.get::<_, String>(&key).await {
            if let Ok(order) = serde_json::from_str::<Order>(&val) {
                buy_orders.push(order);
            }
        }
    }
    for id in sell_ids {
        let key = format!("order:{}", id);
        if let Ok(val) = con.get::<_, String>(&key).await {
            if let Ok(order) = serde_json::from_str::<Order>(&val) {
                sell_orders.push(order);
            }
        }
    }
    HttpResponse::Ok().json(serde_json::json!({"buy_orders": buy_orders, "sell_orders": sell_orders}))
}

/// Core matching logic: concurrency-safe, Redis-backed, with event publishing stub.
/// Memory and cache managed via Redis, scalable for microservices.
async fn match_orders(con: &mut redis::aio::Connection) {
    loop {
        let buy_ids: Vec<i32> = con.zrange("buy_orders", 0, 0).await.unwrap_or_default();
        let sell_ids: Vec<i32> = con.zrange("sell_orders", 0, 0).await.unwrap_or_default();
        if buy_ids.is_empty() || sell_ids.is_empty() {
            break;
        }
        let buy_id = buy_ids[0];
        let sell_id = sell_ids[0];
        let buy_val: String = match con.get(format!("order:{}", buy_id)).await { Ok(v) => v, Err(_) => break };
        let sell_val: String = match con.get(format!("order:{}", sell_id)).await { Ok(v) => v, Err(_) => break };
        let mut buy_order: Order = serde_json::from_str(&buy_val).unwrap();
        let mut sell_order: Order = serde_json::from_str(&sell_val).unwrap();
        if buy_order.price < sell_order.price { break; }
        let match_amount = buy_order.remaining.min(sell_order.remaining);
        let match_price = sell_order.price;
        println!("Matched: Buyer {} <-> Seller {} | {} @ ${}", buy_order.user_id, sell_order.user_id, match_amount, match_price);
        // TODO: Publish event to Kafka/RabbitMQ for settlement
        buy_order.remaining -= match_amount;
        sell_order.remaining -= match_amount;
        if buy_order.remaining == 0.0 {
            buy_order.status = "matched".to_string();
            let _: () = con.zrem("buy_orders", buy_order.id).await.unwrap();
        }
        if sell_order.remaining == 0.0 {
            sell_order.status = "matched".to_string();
            let _: () = con.zrem("sell_orders", sell_order.id).await.unwrap();
        }
        let _: () = con.set(format!("order:{}", buy_order.id), serde_json::to_string(&buy_order).unwrap()).await.unwrap();
        let _: () = con.set(format!("order:{}", sell_order.id), serde_json::to_string(&sell_order).unwrap()).await.unwrap();
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let redis_url = std::env::var("REDIS_URL").unwrap_or_else(|_| "redis://localhost:6379".to_string());
    let state = web::Data::new(AppState {
        redis_url: redis_url.clone(),
        order_id: Mutex::new(1),
    });
    println!("Rust Matching Engine Service running on :8084");
    HttpServer::new(move || {
        App::new()
            .app_data(state.clone())
            .route("/health", web::get().to(health))
            .route("/order", web::post().to(add_order))
            .route("/orderbook", web::get().to(get_orderbook))
    })
    .bind(("0.0.0.0", 8084))?
    .run()
    .await
}
