package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type Order struct {
	ID       int     `json:"id"`
	UserID   string  `json:"user_id"`
	Type     string  `json:"type"` // "buy" or "sell"
	Price    float64 `json:"price"`
	Amount   float64 `json:"amount"`
	Remaining float64 `json:"remaining"`
	Status   string  `json:"status"`
}

var (
	orderID     = 1
	redisClient *redis.Client
	ctx         = context.Background()
	orderIDMutex sync.Mutex // concurrency-safe order ID
)

func loadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func setupRedis() {
	redisURL := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}
	redisClient = redis.NewClient(opt)
}

func main() {
	loadEnv()
	setupRedis()
	r := gin.Default()

	r.POST("/order", addOrderHandler)
	r.GET("/orderbook", getOrderBookHandler)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	log.Println("Matching Engine Service (Redis) running on :8081")
	r.Run(":8081")
}

func addOrderHandler(c *gin.Context) {
	var req Order
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	orderIDMutex.Lock()
	order := Order{
		ID:       orderID,
		UserID:   req.UserID,
		Type:     req.Type,
		Price:    req.Price,
		Amount:   req.Amount,
		Remaining: req.Amount,
		Status:   "open",
	}
	orderID++
	orderIDMutex.Unlock()
	orderKey := fmt.Sprintf("order:%d", order.ID)
	orderJson, _ := json.Marshal(order)
	redisClient.Set(ctx, orderKey, orderJson, 0)
	if order.Type == "buy" {
		redisClient.ZAdd(ctx, "buy_orders", &redis.Z{Score: -order.Price, Member: order.ID})
	} else {
		redisClient.ZAdd(ctx, "sell_orders", &redis.Z{Score: order.Price, Member: order.ID})
	}
	matchOrders()
	c.JSON(http.StatusOK, order)
}

func getOrderBookHandler(c *gin.Context) {
	buyOrderIDs, _ := redisClient.ZRangeWithScores(ctx, "buy_orders", 0, -1).Result()
	sellOrderIDs, _ := redisClient.ZRangeWithScores(ctx, "sell_orders", 0, -1).Result()
	buyOrders := []Order{}
	sellOrders := []Order{}
	for _, z := range buyOrderIDs {
		id, _ := strconv.Atoi(fmt.Sprintf("%v", z.Member))
		order, err := getOrderByID(id)
		if err == nil {
			buyOrders = append(buyOrders, order)
		}
	}
	for _, z := range sellOrderIDs {
		id, _ := strconv.Atoi(fmt.Sprintf("%v", z.Member))
		order, err := getOrderByID(id)
		if err == nil {
			sellOrders = append(sellOrders, order)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"buy_orders":  buyOrders,
		"sell_orders": sellOrders,
	})
}

func getOrderByID(id int) (Order, error) {
	orderKey := fmt.Sprintf("order:%d", id)
	val, err := redisClient.Get(ctx, orderKey).Result()
	if err != nil {
		return Order{}, err
	}
	var order Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return Order{}, err
	}
	return order, nil
}

// matchOrders performs atomic order matching using Redis, concurrency-safe.
// This is the core of the microservice logic, with event publishing stub.
func matchOrders() {
	for {
		buyIDs, _ := redisClient.ZRangeWithScores(ctx, "buy_orders", 0, 0).Result()
		sellIDs, _ := redisClient.ZRangeWithScores(ctx, "sell_orders", 0, 0).Result()
		if len(buyIDs) == 0 || len(sellIDs) == 0 {
			break
		}
		buyID, _ := strconv.Atoi(fmt.Sprintf("%v", buyIDs[0].Member))
		sellID, _ := strconv.Atoi(fmt.Sprintf("%v", sellIDs[0].Member))
		buyOrder, err1 := getOrderByID(buyID)
		sellOrder, err2 := getOrderByID(sellID)
		if err1 != nil || err2 != nil || buyOrder.Price < sellOrder.Price {
			break
		}
		matchAmount := min(buyOrder.Remaining, sellOrder.Remaining)
		matchPrice := sellOrder.Price
		fmt.Printf("Matched: Buyer %s <-> Seller %s | %f @ $%f\n", buyOrder.UserID, sellOrder.UserID, matchAmount, matchPrice)
		// TODO: Publish event to Kafka/RabbitMQ for settlement
		buyOrder.Remaining -= matchAmount
		sellOrder.Remaining -= matchAmount
		if buyOrder.Remaining == 0 {
			buyOrder.Status = "matched"
			redisClient.ZRem(ctx, "buy_orders", buyOrder.ID)
		}
		if sellOrder.Remaining == 0 {
			sellOrder.Status = "matched"
			redisClient.ZRem(ctx, "sell_orders", sellOrder.ID)
		}
		// Update orders in Redis
		orderJson, _ := json.Marshal(buyOrder)
		redisClient.Set(ctx, fmt.Sprintf("order:%d", buyOrder.ID), orderJson, 0)
		orderJson, _ = json.Marshal(sellOrder)
		redisClient.Set(ctx, fmt.Sprintf("order:%d", sellOrder.ID), orderJson, 0)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
