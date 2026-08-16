package handlers

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/application/usecases"
)

type ComandsHandler struct {
	startDayUsecase *usecases.StartDayUsecase
	endDayUsecase   *usecases.EndDayUsecase
}

func NewComandsHandler(
	startUsecase *usecases.StartDayUsecase,
	endUsecase *usecases.EndDayUsecase,
) *ComandsHandler {
	return &ComandsHandler{
		startDayUsecase: startUsecase,
		endDayUsecase:   endUsecase,
	}
}

func (this *ComandsHandler) Handle(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)
out:
	for {
		select {
		case <-ctx.Done():
			break out
		default:
			os.Stdin.SetReadDeadline(time.Now().Add(time.Millisecond * 100))

			text, err := reader.ReadString('\n')
			if err != nil {
				continue
			}

			text = strings.ReplaceAll(text, "\n", "")

			switch text {
			case "endday":
				this.endDay()
			case "startday":
				this.startDay()
			}
		}
	}
}

func (this *ComandsHandler) endDay() {
	this.endDayUsecase.Execute()
}

func (this *ComandsHandler) startDay() {
	this.startDayUsecase.Execute()
}
