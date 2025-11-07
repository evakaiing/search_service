package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"searchservice/internal/handlers"
	"searchservice/internal/models"
	"searchservice/internal/storage"
	client "searchservice/pkg"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	Srv        client.SearchClient
	Req        client.SearchRequest
	Expected   string
	IsError    bool
	StatusCode int
	ErrStr     error
}

func TestFindUsers(t *testing.T) {
	oldPath := storage.DatasetPath
	storage.DatasetPath = "../dataset.xml"
	defer func() { storage.DatasetPath = oldPath }()

	server := httptest.NewServer(http.HandlerFunc(handlers.SearchHandler))
	defer server.Close()

	cases := []TestCase{
		// Unauthorized
		{
			Srv:        client.SearchClient{"", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "Age", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusUnauthorized,
			ErrStr:     fmt.Errorf("bad AccessToken"),
			Expected:   ``,
		},
		// BadRequest - ErrorBadOrderField
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "InvalidField", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusBadRequest,
			ErrStr:     fmt.Errorf("OrderFeld InvalidField invalid"),
			Expected:   ``,
		},
		// Limit < 0
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: -1, Offset: 0, OrderField: "Name", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     fmt.Errorf("limit must be > 0"),
			Expected:   ``,
		},
		// Offset < 0
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 10, Offset: -1, OrderField: "Name", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     fmt.Errorf("offset must be > 0"),
			Expected:   ``,
		},
		// Limit > 25
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 30, Offset: 0, OrderField: "Name", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   ``,
		},
		// NextPage = true
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 25, Offset: 0, OrderField: "Name", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   ``,
		},
	}

	for caseNum, item := range cases {
		resp, err := item.Srv.FindUsers(item.Req)
		if item.StatusCode == http.StatusOK {
			if item.ErrStr != nil && (err == nil || err.Error() != item.ErrStr.Error()) {
				t.Errorf("[%d] expected %v, got %v", caseNum, item.ErrStr, err)
			} else if err == nil && item.Expected != "" {
				marshalJson, err := json.Marshal(resp.Users)
				if err != nil {
					t.Errorf("[%d] unexpected error: %v", caseNum, err)
				} else if string(marshalJson) != item.Expected {
					t.Errorf("[%d] wrong Response:\ngot %v\nexpected %v", caseNum, string(marshalJson), item.Expected)

				}
			}
			continue
		}
		if item.StatusCode == http.StatusOK && err != nil {
			t.Errorf("[%d] unexpected error: %v", caseNum, err)
		} else if item.StatusCode != http.StatusOK && (err == nil || (item.ErrStr != nil && err.Error() != item.ErrStr.Error())) {
			t.Errorf("[%d] expected %v, got %v", caseNum, item.ErrStr, err)
		}
	}
}

func TestFindUsers2(t *testing.T) {
	t.Run("BadRequestUnpackJson", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("invalid json{"))
			if err != nil {
				t.Errorf("[BadRequestUnpackJson] unexpected error: %v", err)
				return
			}
		}))
		defer server.Close()

		searchClient := client.SearchClient{"token1", server.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || !strings.Contains(err.Error(), "cant unpack error json") {
			t.Errorf("expected unpack error, got: %v", err)
		}
	})

	t.Run("BadRequestUnknownError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			resp := models.SearchErrorResponse{Error: "some unknown error"}
			jsonData, err := json.Marshal(resp)
			if err != nil {
				t.Errorf("[BadRequestUnknownError] unexpected error: %v", err)
				return
			}
			w.Write(jsonData)
		}))
		defer server.Close()

		searchClient := client.SearchClient{"token1", server.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || !strings.Contains(err.Error(), "unknown bad request error") {
			t.Errorf("expected unknown bad request error, got: %v", err)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		searchClient := client.SearchClient{"token1", server.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || err.Error() != "SearchServer fatal error" {
			t.Errorf("expected SearchServer fatal error, got: %v", err)
		}
	})

	t.Run("OkWithErrorUnpackJson", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("invalid json{"))
			if err != nil {
				t.Errorf("[OkWithErrorUnpackJson] unexpected error: %v", err)
				return
			}
		}))
		defer server.Close()

		searchClient := client.SearchClient{"token1", server.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || !strings.Contains(err.Error(), "cant unpack result json") {
			t.Errorf("expected unpack result error, got: %v", err)
		}
	})

	t.Run("TimeoutError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer server.Close()

		searchClient := client.SearchClient{"token1", server.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || !strings.Contains(err.Error(), "timeout") {
			t.Errorf("expected timeout error, got: %v", err)
		}
	})

	t.Run("Unknown network error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		url := ts.URL
		ts.Close()

		searchClient := client.SearchClient{"token1", url}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || !strings.Contains(err.Error(), "unknown error") {
			t.Errorf("expected unknown error, got: %v", err)
		}
	})
}

