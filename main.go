package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lovoo/goka"
)

// UserLike — объект, который отправляет пользователь в топик
type UserLike struct {
	Like   bool //лайк поставленный пользователем
	UserId int  // id пользователя
	PostId int  // id статьи которой был поставлен лайк
}

// Codec decodes and encodes from and to []byte
type Codec interface {
	Encode(value interface{}) (data []byte, err error)
	Decode(data []byte) (value interface{}, err error)
}

func Encode(value interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decode(data []byte) (value interface{}, err error) {
	var v interface{}
	buf := new(bytes.Buffer)
	buf.Write(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
}

func runEmitter() {
	// используется userLikeCodec так как отправляем структуру  UserLike
	emitter, err := goka.NewEmitter(brokers, topic, new(userLikeCodec))
	if err != nil {
		log.Fatal(err)
	}
	defer emitter.Finish()

	t := time.NewTicker(100 * time.Second)
	defer t.Stop()

	for range t.C {
		userId := rand.Intn(3)

		fakeUserLike := &UserLike{
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
