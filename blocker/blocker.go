package blocker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lovoo/goka"
)

var (
	Group goka.Group = "blocker"
)

type BlockEvent struct {
	Unblock bool
}

type BlockEventCodec struct{}

func (uc *BlockEventCodec) Encode(value any) ([]byte, error) {
	if _, isBlockEvent := value.(*BlockEvent); !isBlockEvent {
		return nil, fmt.Errorf("тип должен быть *BlockEvent, получен %T", value)
	}
	return json.Marshal(value)
}

func (uc *BlockEventCodec) Decode(data []byte) (any, error) {
	var (
		blockEvent BlockEvent
		err        error
	)
	err = json.Unmarshal(data, &blockEvent)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &blockEvent, nil
}

type BlockValue struct {
	Blocked bool
}

type BlockValueCodec struct{}

func (uc *BlockValueCodec) Encode(value any) ([]byte, error) {
	if _, isBlockValue := value.(*BlockValue); !isBlockValue {
		return nil, fmt.Errorf("тип должен быть *BlockValue, получен %T", value)
	}
	return json.Marshal(value)
}

func (uc *BlockValueCodec) Decode(data []byte) (any, error) {
	var (
		blockValue BlockValue
		err        error
	)
	err = json.Unmarshal(data, &blockValue)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации: %v", err)
	}
	return &blockValue, nil
}

func block(ctx goka.Context, msg interface{}) {
	var blockEvent *BlockEvent
	var ok bool
	var blockValue *BlockValue
	// FIXME: Выполните
	// Получаем значение из контекста
	// Если значение отсутствует, создаем новый BlockValue
	// Если значение существует, приводим его к типу *BlockValue
	if val := ctx.Value(); val != nil {
		blockValue = val.(*BlockValue)
	} else {
		blockValue = &BlockValue{}
	}

	// Приводим сообщение к типу *BlockEvent и проверяем, нужно ли разблокировать
	if blockEvent, ok = msg.(*BlockEvent); !ok || blockEvent == nil {
		return
	}

	// Если нужно Разблокируем
	// Иначе - нет
	blockValue.Blocked = !blockEvent.Unblock
	ctx.SetValue(blockValue)
	log.Printf("[proc] key: %s,  msg: %v, data in group_table %v \n", ctx.Key(), blockEvent, blockValue)
}

func RunBlocker(brokers []string, inputTopic goka.Stream) {
	g := goka.DefineGroup(Group,
		goka.Input(inputTopic, new(BlockEventCodec), block),
		goka.Persist(new(BlockValueCodec)),
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
