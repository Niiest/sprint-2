package emitter

import (
	"example/kafka/user"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lovoo/goka"
)

func RunEmitter(brokers []string, outputTopic goka.Stream) {
	// используется user.LikeCodec так как отправляем структуру Like
	emitter, err := goka.NewEmitter(brokers, outputTopic, new(user.LikeCodec))
	if err != nil {
		log.Fatal(err)
	}
	defer emitter.Finish()

	t := time.NewTicker(3 * time.Second)
	defer t.Stop()

	for range t.C {
		userId := rand.Intn(3)

		fakeUserLike := &user.Like{
			Like:   rand.Intn(2) == 1, // Случайное значение для лайка (true или false)
			UserId: userId,            // Случайный ID пользователя
			PostId: rand.Intn(5),      // Случайный ID статьи
		}

		err = emitter.EmitSync(fmt.Sprintf("user-%d", userId), fakeUserLike)
		if err != nil {
			log.Fatal(err)
		}
	}
}
