package redistest

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"

	"github.com/go-redis/redis/v8"
)

func Main(gctx *gin.Context) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
	})

	ctx := context.Background()

	err := client.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("foo", val)

	//create transaction
	//watch key
	key := "a"
	go func() {
		time.Sleep(1 * time.Second)
		err = client.Set(ctx, key, "abc", time.Hour).Err()
		val := client.Get(ctx, key)
		fmt.Println("Set val in go routine", val)

		if err != nil {
			panic(err)
		}
	}()

	err = client.Watch(ctx, func(tx *redis.Tx) error {
		value := tx.Get(ctx, key)
		time.Sleep(5 * time.Second)
		fmt.Println(value)

		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.LRange(ctx, key, 0, -1)
			pipe.Set(ctx, key, fmt.Sprintf("%s-new", value.Val()), time.Hour)
			pipe.Set(ctx, "b", "ghf", time.Hour)
			return nil
		})
		if err != nil {
			fmt.Println("EEEEr", err.Error())
			return err
		}
		return nil
	}, "c")

	c := client.Get(ctx, key)
	fmt.Println("final", c)
	d := client.Get(ctx, "b")
	fmt.Println(d)
}
