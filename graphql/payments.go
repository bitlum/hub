package graphql

type Payment struct {
	FromNode string
	ToNode   string
	Amount   int64
	Status   string
	Time     int64
	Type     string
}
