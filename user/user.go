package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lovoo/goka"
)

var (
	Group goka.Group = "user-like"
)

// Like — отправляется если статья понравилась или не понравилась
type Like struct {
	Like   bool //лайк поставленный пользователем
	UserId int  // id пользователя
	PostId int  // id статьи которой был поставлен лайк
}

// LikeCodec позволяет сериализовать и десериализовать Like.
type LikeCodec struct{}

// Encode переводит Like в []byte
func (uc *LikeCodec) Encode(value any) ([]byte, error) {
	if _, isUser := value.(*Like); !isUser {
		return nil, fmt.Errorf("тип должен быть *Like, получен %T", value)
	}
	return json.Marshal(value)
}

// Decode переводит user из []byte в структуру user.
func (uc *LikeCodec) Decode(data []byte) (any, error) {
	var (
		c   Like
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &c, nil
}

type Post struct {
	PostLike map[int]bool
}

// PostCodec позволяет сериализовать и десериализовать пользователя в/из групповой таблицы.
type PostCodec struct{}

// Encode переводит Post в []byte
func (up *PostCodec) Encode(value any) ([]byte, error) {
	if _, isUserPost := value.(*Post); !isUserPost {
		return nil, fmt.Errorf("тип должен быть *Post, получен %T", value)
	}
	return json.Marshal(value)
}

// Decode переводит Post из []byte в структуру user.
func (up *PostCodec) Decode(data []byte) (any, error) {
	var (
		p   Post
		err error
	)
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &p, nil
}

func process(ctx goka.Context, msg any) {
	var userLike *Like
	var ok bool
	var userPost *Post

	if userLike, ok = msg.(*Like); !ok || userLike == nil {
		return
	}

	if val := ctx.Value(); val != nil {
		userPost = val.(*Post)
	} else {
		userPost = &Post{PostLike: make(map[int]bool)}
	}

	userPost.PostLike[userLike.PostId] = userLike.Like

	ctx.SetValue(userPost)
	log.Printf("[proc] key: %s,  msg: %v, data in group_table %v \n", ctx.Key(), userLike, userPost)
}

func RunUserProcessor(brokers []string, inputStream goka.Stream) {
	g := goka.DefineGroup(Group,
		goka.Input(inputStream, new(LikeCodec), process),
		goka.Persist(new(PostCodec)),
	)
	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		log.Fatal(err)
	}
	err = p.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func RunUserView(brokers []string) {
	view, err := goka.NewView(brokers,
		goka.GroupTable(Group),
		new(PostCodec),
	)
	if err != nil {
		log.Fatal(err)
	}

	root := mux.NewRouter()
	root.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		value, _ := view.Get(mux.Vars(r)["key"])
		data, _ := json.Marshal(value)
		w.Write(data)
	})
	log.Println("View opened at http://localhost:9097/")
	go func() {
		err = http.ListenAndServe(":9097", root)
		if err != nil {
			log.Fatal(err)
		}
	}()
	err = view.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
