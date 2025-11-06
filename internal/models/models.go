package models

type User struct {
	ID     int
	Name   string
	Age    int
	About  string
	Gender string
}

type SearchResponse struct {
	Users    []User
	NextPage bool
}

const (
	OrderByAsc  = 1
	OrderByAsIs = 0
	OrderByDesc = -1

	ErrorBadOrderField = `OrderField invalid`
)

type SearchErrorResponse struct {
	Error string
}

