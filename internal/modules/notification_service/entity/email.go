package entity

// Email структура для отправки письма через адаптер.
type Email struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Text    string
}

// SendResult содержит данные о результате отправки.
type SendResult struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}
