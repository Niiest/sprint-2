package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

type userLikeCodec struct{}

func (ul *userLikeCodec) Encode(value interface{}) (data []byte, err error) {
	if _, isUserLike := value.(*UserLike); !isUserLike {
		return nil, fmt.Errorf("тип должен быть *UserLike, получен %T", value)
	}
	return json.Marshal(value)
}

func (ul *userLikeCodec) Decode(data []byte) (any, error) {
	var (
		userLike UserLike
		err      error
	)
	err = json.Unmarshal(data, &userLike)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &userLike, nil
}

type UserPost struct {
	PostLike map[int]bool
}

// userPostCodec позволяет сериализовать и десериализовать UserPost для групповой таблицы.
type userPostCodec struct{}

// Encode переводит UserPost в []byte
func (up *userPostCodec) Encode(value any) ([]byte, error) {
	if _, isUserPost := value.(*UserPost); !isUserPost {
		return nil, fmt.Errorf("тип должен быть *UserPost, получен %T", value)
	}
	return json.Marshal(value)
}

// Decode переводит UserPost из []byte в структуру.
func (up *userPostCodec) Decode(data []byte) (any, error) {
	var (
		userPost UserPost
		err      error
	)
	err = json.Unmarshal(data, &userPost)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &userPost, nil
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

// ProcessCallback function is called for every message received by the
// processor.
type ProcessCallback func(ctx goka.Context, msg interface{})

func process(ctx goka.Context, msg any) {
	var userLike *UserLike
	var ok bool
	var userPost *UserPost

	if userLike, ok = msg.(*UserLike); !ok || userLike == nil {
		return
	}

	if val := ctx.Value(); val != nil {
		userPost = val.(*UserPost)
	} else {
		userPost = &UserPost{PostLike: make(map[int]bool)}
	}

	userPost.PostLike[userLike.PostId] = userLike.Like

	ctx.SetValue(userPost)
	log.Printf("[proc] key: %s,  msg: %v, data in group_table %v \n", ctx.Key(), userLike, userPost)
}

// runProcessor обрабатывает сообщения из топиков кафка
func runProcessor() {
	g := goka.DefineGroup(group,
		goka.Input(topic, new(userLikeCodec), process),
		goka.Persist(new(userPostCodec)),
	)
	p, err := goka.NewProcessor(brokers,
		g,
	)
	if err != nil {
		log.Fatal(err)
	}
	err = p.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func runView() {
	view, err := goka.NewView(brokers,
		goka.GroupTable(group),
		new(userPostCodec),
	)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		value, _ := view.Get(mux.Vars(r)["key"])
		data, _ := json.Marshal(value)
		w.Write(data)
	})
	fmt.Println("View opened at http://localhost:9095/")
	go http.ListenAndServe(":9095", router)

	view.Run(context.Background())
}

func main() {
	go runEmitter()
	go runProcessor()
	runView()
}
