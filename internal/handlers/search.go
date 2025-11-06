package handlers

import (
	"fmt"
	"searchservice/internal/models"
	"searchservice/internal/storage"
	"sort"
	"net/http"
	"strconv"
	"encoding/json"
)


func findByQuery(users []storage.UserXML, query string) []storage.UserXML {
	var findedUsers []storage.UserXML
	if query != "" {
		for i := 0; i < len(users); i++ {
			name := fmt.Sprintf("%s + %s", users[i].FirstName, users[i].LastName)
			if name == query || users[i].About == query {
				findedUsers = append(findedUsers, users[i])
			}
		}
		return findedUsers
	}
	return users
}

func sortUsers(users *storage.UsersXML, orderField string, orderBy int) string {
	switch orderField {
	case "Id":
		sort.Slice(users.Users, func(i, j int) bool {
			if orderBy == models.OrderByAsc {
				return users.Users[i].Id < users.Users[j].Id
			} else if orderBy == models.OrderByDesc {
				return users.Users[i].Id > users.Users[j].Id
			}

			return false
		})
	case "Age":
		sort.Slice(users.Users, func(i, j int) bool {
			if orderBy == models.OrderByAsc {
				return users.Users[i].Age < users.Users[j].Age
			} else if orderBy == models.OrderByDesc {
				return users.Users[i].Age > users.Users[j].Age
			}
			return false
		})

	case "Name", "":
		sort.Slice(users.Users, func(i, j int) bool {
			if orderBy == models.OrderByAsc {
				if users.Users[i].FirstName == users.Users[j].FirstName {
					return users.Users[i].LastName < users.Users[j].LastName
				}
				return users.Users[i].FirstName < users.Users[j].FirstName
			} else if orderBy == models.OrderByDesc {
				if users.Users[i].FirstName == users.Users[j].FirstName {
					return users.Users[i].LastName > users.Users[j].LastName
				}
				return users.Users[i].FirstName > users.Users[j].FirstName
			}
			return false
		})
	default:
		err := models.ErrorBadOrderField
		return err
	}
	return ""
}

func getUsers(users storage.UsersXML, limit, offset int, query, orderField string, orderBy int) ([]models.User, string) {
	if query != "" {
		users.Users = findByQuery(users.Users, query)
	}
	err := sortUsers(&users, orderField, orderBy)
	if err != "" {
		return nil, err
	}

	if len(users.Users) < limit+offset {
		users.Users = users.Users[offset:]
	} else {
		users.Users = users.Users[offset : limit+offset]
	}

	var result []models.User
	for _, u := range users.Users {
		result = append(result, models.User{
			ID:     u.Id,
			Name:   fmt.Sprintf("%s + %s", u.FirstName, u.LastName),
			Age:    u.Age,
			About:  u.About,
			Gender: u.Gender,
		})
	}
	return result, ""
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("AccessToken")
	if accessToken == "" {
		http.Error(w, "bad AccessToken", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))

	usersXML, err := storage.LoadUsers(storage.DatasetPath)
	if err != nil {
		http.Error(w, "error open file", http.StatusInternalServerError)
		return
	}

	handledUsers, searchError := getUsers(*usersXML, limit, offset, query, orderField, orderBy)
	if searchError != "" {
		errResp := models.SearchErrorResponse{Error: searchError}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	jsonMarshalData, err := json.Marshal(handledUsers)
	if err != nil {
		http.Error(w, "error marshal json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonMarshalData)
}