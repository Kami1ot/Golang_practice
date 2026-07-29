package main

import "context"

// Status мгновенно сообщает состояние контекста:
// "отменено" или "работаем" — без блокировки.
func Status(ctx context.Context) string {
	// TODO: select по ctx.Done() с веткой default
	return ""
}