func TestSearchSever(t *testing.T) {
	oldPath := storage.DatasetPath
	storage.DatasetPath = "../dataset.xml"
	defer func() { storage.DatasetPath = oldPath }()

	server := httptest.NewServer(http.HandlerFunc(handlers.SearchHandler))
	defer server.Close()
	cases := []TestCase{
		// OrderField=Age different OrderBy
		{
			Srv: client.SearchClient{
				AccessToken: "token1",
				URL:         server.URL,
			},
			Req:        client.SearchRequest{Limit: 2, Offset: 30, OrderField: "Age", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":31,"Name":"Palmer + Scott","Age":37,"About":"Elit fugiat commodo laborum quis eu consequat. In velit magna sit fugiat non proident ipsum tempor eu. Consectetur exercitation labore eiusmod occaecat adipisicing irure consequat fugiat ullamco aliquip nostrud anim irure enim. Duis do amet cillum eiusmod eu sunt. Minim minim sunt sit sit enim velit sint tempor enim sint aliquip voluptate reprehenderit officia. Voluptate magna sit consequat adipisicing ut eu qui.\n","Gender":"male"},{"ID":26,"Name":"Sims + Cotton","Age":39,"About":"Ex cupidatat est velit consequat ad. Tempor non cillum labore non voluptate. Et proident culpa labore deserunt ut aliquip commodo laborum nostrud. Anim minim occaecat est est minim.\n","Gender":"male"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 30, OrderField: "Age", OrderBy: models.OrderByDesc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":14,"Name":"Nicholson + Newman","Age":23,"About":"Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n","Gender":"male"},{"ID":0,"Name":"Boyd + Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "Age", OrderBy: models.OrderByAsIs},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":0,"Name":"Boyd + Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":1,"Name":"Hilda + Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"}]`,
		},
		// OrderField=Id, different OrderBy
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "Id", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":0,"Name":"Boyd + Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":1,"Name":"Hilda + Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 32, OrderField: "Id", OrderBy: models.OrderByDesc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":2,"Name":"Brooks + Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":1,"Name":"Hilda + Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 32, OrderField: "Id", OrderBy: models.OrderByAsIs},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":32,"Name":"Christy + Knapp","Age":40,"About":"Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n","Gender":"female"},{"ID":33,"Name":"Twila + Snow","Age":36,"About":"Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n","Gender":"female"}]`,
		},
		// OrderField=Name, different OrderBy
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "Name", OrderBy: models.OrderByAsc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":15,"Name":"Allison + Valdez","Age":21,"About":"Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n","Gender":"male"},{"ID":16,"Name":"Annie + Osborn","Age":35,"About":"Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n","Gender":"female"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 32, OrderField: "Name", OrderBy: models.OrderByDesc},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":19,"Name":"Bell + Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":16,"Name":"Annie + Osborn","Age":35,"About":"Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n","Gender":"female"}]`,
		},
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 32, OrderField: "Name", OrderBy: models.OrderByAsIs},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":32,"Name":"Christy + Knapp","Age":40,"About":"Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n","Gender":"female"},{"ID":33,"Name":"Twila + Snow","Age":36,"About":"Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n","Gender":"female"}]`,
		},
		// WithQuery
		{
			Srv:        client.SearchClient{"token1", server.URL},
			Req:        client.SearchRequest{Limit: 2, Offset: 0, OrderField: "Name", OrderBy: models.OrderByAsIs, Query: "Annie + Osborn"},
			StatusCode: http.StatusOK,
			ErrStr:     nil,
			Expected:   `[{"ID":16,"Name":"Annie + Osborn","Age":35,"About":"Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n","Gender":"female"}]`,
		},
	}

	for caseNum, item := range cases {
		resp, err := item.Srv.FindUsers(item.Req)
		if item.StatusCode == http.StatusOK {
			if item.ErrStr != nil && (err == nil || err.Error() != item.ErrStr.Error()) {
				t.Errorf("[%d] expected %v, got %v", caseNum, item.ErrStr, err)
			} else if err == nil && item.Expected != "" {
				marshalJson, err := json.Marshal(resp.Users)
				if err != nil {
					t.Errorf("[%d] unexpected error: %v", caseNum, err)
					return
				}

				if string(marshalJson) != item.Expected {
					t.Errorf("[%d] wrong Response:\ngot %v\nexpected %v", caseNum, string(marshalJson), item.Expected)
				}
			}
			continue
		}
		if item.StatusCode == http.StatusOK && err != nil {
			t.Errorf("[%d] unexpected error: %v", caseNum, err)
		} else if item.StatusCode != http.StatusOK && (err == nil || (item.ErrStr != nil && err.Error() != item.ErrStr.Error())) {
			t.Errorf("[%d] expected %v, got %v", caseNum, item.ErrStr, err)
		}
	}
}

func TestSearchServer2(t *testing.T) {

	t.Run("Error open file", func(t *testing.T) {
		oldPath := storage.DatasetPath
		defer func() { storage.DatasetPath = oldPath }()

		storage.DatasetPath = "notexist_file.xml"

		ts := httptest.NewServer(http.HandlerFunc(handlers.SearchHandler))
		defer ts.Close()

		searchClient := client.SearchClient{"token1", ts.URL}
		_, err := searchClient.FindUsers(client.SearchRequest{Limit: 10, Offset: 0})

		if err == nil || err.Error() != "SearchServer fatal error" {
			t.Errorf("expected SearchServer fatal error, got: %v", err)
		}
	})
}
