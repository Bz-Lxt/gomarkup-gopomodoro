package eventbus

import "gopomodoro/internal/model"

type Handler func(ev model.DomainEvent) error
