package redistest

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Handler holds dependencies for the Redis demo handlers
type Handler struct {
	Redis *redis.Client
}

func NewHandler(redis *redis.Client) *Handler {
	return &Handler{
		Redis: redis,
	}
}

// Main demonstrates a Redis WATCH transaction with concurrent modification
func (h *Handler) TxPipelineDemo(gctx *gin.Context) {
	client := h.Redis
	ctx := gctx
	// Response map to store all results
	response := map[string]interface{}{}

	err := client.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		response["error"] = err.Error()
		gctx.JSON(500, response)
		return
	}

	val, err := client.Get(ctx, "foo").Result()
	if err != nil {
		response["error"] = err.Error()
		gctx.JSON(500, response)
		return
	}

	response["foo_initial_value"] = val

	// Create transaction with watch
	key := "a"
	watchResults := map[string]interface{}{}

	// Goroutine to modify the watched key
	go func() {
		time.Sleep(1 * time.Second)
		err = client.Set(ctx, key, "abc", time.Hour).Err()
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			watchResults["goroutine_error"] = err.Error()
		} else {
			watchResults["goroutine_set_value"] = val
		}
	}()

	// Watch transaction
	err = client.Watch(ctx, func(tx *redis.Tx) error {
		// Get the initial value
		value, err := tx.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		watchResults["initial_value"] = value

		// Simulate delay to ensure the goroutine has time to change the key
		time.Sleep(5 * time.Second)

		// Try to execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.LRange(ctx, key, 0, -1)
			pipe.Set(ctx, key, fmt.Sprintf("%s-new", value), time.Hour)
			pipe.Set(ctx, "b", "ghf", time.Hour)
			return nil
		})

		if err != nil {
			watchResults["transaction_error"] = err.Error()
			return err
		}

		return nil
	}, key)

	if err != nil {
		watchResults["watch_error"] = err.Error()
	}

	// Get final values
	aVal, err := client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		watchResults["final_a_error"] = err.Error()
	} else {
		watchResults["final_a_value"] = aVal
	}

	bVal, err := client.Get(ctx, "b").Result()
	if err != nil && err != redis.Nil {
		watchResults["final_b_error"] = err.Error()
	} else {
		watchResults["final_b_value"] = bVal
	}

	response["watch_transaction"] = watchResults

	gctx.JSON(200, response)
}

// SyntaxErrorDemo demonstrates Case 1: A syntax error in a transaction causes all commands to be aborted
func (h *Handler) SyntaxErrorDemo(gctx *gin.Context) {
	client := h.Redis
	ctx := gctx

	// Clear any existing keys
	client.Del(ctx, "key1", "key2", "key3", "key4")

	// Start a Redis transaction
	pipe := client.TxPipeline()

	// Add some valid commands
	pipe.Set(ctx, "key1", "value1", time.Hour)
	pipe.Set(ctx, "key2", "value2", time.Hour)

	// Add a command with syntax error - trying to increment with a string
	// This isn't a true syntax error in go-redis (it gets caught when executing)
	// but this will demonstrate the concept
	incrResult := pipe.Do(ctx, "INCR", "key3", "not-a-number") // INCR only takes one argument

	// Add another valid command
	pipe.Set(ctx, "key4", "value4", time.Hour)

	// Try to execute the transaction
	results, err := pipe.Exec(ctx)

	// Report the results
	response := map[string]interface{}{
		"transaction_error": err != nil,
		"error_message":     fmt.Sprintf("%v", err),
		"results":           fmt.Sprintf("%v", results),
	}

	// Check if any keys were set
	key1, err1 := client.Get(ctx, "key1").Result()
	key2, err2 := client.Get(ctx, "key2").Result()
	key4, err4 := client.Get(ctx, "key4").Result()

	response["key1_value"] = key1
	response["key1_error"] = err1 != nil && err1 != redis.Nil
	response["key2_value"] = key2
	response["key2_error"] = err2 != nil && err2 != redis.Nil
	response["key4_value"] = key4
	response["key4_error"] = err4 != nil && err4 != redis.Nil
	response["incr_result_error"] = incrResult.Err() != nil

	// Summary of what happened
	response["summary"] = "This demonstrates Case 1: If any command has a syntax error, " +
		"the entire transaction is aborted and no commands are executed."

	gctx.JSON(200, response)
}

// LogicErrorDemo demonstrates Case 2: A logical error in a transaction allows other commands to execute
func (h *Handler) LogicErrorDemo(gctx *gin.Context) {
	client := h.Redis
	ctx := gctx

	// Set up initial state - counter is a string, not a number
	client.Set(ctx, "counter", "hello", 0)
	client.Del(ctx, "key1", "key2")

	// Start a transaction
	pipe := client.TxPipeline()

	// Queue first valid command
	pipe.Set(ctx, "key1", "value1", time.Hour)

	// Queue a command that will have a logical error - trying to increment a string
	incrResult := pipe.Incr(ctx, "counter")

	// Queue another valid command
	pipe.Set(ctx, "key2", "value2", time.Hour)

	// Execute the transaction
	results, err := pipe.Exec(ctx)

	// Gather the state after execution
	counter, _ := client.Get(ctx, "counter").Result()
	key1, _ := client.Get(ctx, "key1").Result()
	key2, _ := client.Get(ctx, "key2").Result()

	// Build response
	response := map[string]interface{}{
		"transaction_error": err != nil,
		"overall_error":     fmt.Sprintf("%v", err),
		"results":           fmt.Sprintf("%v", results),
		"counter_value":     counter,
		"key1_value":        key1,
		"key2_value":        key2,
		"incr_result_error": incrResult.Err() != nil,
	}

	// Add a summary of what happened
	response["summary"] = "This demonstrates Case 2: If a command has a logical error during execution, " +
		"other commands in the transaction still execute successfully."

	gctx.JSON(200, response)
}
